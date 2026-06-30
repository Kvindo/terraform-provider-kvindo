// Package main implements a code generator that reads the Kvindo Cloud API swagger.json
// and generates Terraform provider resource and datasource files.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// SwaggerSpec represents the minimal swagger/OpenAPI spec we need.
type SwaggerSpec struct {
	Paths      map[string]PathItem            `json:"paths"`
	Components Components                     `json:"components"`
}

type Components struct {
	Schemas map[string]SchemaObject `json:"schemas"`
}

type PathItem struct {
	Put    *Operation `json:"put"`
	Get    *Operation `json:"get"`
	Delete *Operation `json:"delete"`
}

type Operation struct {
	OperationID string              `json:"operationId"`
	RequestBody *RequestBody        `json:"requestBody"`
	Responses   map[string]Response `json:"responses"`
}

type RequestBody struct {
	Content map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema *SchemaRef `json:"schema"`
}

type Response struct {
	Content map[string]MediaType `json:"content"`
}

type SchemaRef struct {
	Ref        string                 `json:"$ref"`
	Type       string                 `json:"type"`
	Format     string                 `json:"format"`
	Properties map[string]*SchemaRef  `json:"properties"`
	Items      *SchemaRef             `json:"items"`
	AllOf      []*SchemaRef           `json:"allOf"`
}

type SchemaObject struct {
	Type       string                `json:"type"`
	Properties map[string]*SchemaRef `json:"properties"`
	AllOf      []*SchemaRef          `json:"allOf"`
}

// FieldDef describes a single terraform schema field.
type FieldDef struct {
	TFName       string
	APIName      string
	FieldType    string // "string", "bool", "int64", "float64", "list_string", "map_string", "list_object"
	Required     bool
	OptionalOnly bool // Optional but NOT Computed (server never defaults it)
	Computed     bool
	Sensitive    bool
	ObjFields    []FieldDef
}

// ResourceDef describes a complete resource to generate.
type ResourceDef struct {
	Name        string
	APIPath     string
	Fields      []FieldDef // spec fields (type-specific)
	StatusExtra []FieldDef // status fields beyond the base ResourceInfo
}

// ---- Overrides for things the swagger cannot express ----
//
// The swagger marks no fields `required` and only `resourceName`/`id` readOnly, and its status
// field casing does not match the wire. These tables encode the human decisions captured from
// the previous hand-written provider so regeneration reproduces them exactly.

// skipResources are PUT-path resources present in the swagger that the provider does NOT expose
// (removed/unimplemented backends, plus the hand-written transaction resource).
var skipResources = map[string]bool{
	"etcd": true, "etcd_node_group": true, "grafana": true, "nat_gateway": true,
	"ollama": true, "postgresql": true, "postgresql_node_group": true,
	"vm_on_off_maintenance_action": true, "vm_recurrent_command_maintenance_action": true,
	"transaction": true,
}

// resourceNameOverride maps a path-derived name to the provider's resource name.
var resourceNameOverride = map[string]string{
	"route_table_attachments": "route_table_attachment",
}

// requiredSpecFields[resource][tf_field] = true marks a spec field Required.
var requiredSpecFields = map[string]map[string]bool{
	"access_policy":                                      {"content": true},
	"certificate":                                        {"certificate_pem": true, "private_key_pem": true},
	"kubernetes_node_group":                              {"kubernetes_id": true},
	"kubernetes_user":                                    {"kubernetes_id": true},
	"loadbalancer_http_listener":                         {"loadbalancer_id": true},
	"loadbalancer_http_listener_rule":                    {"http_listener_id": true},
	"loadbalancer_https_listener":                        {"loadbalancer_id": true},
	"loadbalancer_https_listener_rule":                   {"https_listener_id": true},
	"loadbalancer_target_group_service_discovery_target": {"target_group_id": true},
	"loadbalancer_target_group_static_target":            {"target_group_id": true, "ip_or_hostname": true},
	"loadbalancer_tcp_listener":                          {"loadbalancer_id": true},
	"loadbalancer_tcp_listener_rule":                     {"tcp_listener_id": true},
	"loadbalancer_tls_listener":                          {"loadbalancer_id": true},
	"loadbalancer_tls_listener_rule":                     {"tls_listener_id": true},
	"loadbalancer_udp_listener":                          {"loadbalancer_id": true},
	"loadbalancer_udp_listener_rule":                     {"udp_listener_id": true},
	"open_vpn_user":                                      {"open_vpn_id": true},
	"quota_change_request":                               {"quota_id": true, "new_quota_limit": true},
	"route_table_attachment":                             {"route_table_id": true, "vpc_id": true},
	"route_table_route":                                  {"route_table_id": true, "destination_cidr": true, "target_ip": true},
	"s3_user":                                            {"bucket_id": true},
	"s3_user_access_policy":                              {"policy_json": true},
	"ssh_key":                                            {"public_key": true},
	"ssh_private_key":                                    {"private_key": true},
	"support_ticket_comment":                             {"ticket_id": true},
	"user":                                               {"email": true},
	"user_token":                                         {"user_id": true},
	"volume_attachment":                                  {"volume_id": true, "vm_id": true},
	"vpc_peering_external_peer":                          {"vpc_peering_id": true},
	"vpc_peering_peer":                                   {"vpc_peering_id": true},
	"vpc_subnet":                                         {"vpc_id": true, "ipv4_cidr": true},
}

// optionalOnlySpecFields[resource][tf_field] = true marks a spec field Optional (NOT Computed):
// the server never defaults it, so making it Computed would produce spurious plan diffs.
var optionalOnlySpecFields = map[string]map[string]bool{
	"gitlab":                    {"floating_ip_id": true, "record_name": true},
	"loadbalancer":              {"floating_ip_id": true},
	"open_vpn":                  {"floating_ip_id": true},
	"postgresql_standalone":     {"parameters_set_id": true, "floating_ip_id": true},
	"victoria_metrics":          {"dns_record_name": true},
	"vm":                        {"floating_ip_id": true, "security_group_ids": true},
	"vpc":                       {"nat_floating_ip_id": true},
	"vpc_peering_external_peer": {"ssh_private_key_id": true},
}

// sensitiveSpecFields[resource][tf_field] = true marks a spec field Sensitive. The swagger
// keyword heuristic over-matches (it would mark public_key etc.), so sensitivity is explicit.
var sensitiveSpecFields = map[string]map[string]bool{
	"certificate":           {"private_key_pem": true},
	"gitlab":                {"root_password": true},
	"postgresql_standalone": {"root_password": true},
	"ssh_private_key":       {"private_key": true},
}

// sensitiveStatusFields[tf_field] = true marks a status field Sensitive (consistent by name
// across resources).
var sensitiveStatusFields = map[string]bool{
	"token": true, "kubeconfig": true, "secret_key": true, "config": true,
	"windows_administrator_password": true,
}

// baseInfoFields are the ResourceInfo fields handled by the common status block; they must NOT
// be emitted as resource-specific status extras.
var baseInfoFields = map[string]bool{
	"state": true, "createTime": true, "createdByUser": true,
	"lastChangeRequest": true, "pricing": true,
}

func main() {
	swaggerPath := flag.String("swagger", "../../kvindo-api.json", "Path to swagger.json")
	outputDir := flag.String("output", "../../internal/provider", "Output directory for generated files")
	flag.Parse()

	data, err := os.ReadFile(*swaggerPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading swagger file: %v\n", err)
		os.Exit(1)
	}

	var spec SwaggerSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing swagger JSON: %v\n", err)
		os.Exit(1)
	}

	resources := extractResources(spec)
	fmt.Printf("Found %d resources\n", len(resources))

	for _, r := range resources {
		resContent := generateResourceFile(r)
		resPath := filepath.Join(*outputDir, "resource_"+r.Name+".go")
		if err := os.WriteFile(resPath, []byte(resContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", resPath, err)
			os.Exit(1)
		}
		fmt.Printf("Generated resource: %s\n", resPath)

		dsContent := generateDatasourceFile(r)
		dsPath := filepath.Join(*outputDir, "datasource_"+r.Name+".go")
		if err := os.WriteFile(dsPath, []byte(dsContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", dsPath, err)
			os.Exit(1)
		}
		fmt.Printf("Generated datasource: %s\n", dsPath)
	}

	// Regenerate provider.go
	providerContent := generateProviderFile(resources)
	providerPath := filepath.Join(*outputDir, "provider.go")
	if err := os.WriteFile(providerPath, []byte(providerContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing provider.go: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated provider: %s\n", providerPath)

	fmt.Println("Generation complete!")
}

// skipPaths are path fragments that should not generate resource files.
var skipPaths = []string{
	"/internal", "/acquiring", "/health-check", "/request/", "/get-by-labels",
}

func shouldSkip(path string) bool {
	for _, skip := range skipPaths {
		if strings.Contains(path, skip) {
			return true
		}
	}
	// Only process paths that match /api/v1/{resource} (no further path segments after resource)
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	if len(parts) != 1 || parts[0] == "" {
		return true
	}
	return false
}

func extractResources(spec SwaggerSpec) []ResourceDef {
	var resources []ResourceDef

	paths := make([]string, 0, len(spec.Paths))
	for p := range spec.Paths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, path := range paths {
		if !strings.HasPrefix(path, "/api/v1/") {
			continue
		}
		if shouldSkip(path) {
			continue
		}

		item := spec.Paths[path]
		if item.Put == nil {
			continue
		}

		resourceSlug := strings.TrimPrefix(path, "/api/v1/")
		resourceName := toSnakeCase(strings.ReplaceAll(resourceSlug, "-", "_"))
		if override, ok := resourceNameOverride[resourceName]; ok {
			resourceName = override
		}
		if skipResources[resourceName] {
			continue
		}

		var fields, statusExtra []FieldDef

		if item.Put.RequestBody != nil {
			if ct, ok := item.Put.RequestBody.Content["application/json"]; ok && ct.Schema != nil {
				schemaName := extractRefName(ct.Schema.Ref)
				if schemaName != "" {
					if schema, ok := spec.Components.Schemas[schemaName]; ok {
						fields, statusExtra = extractFields(schema, spec.Components.Schemas)
					}
				}
			}
		}

		// Apply per-field overrides the swagger cannot express.
		req := requiredSpecFields[resourceName]
		optOnly := optionalOnlySpecFields[resourceName]
		sens := sensitiveSpecFields[resourceName]
		for i := range fields {
			tf := fields[i].TFName
			fields[i].Required = req[tf]
			fields[i].OptionalOnly = optOnly[tf]
			fields[i].Sensitive = sens[tf]
		}
		for i := range statusExtra {
			statusExtra[i].Sensitive = sensitiveStatusFields[statusExtra[i].TFName]
		}

		resources = append(resources, ResourceDef{
			Name:        resourceName,
			APIPath:     path,
			Fields:      fields,
			StatusExtra: statusExtra,
		})
	}

	return resources
}

// commonFields are fields handled by common schema and should be skipped in resource-specific fields.
// resourceName is a legacy readOnly field removed from the provider; it is skipped entirely.
var commonFieldNames = map[string]bool{
	"id": true, "name": true, "description": true, "labels": true,
	"folderId": true, "deleteProtection": true, "resourceName": true,
}

func extractFields(schema SchemaObject, schemas map[string]SchemaObject) (fields, infoFields []FieldDef) {
	props := schema.Properties
	if props == nil && len(schema.AllOf) > 0 {
		props = make(map[string]*SchemaRef)
		for _, s := range schema.AllOf {
			if s.Ref != "" {
				refName := extractRefName(s.Ref)
				if refSchema, ok := schemas[refName]; ok {
					for k, v := range refSchema.Properties {
						props[k] = v
					}
				}
			}
			for k, v := range s.Properties {
				props[k] = v
			}
		}
	}

	for apiName, prop := range props {
		if commonFieldNames[apiName] {
			continue
		}

		tfName := camelToSnake(apiName)
		isInfo := apiName == "info" || strings.HasPrefix(apiName, "info")

		if apiName == "info" && prop.Ref != "" {
			// Extract status (info) sub-fields beyond the base ResourceInfo. The base fields
			// (state/createTime/createdByUser/lastChangeRequest/pricing) are rendered by the
			// common status block, so only resource-specific extras are collected here.
			refName := extractRefName(prop.Ref)
			if infoSchema, ok := schemas[refName]; ok {
				infoProps := infoSchema.Properties
				if infoProps == nil && len(infoSchema.AllOf) > 0 {
					infoProps = make(map[string]*SchemaRef)
					for _, s := range infoSchema.AllOf {
						if s.Ref != "" {
							if rs, ok := schemas[extractRefName(s.Ref)]; ok {
								for k, v := range rs.Properties {
									infoProps[k] = v
								}
							}
						}
						for k, v := range s.Properties {
							infoProps[k] = v
						}
					}
				}
				for infoFieldName, infoFieldProp := range infoProps {
					if baseInfoFields[infoFieldName] {
						continue
					}
					infoFields = append(infoFields, FieldDef{
						TFName:    camelToSnake(infoFieldName),
						APIName:   infoFieldName,
						FieldType: swaggerTypeToFieldType(infoFieldProp),
						Computed:  true,
					})
				}
			}
			continue
		}

		if isInfo {
			continue
		}

		ft := swaggerTypeToFieldType(prop)
		var objFields []FieldDef

		switch ft {
		case "object":
			objFields = extractObjFields(prop, schemas)
			if len(objFields) == 0 {
				ft = "string" // $ref to a scalar/enum
			}
		case "list_object":
			objFields = extractObjFields(prop.Items, schemas)
			if len(objFields) == 0 {
				ft = "list_string" // array of scalars/enums
			}
		}

		fields = append(fields, FieldDef{
			TFName:    tfName,
			APIName:   apiName,
			FieldType: ft,
			ObjFields: objFields,
		})
	}

	sort.Slice(fields, func(i, j int) bool { return fields[i].TFName < fields[j].TFName })
	sort.Slice(infoFields, func(i, j int) bool { return infoFields[i].TFName < infoFields[j].TFName })

	return fields, infoFields
}

func swaggerTypeToFieldType(prop *SchemaRef) string {
	if prop == nil {
		return "string"
	}
	// A $ref points at another schema — a nested object (downgraded to string later if the
	// target turns out to be a scalar/enum with no properties).
	if prop.Ref != "" {
		return "object"
	}
	switch prop.Type {
	case "boolean":
		return "bool"
	case "integer":
		return "int64"
	case "number":
		return "float64"
	case "array":
		if prop.Items != nil {
			if prop.Items.Type == "string" {
				return "list_string"
			}
			// array of objects (or $ref elements)
			return "list_object"
		}
		return "list_string"
	case "object":
		if len(prop.Properties) > 0 {
			return "object"
		}
		return "map_string"
	default:
		return "string"
	}
}

// extractObjFields recursively extracts the sub-fields of a nested object schema (a $ref or an
// inline object). Returns nil for scalar/enum targets (the caller downgrades such fields).
func extractObjFields(ref *SchemaRef, schemas map[string]SchemaObject) []FieldDef {
	if ref == nil {
		return nil
	}
	var props map[string]*SchemaRef
	if ref.Ref != "" {
		if os, ok := schemas[extractRefName(ref.Ref)]; ok {
			props = os.Properties
			if props == nil && len(os.AllOf) > 0 {
				props = map[string]*SchemaRef{}
				for _, s := range os.AllOf {
					if s.Ref != "" {
						if rs, ok := schemas[extractRefName(s.Ref)]; ok {
							for k, v := range rs.Properties {
								props[k] = v
							}
						}
					}
					for k, v := range s.Properties {
						props[k] = v
					}
				}
			}
		}
	} else if len(ref.Properties) > 0 {
		props = ref.Properties
	}
	var fields []FieldDef
	for name, p := range props {
		ft := swaggerTypeToFieldType(p)
		fd := FieldDef{TFName: camelToSnake(name), APIName: name, FieldType: ft, Sensitive: sensitiveStatusFields[camelToSnake(name)]}
		switch ft {
		case "object":
			fd.ObjFields = extractObjFields(p, schemas)
			if len(fd.ObjFields) == 0 {
				fd.FieldType = "string" // $ref to a scalar/enum
			}
		case "list_object":
			fd.ObjFields = extractObjFields(p.Items, schemas)
			if len(fd.ObjFields) == 0 {
				fd.FieldType = "list_string" // array of scalars/enums
			}
		}
		fields = append(fields, fd)
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].TFName < fields[j].TFName })
	return fields
}

func extractRefName(ref string) string {
	if ref == "" {
		return ""
	}
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

// binaryUnitAcronyms normalizes mixed-case binary unit suffixes (e.g. GiB → Gib)
// before camelToSnake so they don't get split into separate snake segments (gi_b).
var binaryUnitAcronyms = strings.NewReplacer(
	"GiB", "Gib",
	"MiB", "Mib",
	"TiB", "Tib",
	"KiB", "Kib",
	"PiB", "Pib",
)

func camelToSnake(s string) string {
	s = binaryUnitAcronyms.Replace(s)
	var result []rune
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 {
			// Check if previous char was lowercase or next is lowercase
			if unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1])) {
				result = append(result, '_')
			}
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func toSnakeCase(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "-", "_")
}

func toTitle(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			runes := []rune(p)
			parts[i] = string(unicode.ToUpper(runes[0])) + string(runes[1:])
		}
	}
	return strings.Join(parts, "")
}

func structName(name string) string {
	return toTitle(name)
}

func typeToGoType(ft string) string {
	switch ft {
	case "bool":
		return "types.Bool"
	case "int64":
		return "types.Int64"
	case "float64":
		return "types.Float64"
	case "list_string", "list_object":
		return "types.List"
	case "map_string":
		return "types.Map"
	case "object":
		return "types.Object"
	default:
		return "types.String"
	}
}

// descVarName is the package-level variable name for a nested field's []objField descriptor.
func descVarName(sn string, f FieldDef) string {
	return lowerFirst(sn) + toTitle(f.TFName) + "ObjFields"
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// descLiteral emits a recursive []objField literal describing a nested object's sub-fields.
func descLiteral(fields []FieldDef) string {
	var b strings.Builder
	b.WriteString("[]objField{")
	for _, f := range fields {
		b.WriteString(fmt.Sprintf("{TF: %q, API: %q, Kind: %q", f.TFName, f.APIName, f.FieldType))
		if f.Sensitive {
			b.WriteString(", Sensitive: true")
		}
		if f.FieldType == "object" || f.FieldType == "list_object" {
			b.WriteString(", Obj: " + descLiteral(f.ObjFields))
		}
		b.WriteString("}, ")
	}
	b.WriteString("}")
	return b.String()
}

// emitObjDescriptors writes the package-level []objField descriptor vars for a resource's nested
// spec fields (object / list_object). The datasource file reuses these same vars.
func emitObjDescriptors(sb *strings.Builder, sn string, r ResourceDef) {
	for _, f := range r.Fields {
		if f.FieldType == "object" || f.FieldType == "list_object" {
			sb.WriteString(fmt.Sprintf("var %s = %s\n\n", descVarName(sn, f), descLiteral(f.ObjFields)))
		}
	}
}

// hasRequiredSpec reports whether any spec field is Required (=> the spec block is Required).
func hasRequiredSpec(r ResourceDef) bool {
	for _, f := range r.Fields {
		if f.Required {
			return true
		}
	}
	return false
}

// emitImports writes the import block for a resource file based on which features are used.
func emitResourceImports(sb *strings.Builder, r ResourceDef) {
	// attr is only needed for the status-extras buildInfoObj literals; nested objects are handled
	// by runtime helpers in nested_objects.go.
	needsAttr := len(r.StatusExtra) > 0
	needsBool, needsInt64, needsFloat64 := false, false, false
	// Plan modifiers are emitted only for top-level Optional+Computed scalar spec fields; nested
	// object/list_object schemas are built at runtime without per-field modifiers.
	for _, f := range r.Fields {
		if f.Required || f.OptionalOnly {
			continue
		}
		switch f.FieldType {
		case "bool":
			needsBool = true
		case "int64":
			needsInt64 = true
		case "float64":
			needsFloat64 = true
		}
	}
	sb.WriteString("package provider\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"fmt\"\n\n")
	if needsAttr {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/attr\"\n")
	}
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema\"\n")
	if needsBool {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier\"\n")
	}
	if needsFloat64 {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier\"\n")
	}
	if needsInt64 {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier\"\n")
	}
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/types\"\n")
	sb.WriteString("\t\"github.com/kvindo/terraform-provider-kvindo/internal/client\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("var _ = fmt.Sprintf\n\n")
}

// emitSpecModel writes the per-resource spec struct (omitted when there are no spec fields).
func emitSpecModel(sb *strings.Builder, sn string, r ResourceDef) {
	if len(r.Fields) == 0 {
		return
	}
	sb.WriteString(fmt.Sprintf("type %sSpecModel struct {\n", sn))
	for _, f := range r.Fields {
		sb.WriteString(fmt.Sprintf("\t%s %s `tfsdk:%q`\n", toTitle(f.TFName), typeToGoType(f.FieldType), f.TFName))
	}
	sb.WriteString("}\n\n")
}

// emitStatusSchemaArg returns the argument expression for commonInfoSchema / commonInfoDatasourceSchema.
func emitStatusSchemaArg(r ResourceDef) string {
	if len(r.StatusExtra) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString("map[string]schema.Attribute{")
	for _, f := range r.StatusExtra {
		var attrType string
		switch f.FieldType {
		case "int64":
			attrType = "schema.Int64Attribute{Computed: true}"
		case "bool":
			attrType = "schema.BoolAttribute{Computed: true}"
		default:
			if f.Sensitive {
				attrType = "schema.StringAttribute{Computed: true, Sensitive: true}"
			} else {
				attrType = "schema.StringAttribute{Computed: true}"
			}
		}
		b.WriteString(fmt.Sprintf("%q: %s, ", f.TFName, attrType))
	}
	b.WriteString("}")
	return b.String()
}

// emitStatusAssign writes the `state.Status = ...` line for populate/read. dataVar is the
// response map variable ("data" in resources, "apiData" in datasources).
func emitStatusAssign(sb *strings.Builder, r ResourceDef, dataVar string) {
	if len(r.StatusExtra) == 0 {
		sb.WriteString(fmt.Sprintf("\tstate.Status = simpleStateInfoObj(%s)\n", dataVar))
		return
	}
	sb.WriteString("\tstate.Status = buildInfoObj(" + dataVar + ",\n")
	sb.WriteString("\t\tmap[string]attr.Type{\n")
	for _, f := range r.StatusExtra {
		var at string
		switch f.FieldType {
		case "int64":
			at = "types.Int64Type"
		case "bool":
			at = "types.BoolType"
		default:
			at = "types.StringType"
		}
		sb.WriteString(fmt.Sprintf("\t\t\t%q: %s,\n", f.TFName, at))
	}
	sb.WriteString("\t\t},\n")
	sb.WriteString("\t\tmap[string]attr.Value{\n")
	for _, f := range r.StatusExtra {
		var getter string
		switch f.FieldType {
		case "int64":
			getter = "getInt64FromInfo"
		case "bool":
			getter = "getBoolFromInfo"
		default:
			getter = "getStringFromInfo"
		}
		sb.WriteString(fmt.Sprintf("\t\t\t%q: %s(%s, %q),\n", f.TFName, getter, dataVar, f.APIName))
	}
	sb.WriteString("\t\t})\n")
}

func generateResourceFile(r ResourceDef) string {
	sn := structName(r.Name)
	hasSpec := len(r.Fields) > 0
	var sb strings.Builder

	emitResourceImports(&sb, r)
	emitObjDescriptors(&sb, sn, r)
	emitSpecModel(&sb, sn, r)

	// Model
	sb.WriteString(fmt.Sprintf("type %sResourceModel struct {\n", sn))
	sb.WriteString("\tID       types.String  `tfsdk:\"id\"`\n")
	sb.WriteString("\tMetadata metadataModel `tfsdk:\"metadata\"`\n")
	if hasSpec {
		sb.WriteString(fmt.Sprintf("\tSpec     %sSpecModel `tfsdk:\"spec\"`\n", sn))
	}
	sb.WriteString("\tStatus   types.Object  `tfsdk:\"status\"`\n")
	sb.WriteString("}\n\n")

	// Resource struct
	sb.WriteString(fmt.Sprintf("type %sResource struct { client *client.Client }\n\n", sn))
	sb.WriteString(fmt.Sprintf("func New%sResource() resource.Resource { return &%sResource{} }\n\n", sn, sn))

	// Metadata
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tresp.TypeName = req.ProviderTypeName + \"_%s\"\n}\n\n", r.Name))

	// Full resource schema attrs (exported so the transaction resource reuses one definition).
	sb.WriteString(fmt.Sprintf("func %sResourceSchemaAttrs() map[string]schema.Attribute {\n", sn))
	if hasSpec {
		sb.WriteString("\tspecAttrs := map[string]schema.Attribute{\n")
		for _, f := range r.Fields {
			var expr string
			switch f.FieldType {
			case "object":
				expr = fmt.Sprintf("objResourceSchema(%s)", descVarName(sn, f))
			case "list_object":
				expr = fmt.Sprintf("listObjResourceSchema(%s)", descVarName(sn, f))
			default:
				expr = resourceAttrDef(f)
			}
			sb.WriteString(fmt.Sprintf("\t\t%q: %s,\n", f.TFName, expr))
		}
		sb.WriteString("\t}\n")
	}
	sb.WriteString("\treturn map[string]schema.Attribute{\n")
	sb.WriteString("\t\t\"id\": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},\n")
	sb.WriteString("\t\t\"metadata\": metadataResourceSchema(),\n")
	if hasSpec {
		if hasRequiredSpec(r) {
			sb.WriteString("\t\t\"spec\": schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},\n")
		} else {
			sb.WriteString("\t\t\"spec\": schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},\n")
		}
	}
	sb.WriteString(fmt.Sprintf("\t\t\"status\": commonInfoSchema(%s),\n", emitStatusSchemaArg(r)))
	sb.WriteString("\t}\n}\n\n")

	// Schema
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tresp.Schema = schema.Schema{Attributes: %sResourceSchemaAttrs()}\n}\n\n", sn))

	// Configure
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {\n", sn))
	sb.WriteString("\tif req.ProviderData == nil { return }\n")
	sb.WriteString("\tpd, ok := req.ProviderData.(*KvindoProviderData)\n")
	sb.WriteString("\tif !ok { resp.Diagnostics.AddError(\"Unexpected Provider Data\", fmt.Sprintf(\"Expected *KvindoProviderData, got %T\", req.ProviderData)); return }\n")
	sb.WriteString("\tr.client = pd.Client\n}\n\n")

	// Build request map
	sb.WriteString(fmt.Sprintf("func build%sRequestMap(ctx context.Context, plan %sResourceModel) map[string]interface{} {\n", sn, sn))
	sb.WriteString("\tm := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)\n")
	if hasSpec {
		sb.WriteString("\tspec := m[\"spec\"].(map[string]interface{})\n")
		for _, f := range r.Fields {
			writeSpecFieldToRequest(&sb, sn, f)
		}
	}
	sb.WriteString("\treturn m\n}\n\n")

	// Populate state
	sb.WriteString(fmt.Sprintf("func populate%sState(ctx context.Context, data map[string]interface{}, state *%sResourceModel) error {\n", sn, sn))
	sb.WriteString("\tif err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil { return err }\n")
	sb.WriteString("\tstate.ID = state.Metadata.ID\n")
	if hasSpec {
		sb.WriteString("\tspec := getSpec(data)\n")
		for _, f := range r.Fields {
			writeSpecFieldFromResponse(&sb, sn, f)
		}
	}
	emitStatusAssign(&sb, r, "data")
	sb.WriteString("\treturn nil\n}\n\n")

	// Create
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar plan %sResourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString("\tplan.ID = types.StringValue(newULID())\n")
	sb.WriteString(fmt.Sprintf("\tbody := build%sRequestMap(ctx, plan)\n", sn))
	sb.WriteString(fmt.Sprintf("\tmodResp, err := r.client.Put(ctx, %q, body)\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Create Error\", err.Error()); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := r.client.PollUntilDone(ctx, %q, modResp.RequestId); err != nil { resp.Diagnostics.AddError(\"Create Poll Error\", err.Error()); return }\n", r.APIPath))
	sb.WriteString("\tresourceId := modResp.ResourceId\n\tif resourceId == \"\" { resourceId = plan.ID.ValueString() }\n")
	sb.WriteString(fmt.Sprintf("\tapiData, err := r.client.Get(ctx, %q, resourceId)\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Read After Create Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.Diagnostics.AddError(\"Read After Create Error\", \"resource not found after creation\"); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := populate%sState(ctx, apiData, &plan); err != nil { resp.Diagnostics.AddError(\"State Error\", err.Error()); return }\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, plan)...)\n}\n\n")

	// Read
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar state %sResourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.State.Get(ctx, &state)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString(fmt.Sprintf("\tapiData, err := r.client.Get(ctx, %q, state.ID.ValueString())\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Read Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.State.RemoveResource(ctx); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := populate%sState(ctx, apiData, &state); err != nil { resp.Diagnostics.AddError(\"State Error\", err.Error()); return }\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, state)...)\n}\n\n")

	// Update
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar plan, state %sResourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)\n")
	sb.WriteString("\tresp.Diagnostics.Append(req.State.Get(ctx, &state)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString("\tplan.ID = state.ID\n")
	sb.WriteString(fmt.Sprintf("\tbody := build%sRequestMap(ctx, plan)\n", sn))
	sb.WriteString(fmt.Sprintf("\tmodResp, err := r.client.Put(ctx, %q, body)\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Update Error\", err.Error()); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := r.client.PollUntilDone(ctx, %q, modResp.RequestId); err != nil { resp.Diagnostics.AddError(\"Update Poll Error\", err.Error()); return }\n", r.APIPath))
	sb.WriteString(fmt.Sprintf("\tapiData, err := r.client.Get(ctx, %q, plan.ID.ValueString())\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Read After Update Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.Diagnostics.AddError(\"Read After Update Error\", \"not found\"); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := populate%sState(ctx, apiData, &plan); err != nil { resp.Diagnostics.AddError(\"State Error\", err.Error()); return }\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, plan)...)\n}\n\n")

	// Delete
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar state %sResourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.State.Get(ctx, &state)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString(fmt.Sprintf("\tmodResp, err := r.client.Delete(ctx, %q, state.ID.ValueString())\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Delete Error\", err.Error()); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := r.client.PollUntilDone(ctx, %q, modResp.RequestId); err != nil { resp.Diagnostics.AddError(\"Delete Poll Error\", err.Error()); return }\n}\n\n", r.APIPath))

	// ImportState
	sb.WriteString(fmt.Sprintf("func (r *%sResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar state %sResourceModel\n", sn))
	sb.WriteString("\tstate.ID = types.StringValue(req.ID)\n")
	sb.WriteString(fmt.Sprintf("\tapiData, err := r.client.Get(ctx, %q, req.ID)\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Import Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.Diagnostics.AddError(\"Import Error\", \"not found\"); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := populate%sState(ctx, apiData, &state); err != nil { resp.Diagnostics.AddError(\"State Error\", err.Error()); return }\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, state)...)\n}\n")

	return sb.String()
}

// resourceAttrDef returns the resource schema.Attribute literal for a spec field.
func resourceAttrDef(f FieldDef) string {
	sens := ""
	if f.Sensitive {
		sens = ", Sensitive: true"
	}
	switch f.FieldType {
	case "string":
		if f.Required {
			return fmt.Sprintf("schema.StringAttribute{Required: true%s}", sens)
		}
		if f.OptionalOnly {
			return fmt.Sprintf("schema.StringAttribute{Optional: true%s}", sens)
		}
		return fmt.Sprintf("schema.StringAttribute{Optional: true, Computed: true%s, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}", sens)
	case "bool":
		if f.Required {
			return "schema.BoolAttribute{Required: true}"
		}
		if f.OptionalOnly {
			return "schema.BoolAttribute{Optional: true}"
		}
		return "schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}}"
	case "int64":
		if f.Required {
			return "schema.Int64Attribute{Required: true}"
		}
		if f.OptionalOnly {
			return "schema.Int64Attribute{Optional: true}"
		}
		return "schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}}"
	case "float64":
		if f.Required {
			return "schema.Float64Attribute{Required: true}"
		}
		if f.OptionalOnly {
			return "schema.Float64Attribute{Optional: true}"
		}
		return "schema.Float64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Float64{float64planmodifier.UseStateForUnknown()}}"
	case "list_string":
		if f.Required {
			return "schema.ListAttribute{Required: true, ElementType: types.StringType}"
		}
		if f.OptionalOnly {
			return "schema.ListAttribute{Optional: true, ElementType: types.StringType}"
		}
		return "schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType}"
	case "map_string":
		if f.Required {
			return "schema.MapAttribute{Required: true, ElementType: types.StringType}"
		}
		if f.OptionalOnly {
			return "schema.MapAttribute{Optional: true, ElementType: types.StringType}"
		}
		return "schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType}"
	case "list_object":
		var b strings.Builder
		b.WriteString("schema.ListNestedAttribute{Optional: true, Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{")
		for _, of := range f.ObjFields {
			b.WriteString(fmt.Sprintf("%q: %s,", of.TFName, resourceAttrDef(of)))
		}
		b.WriteString("}}}")
		return b.String()
	}
	return "schema.StringAttribute{Optional: true, Computed: true}"
}

// resourceComputedAttrDef is retained for compatibility (status extras use it indirectly).
func resourceComputedAttrDef(f FieldDef) string {
	if f.Sensitive {
		return "schema.StringAttribute{Computed: true, Sensitive: true}"
	}
	switch f.FieldType {
	case "bool":
		return "schema.BoolAttribute{Computed: true}"
	case "int64":
		return "schema.Int64Attribute{Computed: true}"
	case "float64":
		return "schema.Float64Attribute{Computed: true}"
	}
	return "schema.StringAttribute{Computed: true}"
}

// writeSpecFieldToRequest emits code writing a spec field from plan.Spec into the request spec map.
func writeSpecFieldToRequest(sb *strings.Builder, sn string, f FieldDef) {
	F := toTitle(f.TFName)
	api := f.APIName
	switch f.FieldType {
	case "string":
		sb.WriteString(fmt.Sprintf("\tif !plan.Spec.%s.IsNull() && !plan.Spec.%s.IsUnknown() { spec[%q] = plan.Spec.%s.ValueString() }\n", F, F, api, F))
	case "bool":
		sb.WriteString(fmt.Sprintf("\tif !plan.Spec.%s.IsNull() && !plan.Spec.%s.IsUnknown() { spec[%q] = plan.Spec.%s.ValueBool() }\n", F, F, api, F))
	case "int64":
		sb.WriteString(fmt.Sprintf("\tif !plan.Spec.%s.IsNull() && !plan.Spec.%s.IsUnknown() { spec[%q] = plan.Spec.%s.ValueInt64() }\n", F, F, api, F))
	case "float64":
		sb.WriteString(fmt.Sprintf("\tif !plan.Spec.%s.IsNull() && !plan.Spec.%s.IsUnknown() { spec[%q] = plan.Spec.%s.ValueFloat64() }\n", F, F, api, F))
	case "list_string":
		sb.WriteString(fmt.Sprintf("\tif !plan.Spec.%s.IsNull() && !plan.Spec.%s.IsUnknown() { spec[%q] = stringListToInterface(ctx, plan.Spec.%s) }\n", F, F, api, F))
	case "map_string":
		sb.WriteString(fmt.Sprintf("\tif !plan.Spec.%s.IsNull() && !plan.Spec.%s.IsUnknown() { spec[%q] = stringMapToInterface(ctx, plan.Spec.%s) }\n", F, F, api, F))
	case "object":
		sb.WriteString(fmt.Sprintf("\tif !plan.Spec.%s.IsNull() && !plan.Spec.%s.IsUnknown() { spec[%q] = objToAPI(plan.Spec.%s, %s) }\n", F, F, api, F, descVarName(sn, f)))
	case "list_object":
		sb.WriteString(fmt.Sprintf("\tif !plan.Spec.%s.IsNull() && !plan.Spec.%s.IsUnknown() { spec[%q] = listObjToAPI(plan.Spec.%s, %s) }\n", F, F, api, F, descVarName(sn, f)))
	}
}

// writeSpecFieldFromResponse emits code reading a spec field from the response spec map into state.Spec.
func writeSpecFieldFromResponse(sb *strings.Builder, sn string, f FieldDef) {
	F := toTitle(f.TFName)
	api := f.APIName
	switch f.FieldType {
	case "string":
		sb.WriteString(fmt.Sprintf("\tstate.Spec.%s = getString(spec, %q)\n", F, api))
	case "bool":
		sb.WriteString(fmt.Sprintf("\tstate.Spec.%s = getBool(spec, %q)\n", F, api))
	case "int64":
		sb.WriteString(fmt.Sprintf("\tstate.Spec.%s = getInt64(spec, %q)\n", F, api))
	case "float64":
		sb.WriteString(fmt.Sprintf("\tstate.Spec.%s = getFloat64(spec, %q)\n", F, api))
	case "list_string":
		sb.WriteString(fmt.Sprintf("\tstate.Spec.%s = getStringList(ctx, spec, %q)\n", F, api))
	case "map_string":
		sb.WriteString(fmt.Sprintf("\tstate.Spec.%s = getStringMap(spec, %q)\n", F, api))
	case "object":
		sb.WriteString(fmt.Sprintf("\tstate.Spec.%s = objFromAPI(objMap(spec, %q), %s)\n", F, api, descVarName(sn, f)))
	case "list_object":
		sb.WriteString(fmt.Sprintf("\tstate.Spec.%s = listObjFromAPI(objList(spec, %q), %s)\n", F, api, descVarName(sn, f)))
	}
}

func generateDatasourceFile(r ResourceDef) string {
	sn := structName(r.Name)
	hasSpec := len(r.Fields) > 0
	var sb strings.Builder

	// attr is only needed for the status-extras buildInfoObj literals.
	needsAttr := len(r.StatusExtra) > 0

	sb.WriteString("package provider\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"fmt\"\n\n")
	if needsAttr {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/attr\"\n")
	}
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/datasource\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/datasource/schema\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/types\"\n")
	sb.WriteString("\t\"github.com/kvindo/terraform-provider-kvindo/internal/client\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("var _ = fmt.Sprintf\n\n")

	// Model (reuses the resource's spec struct)
	sb.WriteString(fmt.Sprintf("type %sDataSourceModel struct {\n", sn))
	sb.WriteString("\tID       types.String  `tfsdk:\"id\"`\n")
	sb.WriteString("\tName     types.String  `tfsdk:\"name\"`\n")
	sb.WriteString("\tMetadata metadataModel `tfsdk:\"metadata\"`\n")
	if hasSpec {
		sb.WriteString(fmt.Sprintf("\tSpec     %sSpecModel `tfsdk:\"spec\"`\n", sn))
	}
	sb.WriteString("\tStatus   types.Object  `tfsdk:\"status\"`\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("type %sDataSource struct { client *client.Client }\n\n", sn))
	sb.WriteString(fmt.Sprintf("func New%sDataSource() datasource.DataSource { return &%sDataSource{} }\n\n", sn, sn))

	sb.WriteString(fmt.Sprintf("func (d *%sDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tresp.TypeName = req.ProviderTypeName + \"_%s\"\n}\n\n", r.Name))

	sb.WriteString(fmt.Sprintf("func (d *%sDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {\n", sn))
	if hasSpec {
		sb.WriteString("\tspecAttrs := map[string]schema.Attribute{\n")
		for _, f := range r.Fields {
			var expr string
			switch f.FieldType {
			case "object":
				expr = fmt.Sprintf("objDatasourceSchema(%s)", descVarName(sn, f))
			case "list_object":
				expr = fmt.Sprintf("listObjDatasourceSchema(%s)", descVarName(sn, f))
			default:
				expr = datasourceAttrDef(f)
			}
			sb.WriteString(fmt.Sprintf("\t\t%q: %s,\n", f.TFName, expr))
		}
		sb.WriteString("\t}\n")
	}
	sb.WriteString("\tresp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{\n")
	// Look the resource up by either id or name (exactly one); the other is computed.
	sb.WriteString("\t\t\"id\": schema.StringAttribute{Optional: true, Computed: true},\n")
	sb.WriteString("\t\t\"name\": schema.StringAttribute{Optional: true, Computed: true},\n")
	sb.WriteString("\t\t\"metadata\": metadataDatasourceSchema(),\n")
	if hasSpec {
		sb.WriteString("\t\t\"spec\": schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},\n")
	}
	sb.WriteString(fmt.Sprintf("\t\t\"status\": commonInfoDatasourceSchema(%s),\n", emitDatasourceStatusSchemaArg(r)))
	sb.WriteString("\t}}\n}\n\n")

	sb.WriteString(fmt.Sprintf("func (d *%sDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {\n", sn))
	sb.WriteString("\tif req.ProviderData == nil { return }\n")
	sb.WriteString("\tpd, ok := req.ProviderData.(*KvindoProviderData)\n")
	sb.WriteString("\tif !ok { resp.Diagnostics.AddError(\"Unexpected Provider Data\", fmt.Sprintf(\"Expected *KvindoProviderData, got %T\", req.ProviderData)); return }\n")
	sb.WriteString("\td.client = pd.Client\n}\n\n")

	sb.WriteString(fmt.Sprintf("func (d *%sDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar state %sDataSourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.Config.Get(ctx, &state)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString("\tvar apiData map[string]interface{}\n\tvar err error\n")
	sb.WriteString("\tidSet := !state.ID.IsNull() && state.ID.ValueString() != \"\"\n")
	sb.WriteString("\tnameSet := !state.Name.IsNull() && state.Name.ValueString() != \"\"\n")
	sb.WriteString("\tif idSet == nameSet {\n")
	sb.WriteString("\t\tresp.Diagnostics.AddError(\"Invalid lookup\", \"exactly one of \\\"id\\\" or \\\"name\\\" must be set\"); return\n\t}\n")
	sb.WriteString("\tif idSet {\n")
	sb.WriteString(fmt.Sprintf("\t\tapiData, err = d.client.Get(ctx, %q, state.ID.ValueString())\n", r.APIPath))
	sb.WriteString("\t} else {\n")
	sb.WriteString(fmt.Sprintf("\t\tapiData, err = d.client.GetByName(ctx, %q, state.Name.ValueString())\n", r.APIPath))
	sb.WriteString("\t}\n")
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Read Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.Diagnostics.AddError(\"Not Found\", \"resource not found\"); return }\n")
	sb.WriteString("\tif err := setCommonFieldsNested(ctx, apiData, &state.Metadata); err != nil { resp.Diagnostics.AddError(\"State Error\", err.Error()); return }\n")
	sb.WriteString("\tstate.ID = state.Metadata.ID\n")
	sb.WriteString("\tstate.Name = state.Metadata.Name\n")
	if hasSpec {
		sb.WriteString("\tspec := getSpec(apiData)\n")
		for _, f := range r.Fields {
			writeSpecFieldFromResponse(&sb, sn, f)
		}
	}
	emitStatusAssign(&sb, r, "apiData")
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, state)...)\n}\n")

	return sb.String()
}

// datasourceAttrDef returns the datasource schema.Attribute literal for a spec field (all Computed).
func datasourceAttrDef(f FieldDef) string {
	switch f.FieldType {
	case "string":
		if f.Sensitive {
			return "schema.StringAttribute{Computed: true, Sensitive: true}"
		}
		return "schema.StringAttribute{Computed: true}"
	case "bool":
		return "schema.BoolAttribute{Computed: true}"
	case "int64":
		return "schema.Int64Attribute{Computed: true}"
	case "float64":
		return "schema.Float64Attribute{Computed: true}"
	case "list_string":
		return "schema.ListAttribute{Computed: true, ElementType: types.StringType}"
	case "map_string":
		return "schema.MapAttribute{Computed: true, ElementType: types.StringType}"
	case "list_object":
		var b strings.Builder
		b.WriteString("schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{")
		for _, of := range f.ObjFields {
			b.WriteString(fmt.Sprintf("%q: %s,", of.TFName, datasourceAttrDef(of)))
		}
		b.WriteString("}}}")
		return b.String()
	}
	return "schema.StringAttribute{Computed: true}"
}

// emitDatasourceStatusSchemaArg returns the arg for commonInfoDatasourceSchema.
func emitDatasourceStatusSchemaArg(r ResourceDef) string {
	if len(r.StatusExtra) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString("map[string]schema.Attribute{")
	for _, f := range r.StatusExtra {
		var attrType string
		switch f.FieldType {
		case "int64":
			attrType = "schema.Int64Attribute{Computed: true}"
		case "bool":
			attrType = "schema.BoolAttribute{Computed: true}"
		default:
			if f.Sensitive {
				attrType = "schema.StringAttribute{Computed: true, Sensitive: true}"
			} else {
				attrType = "schema.StringAttribute{Computed: true}"
			}
		}
		b.WriteString(fmt.Sprintf("%q: %s, ", f.TFName, attrType))
	}
	b.WriteString("}")
	return b.String()
}

func generateProviderFile(resources []ResourceDef) string {
	var sb strings.Builder
	sb.WriteString("package provider\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n\t\"os\"\n\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/datasource\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/provider\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/provider/schema\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/types\"\n")
	sb.WriteString("\t\"github.com/kvindo/terraform-provider-kvindo/internal/client\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("const defaultEndpoint = \"https://cloud-api.kvindo.com\"\n\n")
	sb.WriteString("var _ provider.Provider = &KvindoProvider{}\n\n")
	sb.WriteString("type KvindoProvider struct { version string }\n")
	sb.WriteString("type KvindoProviderModel struct {\n\tEndpoint types.String `tfsdk:\"endpoint\"`\n\tToken    types.String `tfsdk:\"token\"`\n}\n\n")
	sb.WriteString("func New(version string) func() provider.Provider {\n\treturn func() provider.Provider { return &KvindoProvider{version: version} }\n}\n\n")
	sb.WriteString("func (p *KvindoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {\n\tresp.TypeName = \"kvindo\"\n\tresp.Version = p.version\n}\n\n")
	sb.WriteString("func (p *KvindoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {\n")
	sb.WriteString("\tresp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{\n")
	sb.WriteString("\t\t\"endpoint\": schema.StringAttribute{Optional: true, Description: \"API endpoint, defaults to https://cloud-api.kvindo.com\"},\n")
	sb.WriteString("\t\t\"token\":    schema.StringAttribute{Optional: true, Sensitive: true, Description: \"API bearer token\"},\n")
	sb.WriteString("\t}}\n}\n\n")
	sb.WriteString("func (p *KvindoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {\n")
	sb.WriteString("\tvar config KvindoProviderModel\n\tresp.Diagnostics.Append(req.Config.Get(ctx, &config)...)\n\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString("\tendpoint := defaultEndpoint\n\tif !config.Endpoint.IsNull() && !config.Endpoint.IsUnknown() && config.Endpoint.ValueString() != \"\" {\n\t\tendpoint = config.Endpoint.ValueString()\n\t} else if v := os.Getenv(\"KVINDO_ENDPOINT\"); v != \"\" {\n\t\tendpoint = v\n\t}\n")
	sb.WriteString("\ttoken := \"\"\n\tif !config.Token.IsNull() && !config.Token.IsUnknown() { token = config.Token.ValueString() }\n\tif token == \"\" { token = os.Getenv(\"KVINDO_TOKEN\") }\n")
	sb.WriteString("\tif token == \"\" { resp.Diagnostics.AddError(\"Missing API Token\", \"Set token in provider config or KVINDO_TOKEN env var\"); return }\n")
	sb.WriteString("\tpd := &KvindoProviderData{Client: client.New(endpoint, token, p.version)}\n\tresp.DataSourceData = pd\n\tresp.ResourceData = pd\n}\n\n")

	// Resources
	sb.WriteString("func (p *KvindoProvider) Resources(_ context.Context) []func() resource.Resource {\n\treturn []func() resource.Resource{\n")
	for _, r := range resources {
		sb.WriteString(fmt.Sprintf("\t\tNew%sResource,\n", structName(r.Name)))
	}
	// Transaction is hand-written (not generated) but must be registered.
	sb.WriteString("\t\tNewTransactionResource,\n")
	sb.WriteString("\t}\n}\n\n")

	// DataSources
	sb.WriteString("func (p *KvindoProvider) DataSources(_ context.Context) []func() datasource.DataSource {\n\treturn []func() datasource.DataSource{\n")
	for _, r := range resources {
		sb.WriteString(fmt.Sprintf("\t\tNew%sDataSource,\n", structName(r.Name)))
	}
	sb.WriteString("\t}\n}\n")

	return sb.String()
}
