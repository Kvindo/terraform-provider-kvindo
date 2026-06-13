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
	TFName    string
	APIName   string
	FieldType string // "string", "bool", "int64", "float64", "list_string", "map_string", "list_object"
	Required  bool
	Computed  bool
	Sensitive bool
	ObjFields []FieldDef
}

// ResourceDef describes a complete resource to generate.
type ResourceDef struct {
	Name       string
	APIPath    string
	Fields     []FieldDef
	InfoFields []FieldDef
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

		var fields, infoFields []FieldDef

		if item.Put.RequestBody != nil {
			if ct, ok := item.Put.RequestBody.Content["application/json"]; ok && ct.Schema != nil {
				schemaName := extractRefName(ct.Schema.Ref)
				if schemaName != "" {
					if schema, ok := spec.Components.Schemas[schemaName]; ok {
						fields, infoFields = extractFields(schema, spec.Components.Schemas)
					}
				}
			}
		}

		resources = append(resources, ResourceDef{
			Name:       resourceName,
			APIPath:    path,
			Fields:     fields,
			InfoFields: infoFields,
		})
	}

	return resources
}

// commonFields are fields handled by common schema and should be skipped in resource-specific fields.
var commonFieldNames = map[string]bool{
	"id": true, "name": true, "description": true, "labels": true,
	"folderId": true, "deleteProtection": true,
}

// sensitiveKeywords indicate a field should be marked sensitive.
var sensitiveKeywords = []string{
	"password", "token", "key", "secret", "pem", "config", "kubeconfig", "privateKey",
}

func isSensitive(name string) bool {
	lower := strings.ToLower(name)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
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
			// Extract info sub-fields
			refName := extractRefName(prop.Ref)
			if infoSchema, ok := schemas[refName]; ok {
				for infoFieldName, infoFieldProp := range infoSchema.Properties {
					infoTFName := "info_" + camelToSnake(infoFieldName)
					ft := swaggerTypeToFieldType(infoFieldProp)
					infoFields = append(infoFields, FieldDef{
						TFName:    infoTFName,
						APIName:   "info." + infoFieldName,
						FieldType: ft,
						Computed:  true,
						Sensitive: isSensitive(infoFieldName),
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

		if ft == "list_object" && prop.Items != nil {
			if prop.Items.Ref != "" {
				refName := extractRefName(prop.Items.Ref)
				if objSchema, ok := schemas[refName]; ok {
					for objFieldName, objProp := range objSchema.Properties {
						objFields = append(objFields, FieldDef{
							TFName:    camelToSnake(objFieldName),
							APIName:   objFieldName,
							FieldType: swaggerTypeToFieldType(objProp),
						})
					}
				}
			} else if prop.Items.Properties != nil {
				for objFieldName, objProp := range prop.Items.Properties {
					objFields = append(objFields, FieldDef{
						TFName:    camelToSnake(objFieldName),
						APIName:   objFieldName,
						FieldType: swaggerTypeToFieldType(objProp),
					})
				}
			}
		}

		fields = append(fields, FieldDef{
			TFName:    tfName,
			APIName:   apiName,
			FieldType: ft,
			Sensitive: isSensitive(apiName),
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
			// array of objects
			return "list_object"
		}
		return "list_string"
	case "object":
		if prop.Properties == nil && prop.Items == nil {
			return "map_string"
		}
		return "map_string"
	default:
		return "string"
	}
}

func extractRefName(ref string) string {
	if ref == "" {
		return ""
	}
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

func camelToSnake(s string) string {
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
	default:
		return "types.String"
	}
}

func generateResourceFile(r ResourceDef) string {
	sn := structName(r.Name)
	var sb strings.Builder

	// Determine needed imports
	needsFloat64 := false
	needsAttr := false
	needsPlanmodifier := false
	needsBool := false
	needsInt64 := false
	needsString := false

	for _, f := range r.Fields {
		if f.FieldType == "float64" {
			needsFloat64 = true
		}
		if f.FieldType == "list_object" {
			needsAttr = true
		}
		if !f.Required && !f.Computed {
			needsPlanmodifier = true
			if f.FieldType == "bool" {
				needsBool = true
			} else if f.FieldType == "int64" {
				needsInt64 = true
			} else if f.FieldType == "float64" {
				needsFloat64 = true
			} else if f.FieldType == "string" {
				needsString = true
			}
		}
	}
	// All resources have common fields which use planmodifier
	needsPlanmodifier = true
	needsString = true
	needsBool = true
	needsInt64 = true

	sb.WriteString("package provider\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"fmt\"\n\n")
	if needsAttr {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/attr\"\n")
	}
	sb.WriteString("\t\"github.com/google/uuid\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema\"\n")
	if needsBool && needsPlanmodifier {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier\"\n")
	}
	if needsFloat64 && needsPlanmodifier {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier\"\n")
	}
	if needsInt64 && needsPlanmodifier {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier\"\n")
	}
	if needsPlanmodifier {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier\"\n")
	}
	if needsString && needsPlanmodifier {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier\"\n")
	}
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/types\"\n")
	sb.WriteString("\t\"github.com/kvindo/terraform-provider-kvindo/internal/client\"\n")
	sb.WriteString(")\n\n")

	sb.WriteString("var _ = fmt.Sprintf\n\n")

	// Model
	sb.WriteString(fmt.Sprintf("type %sResourceModel struct {\n", sn))
	sb.WriteString("\tID               types.String `tfsdk:\"id\"`\n")
	sb.WriteString("\tName             types.String `tfsdk:\"name\"`\n")
	sb.WriteString("\tDescription      types.String `tfsdk:\"description\"`\n")
	sb.WriteString("\tFolderID         types.String `tfsdk:\"folder_id\"`\n")
	sb.WriteString("\tDeleteProtection types.Bool   `tfsdk:\"delete_protection\"`\n")
	sb.WriteString("\tLabels           types.Map    `tfsdk:\"labels\"`\n")
	for _, f := range r.Fields {
		sb.WriteString(fmt.Sprintf("\t%s %s `tfsdk:\"%s\"`\n", toTitle(f.TFName), typeToGoType(f.FieldType), f.TFName))
	}
	for _, f := range r.InfoFields {
		sb.WriteString(fmt.Sprintf("\t%s %s `tfsdk:\"%s\"`\n", toTitle(f.TFName), typeToGoType(f.FieldType), f.TFName))
	}
	sb.WriteString("}\n\n")

	// Resource struct
	sb.WriteString(fmt.Sprintf("type %sResource struct { client *client.Client }\n\n", sn))
	sb.WriteString(fmt.Sprintf("func New%sResource() resource.Resource { return &%sResource{} }\n\n", sn, sn))

	// Metadata
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tresp.TypeName = req.ProviderTypeName + \"_%s\"\n}\n\n", r.Name))

	// Schema
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {\n", sn))
	sb.WriteString("\tattrs := commonSchemaAttributes()\n")
	for _, f := range r.Fields {
		sb.WriteString(fmt.Sprintf("\tattrs[\"%s\"] = %s\n", f.TFName, resourceAttrDef(f)))
	}
	for _, f := range r.InfoFields {
		sb.WriteString(fmt.Sprintf("\tattrs[\"%s\"] = %s\n", f.TFName, resourceComputedAttrDef(f)))
	}
	sb.WriteString("\tresp.Schema = schema.Schema{Attributes: attrs}\n}\n\n")

	// Configure
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {\n", sn))
	sb.WriteString("\tif req.ProviderData == nil { return }\n")
	sb.WriteString("\tpd, ok := req.ProviderData.(*KvindoProviderData)\n")
	sb.WriteString("\tif !ok { resp.Diagnostics.AddError(\"Unexpected Provider Data\", fmt.Sprintf(\"Expected *KvindoProviderData, got %T\", req.ProviderData)); return }\n")
	sb.WriteString("\tr.client = pd.Client\n}\n\n")

	// Build request map
	sb.WriteString(fmt.Sprintf("func build%sRequest(ctx context.Context, plan %sResourceModel) map[string]interface{} {\n", sn, sn))
	sb.WriteString("\tm := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)\n")
	for _, f := range r.Fields {
		writeFieldToRequest(&sb, f)
	}
	sb.WriteString("\treturn m\n}\n\n")

	// Populate state
	sb.WriteString(fmt.Sprintf("func populate%sState(ctx context.Context, data map[string]interface{}, state *%sResourceModel) error {\n", sn, sn))
	sb.WriteString("\tif err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil { return err }\n")
	for _, f := range r.Fields {
		writeFieldFromResponse(&sb, f, "state")
	}
	for _, f := range r.InfoFields {
		writeInfoFieldFromResponse(&sb, f, "state")
	}
	sb.WriteString("\treturn nil\n}\n\n")

	// CRUD methods
	sb.WriteString(fmt.Sprintf("func (r *%sResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar plan %sResourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString("\tplan.ID = types.StringValue(uuid.New().String())\n")
	sb.WriteString(fmt.Sprintf("\tbody := build%sRequest(ctx, plan)\n", sn))
	sb.WriteString(fmt.Sprintf("\tmodResp, err := r.client.Put(ctx, \"%s\", body)\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Create Error\", err.Error()); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := r.client.PollUntilDone(ctx, \"%s\", modResp.RequestId); err != nil { resp.Diagnostics.AddError(\"Create Poll Error\", err.Error()); return }\n", r.APIPath))
	sb.WriteString("\tresourceId := modResp.ResourceId\n\tif resourceId == \"\" { resourceId = plan.ID.ValueString() }\n")
	sb.WriteString(fmt.Sprintf("\tapiData, err := r.client.Get(ctx, \"%s\", resourceId)\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Read After Create Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.Diagnostics.AddError(\"Read After Create Error\", \"resource not found after creation\"); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := populate%sState(ctx, apiData, &plan); err != nil { resp.Diagnostics.AddError(\"State Error\", err.Error()); return }\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, plan)...)\n}\n\n")

	sb.WriteString(fmt.Sprintf("func (r *%sResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar state %sResourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.State.Get(ctx, &state)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString(fmt.Sprintf("\tapiData, err := r.client.Get(ctx, \"%s\", state.ID.ValueString())\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Read Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.State.RemoveResource(ctx); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := populate%sState(ctx, apiData, &state); err != nil { resp.Diagnostics.AddError(\"State Error\", err.Error()); return }\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, state)...)\n}\n\n")

	sb.WriteString(fmt.Sprintf("func (r *%sResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar plan, state %sResourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)\n")
	sb.WriteString("\tresp.Diagnostics.Append(req.State.Get(ctx, &state)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString("\tplan.ID = state.ID\n")
	sb.WriteString(fmt.Sprintf("\tbody := build%sRequest(ctx, plan)\n", sn))
	sb.WriteString(fmt.Sprintf("\tmodResp, err := r.client.Put(ctx, \"%s\", body)\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Update Error\", err.Error()); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := r.client.PollUntilDone(ctx, \"%s\", modResp.RequestId); err != nil { resp.Diagnostics.AddError(\"Update Poll Error\", err.Error()); return }\n", r.APIPath))
	sb.WriteString(fmt.Sprintf("\tapiData, err := r.client.Get(ctx, \"%s\", plan.ID.ValueString())\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Read After Update Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.Diagnostics.AddError(\"Read After Update Error\", \"not found\"); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := populate%sState(ctx, apiData, &plan); err != nil { resp.Diagnostics.AddError(\"State Error\", err.Error()); return }\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, plan)...)\n}\n\n")

	sb.WriteString(fmt.Sprintf("func (r *%sResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar state %sResourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.State.Get(ctx, &state)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString(fmt.Sprintf("\tmodResp, err := r.client.Delete(ctx, \"%s\", state.ID.ValueString())\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Delete Error\", err.Error()); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := r.client.PollUntilDone(ctx, \"%s\", modResp.RequestId); err != nil { resp.Diagnostics.AddError(\"Delete Poll Error\", err.Error()); return }\n}\n\n", r.APIPath))

	sb.WriteString(fmt.Sprintf("func (r *%sResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar state %sResourceModel\n", sn))
	sb.WriteString("\tstate.ID = types.StringValue(req.ID)\n")
	sb.WriteString(fmt.Sprintf("\tapiData, err := r.client.Get(ctx, \"%s\", req.ID)\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Import Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.Diagnostics.AddError(\"Import Error\", \"not found\"); return }\n")
	sb.WriteString(fmt.Sprintf("\tif err := populate%sState(ctx, apiData, &state); err != nil { resp.Diagnostics.AddError(\"State Error\", err.Error()); return }\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, state)...)\n}\n")

	_ = needsAttr
	return sb.String()
}

func resourceAttrDef(f FieldDef) string {
	switch f.FieldType {
	case "string":
		req := ""
		if f.Required {
			req = "Required: true,"
		} else {
			req = "Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},"
		}
		sens := ""
		if f.Sensitive {
			sens = "Sensitive: true,"
		}
		return fmt.Sprintf("schema.StringAttribute{%s %s}", req, sens)
	case "bool":
		req := ""
		if f.Required {
			req = "Required: true,"
		} else {
			req = "Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},"
		}
		return fmt.Sprintf("schema.BoolAttribute{%s}", req)
	case "int64":
		req := ""
		if f.Required {
			req = "Required: true,"
		} else {
			req = "Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},"
		}
		return fmt.Sprintf("schema.Int64Attribute{%s}", req)
	case "float64":
		req := ""
		if f.Required {
			req = "Required: true,"
		} else {
			req = "Optional: true, Computed: true, PlanModifiers: []planmodifier.Float64{float64planmodifier.UseStateForUnknown()},"
		}
		return fmt.Sprintf("schema.Float64Attribute{%s}", req)
	case "list_string":
		req := ""
		if f.Required {
			req = "Required: true,"
		} else {
			req = "Optional: true, Computed: true,"
		}
		return fmt.Sprintf("schema.ListAttribute{%s ElementType: types.StringType}", req)
	case "map_string":
		req := ""
		if f.Required {
			req = "Required: true,"
		} else {
			req = "Optional: true, Computed: true,"
		}
		return fmt.Sprintf("schema.MapAttribute{%s ElementType: types.StringType}", req)
	case "list_object":
		var nested strings.Builder
		nested.WriteString("schema.ListNestedAttribute{Optional: true, Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{")
		for _, of := range f.ObjFields {
			nested.WriteString(fmt.Sprintf("\"%s\": %s,", of.TFName, resourceAttrDef(of)))
		}
		nested.WriteString("}}}")
		return nested.String()
	}
	return "schema.StringAttribute{Optional: true, Computed: true}"
}

func resourceComputedAttrDef(f FieldDef) string {
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
	}
	return "schema.StringAttribute{Computed: true}"
}

func writeFieldToRequest(sb *strings.Builder, f FieldDef) {
	tfField := toTitle(f.TFName)
	apiName := f.APIName

	switch f.FieldType {
	case "string":
		sb.WriteString(fmt.Sprintf("\tif !plan.%s.IsNull() && !plan.%s.IsUnknown() { m[\"%s\"] = plan.%s.ValueString() }\n", tfField, tfField, apiName, tfField))
	case "bool":
		sb.WriteString(fmt.Sprintf("\tif !plan.%s.IsNull() && !plan.%s.IsUnknown() { m[\"%s\"] = plan.%s.ValueBool() }\n", tfField, tfField, apiName, tfField))
	case "int64":
		sb.WriteString(fmt.Sprintf("\tif !plan.%s.IsNull() && !plan.%s.IsUnknown() { m[\"%s\"] = plan.%s.ValueInt64() }\n", tfField, tfField, apiName, tfField))
	case "float64":
		sb.WriteString(fmt.Sprintf("\tif !plan.%s.IsNull() && !plan.%s.IsUnknown() { m[\"%s\"] = plan.%s.ValueFloat64() }\n", tfField, tfField, apiName, tfField))
	case "list_string":
		sb.WriteString(fmt.Sprintf("\tif !plan.%s.IsNull() && !plan.%s.IsUnknown() { m[\"%s\"] = stringListToInterface(ctx, plan.%s) }\n", tfField, tfField, apiName, tfField))
	case "map_string":
		sb.WriteString(fmt.Sprintf("\tif !plan.%s.IsNull() && !plan.%s.IsUnknown() { m[\"%s\"] = stringMapToInterface(ctx, plan.%s) }\n", tfField, tfField, apiName, tfField))
	case "list_object":
		sb.WriteString(fmt.Sprintf("\tif !plan.%s.IsNull() && !plan.%s.IsUnknown() {\n", tfField, tfField))
		sb.WriteString("\t\tvar items []map[string]interface{}\n")
		sb.WriteString(fmt.Sprintf("\t\tfor _, elem := range plan.%s.Elements() {\n", tfField))
		sb.WriteString("\t\t\tif ov, ok := elem.(types.Object); ok {\n")
		sb.WriteString("\t\t\t\titem := map[string]interface{}{}\n")
		for _, of := range f.ObjFields {
			switch of.FieldType {
			case "string":
				sb.WriteString(fmt.Sprintf("\t\t\t\tif v, ok := ov.Attributes()[\"%s\"]; ok { if sv, ok2 := v.(types.String); ok2 && !sv.IsNull() { item[\"%s\"] = sv.ValueString() } }\n", of.TFName, of.APIName))
			case "int64":
				sb.WriteString(fmt.Sprintf("\t\t\t\tif v, ok := ov.Attributes()[\"%s\"]; ok { if iv, ok2 := v.(types.Int64); ok2 && !iv.IsNull() { item[\"%s\"] = iv.ValueInt64() } }\n", of.TFName, of.APIName))
			case "bool":
				sb.WriteString(fmt.Sprintf("\t\t\t\tif v, ok := ov.Attributes()[\"%s\"]; ok { if bv, ok2 := v.(types.Bool); ok2 && !bv.IsNull() { item[\"%s\"] = bv.ValueBool() } }\n", of.TFName, of.APIName))
			}
		}
		sb.WriteString("\t\t\t\titems = append(items, item)\n\t\t\t}\n\t\t}\n")
		sb.WriteString(fmt.Sprintf("\t\tm[\"%s\"] = items\n\t}\n", apiName))
	}
}

func writeFieldFromResponse(sb *strings.Builder, f FieldDef, stateVar string) {
	tfField := toTitle(f.TFName)
	apiName := f.APIName

	switch f.FieldType {
	case "string":
		sb.WriteString(fmt.Sprintf("\t%s.%s = getString(data, \"%s\")\n", stateVar, tfField, apiName))
	case "bool":
		sb.WriteString(fmt.Sprintf("\t%s.%s = getBool(data, \"%s\")\n", stateVar, tfField, apiName))
	case "int64":
		sb.WriteString(fmt.Sprintf("\t%s.%s = getInt64(data, \"%s\")\n", stateVar, tfField, apiName))
	case "float64":
		sb.WriteString(fmt.Sprintf("\t%s.%s = getFloat64(data, \"%s\")\n", stateVar, tfField, apiName))
	case "list_string":
		sb.WriteString(fmt.Sprintf("\t%s.%s = getStringList(ctx, data, \"%s\")\n", stateVar, tfField, apiName))
	case "map_string":
		sb.WriteString(fmt.Sprintf("\t%s.%s = getStringMap(data, \"%s\")\n", stateVar, tfField, apiName))
	case "list_object":
		sb.WriteString(fmt.Sprintf("\t{\n\t\trawList, _ := data[\"%s\"].([]interface{})\n", apiName))
		sb.WriteString("\t\tattrTypes := map[string]attr.Type{\n")
		for _, of := range f.ObjFields {
			switch of.FieldType {
			case "string":
				sb.WriteString(fmt.Sprintf("\t\t\t\"%s\": types.StringType,\n", of.TFName))
			case "bool":
				sb.WriteString(fmt.Sprintf("\t\t\t\"%s\": types.BoolType,\n", of.TFName))
			case "int64":
				sb.WriteString(fmt.Sprintf("\t\t\t\"%s\": types.Int64Type,\n", of.TFName))
			}
		}
		sb.WriteString("\t\t}\n\t\tobjs := make([]attr.Value, 0, len(rawList))\n")
		sb.WriteString("\t\tfor _, item := range rawList {\n\t\t\tif m, ok := item.(map[string]interface{}); ok {\n")
		sb.WriteString("\t\t\t\tattrs := map[string]attr.Value{\n")
		for _, of := range f.ObjFields {
			switch of.FieldType {
			case "string":
				sb.WriteString(fmt.Sprintf("\t\t\t\t\t\"%s\": getString(m, \"%s\"),\n", of.TFName, of.APIName))
			case "bool":
				sb.WriteString(fmt.Sprintf("\t\t\t\t\t\"%s\": getBool(m, \"%s\"),\n", of.TFName, of.APIName))
			case "int64":
				sb.WriteString(fmt.Sprintf("\t\t\t\t\t\"%s\": getInt64(m, \"%s\"),\n", of.TFName, of.APIName))
			}
		}
		sb.WriteString("\t\t\t\t}\n\t\t\t\tobj, _ := types.ObjectValue(attrTypes, attrs)\n")
		sb.WriteString("\t\t\t\tobjs = append(objs, obj)\n\t\t\t}\n\t\t}\n")
		sb.WriteString(fmt.Sprintf("\t\t%s.%s, _ = types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objs)\n\t}\n", stateVar, tfField))
	}
}

func writeInfoFieldFromResponse(sb *strings.Builder, f FieldDef, stateVar string) {
	tfField := toTitle(f.TFName)
	// Strip "info." prefix from APIName to get the info sub-key
	infoKey := strings.TrimPrefix(f.APIName, "info.")

	switch f.FieldType {
	case "string":
		sb.WriteString(fmt.Sprintf("\t%s.%s = getStringFromInfo(data, \"%s\")\n", stateVar, tfField, infoKey))
	case "bool":
		sb.WriteString(fmt.Sprintf("\t%s.%s = getBoolFromInfo(data, \"%s\")\n", stateVar, tfField, infoKey))
	case "int64":
		sb.WriteString(fmt.Sprintf("\t%s.%s = getInt64FromInfo(data, \"%s\")\n", stateVar, tfField, infoKey))
	default:
		sb.WriteString(fmt.Sprintf("\t%s.%s = getStringFromInfo(data, \"%s\")\n", stateVar, tfField, infoKey))
	}
}

func generateDatasourceFile(r ResourceDef) string {
	sn := structName(r.Name)
	var sb strings.Builder

	needsAttr := false
	for _, f := range r.Fields {
		if f.FieldType == "list_object" {
			needsAttr = true
		}
	}

	sb.WriteString("package provider\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"fmt\"\n\n")
	if needsAttr {
		sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/attr\"\n")
	}
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/datasource\"\n")
	sb.WriteString("\tdschema \"github.com/hashicorp/terraform-plugin-framework/datasource/schema\"\n")
	sb.WriteString("\t\"github.com/hashicorp/terraform-plugin-framework/types\"\n")
	sb.WriteString("\t\"github.com/kvindo/terraform-provider-kvindo/internal/client\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("var _ = fmt.Sprintf\n\n")

	// Model
	sb.WriteString(fmt.Sprintf("type %sDataSourceModel struct {\n", sn))
	sb.WriteString("\tID               types.String `tfsdk:\"id\"`\n")
	sb.WriteString("\tName             types.String `tfsdk:\"name\"`\n")
	sb.WriteString("\tDescription      types.String `tfsdk:\"description\"`\n")
	sb.WriteString("\tFolderID         types.String `tfsdk:\"folder_id\"`\n")
	sb.WriteString("\tDeleteProtection types.Bool   `tfsdk:\"delete_protection\"`\n")
	sb.WriteString("\tLabels           types.Map    `tfsdk:\"labels\"`\n")
	for _, f := range r.Fields {
		sb.WriteString(fmt.Sprintf("\t%s %s `tfsdk:\"%s\"`\n", toTitle(f.TFName), typeToGoType(f.FieldType), f.TFName))
	}
	for _, f := range r.InfoFields {
		sb.WriteString(fmt.Sprintf("\t%s %s `tfsdk:\"%s\"`\n", toTitle(f.TFName), typeToGoType(f.FieldType), f.TFName))
	}
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("type %sDataSource struct { client *client.Client }\n\n", sn))
	sb.WriteString(fmt.Sprintf("func New%sDataSource() datasource.DataSource { return &%sDataSource{} }\n\n", sn, sn))

	sb.WriteString(fmt.Sprintf("func (d *%sDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tresp.TypeName = req.ProviderTypeName + \"_%s\"\n}\n\n", r.Name))

	sb.WriteString(fmt.Sprintf("func (d *%sDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {\n", sn))
	sb.WriteString("\tattrs := commonDatasourceSchemaAttributes()\n")
	for _, f := range r.Fields {
		sb.WriteString(fmt.Sprintf("\tattrs[\"%s\"] = %s\n", f.TFName, datasourceAttrDef(f)))
	}
	for _, f := range r.InfoFields {
		sb.WriteString(fmt.Sprintf("\tattrs[\"%s\"] = %s\n", f.TFName, datasourceComputedAttrDef(f)))
	}
	sb.WriteString("\tresp.Schema = datasource.Schema{Attributes: attrs}\n}\n\n")

	sb.WriteString(fmt.Sprintf("func (d *%sDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {\n", sn))
	sb.WriteString("\tif req.ProviderData == nil { return }\n")
	sb.WriteString("\tpd, ok := req.ProviderData.(*KvindoProviderData)\n")
	sb.WriteString("\tif !ok { resp.Diagnostics.AddError(\"Unexpected Provider Data\", fmt.Sprintf(\"Expected *KvindoProviderData, got %T\", req.ProviderData)); return }\n")
	sb.WriteString("\td.client = pd.Client\n}\n\n")

	sb.WriteString(fmt.Sprintf("func (d *%sDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {\n", sn))
	sb.WriteString(fmt.Sprintf("\tvar state %sDataSourceModel\n", sn))
	sb.WriteString("\tresp.Diagnostics.Append(req.Config.Get(ctx, &state)...)\n")
	sb.WriteString("\tif resp.Diagnostics.HasError() { return }\n")
	sb.WriteString(fmt.Sprintf("\tapiData, err := d.client.Get(ctx, \"%s\", state.ID.ValueString())\n", r.APIPath))
	sb.WriteString("\tif err != nil { resp.Diagnostics.AddError(\"Read Error\", err.Error()); return }\n")
	sb.WriteString("\tif apiData == nil { resp.Diagnostics.AddError(\"Not Found\", \"resource not found\"); return }\n")
	sb.WriteString("\tif err := setCommonFields(ctx, apiData, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {\n")
	sb.WriteString("\t\tresp.Diagnostics.AddError(\"State Error\", err.Error()); return\n\t}\n")
	for _, f := range r.Fields {
		writeFieldFromResponse(&sb, f, "state")
	}
	// Fix: data -> apiData
	// (writeFieldFromResponse uses "data" as the variable name - need to replace)
	for _, f := range r.InfoFields {
		writeInfoFieldFromResponse(&sb, f, "state")
	}
	sb.WriteString("\tresp.Diagnostics.Append(resp.State.Set(ctx, state)...)\n}\n")

	_ = needsAttr
	content := sb.String()
	// Fix data variable references: writeFieldFromResponse uses "data" not "apiData"
	content = strings.ReplaceAll(content, "getString(data,", "getString(apiData,")
	content = strings.ReplaceAll(content, "getBool(data,", "getBool(apiData,")
	content = strings.ReplaceAll(content, "getInt64(data,", "getInt64(apiData,")
	content = strings.ReplaceAll(content, "getFloat64(data,", "getFloat64(apiData,")
	content = strings.ReplaceAll(content, "getStringList(ctx, data,", "getStringList(ctx, apiData,")
	content = strings.ReplaceAll(content, "getStringMap(data,", "getStringMap(apiData,")
	content = strings.ReplaceAll(content, "getStringFromInfo(data,", "getStringFromInfo(apiData,")
	content = strings.ReplaceAll(content, "getInt64FromInfo(data,", "getInt64FromInfo(apiData,")
	content = strings.ReplaceAll(content, "getBoolFromInfo(data,", "getBoolFromInfo(apiData,")
	content = strings.ReplaceAll(content, `data["`, `apiData["`)
	content = strings.ReplaceAll(content, "dschema.", "schema.")  // fix alias usage
	return content
}

func datasourceAttrDef(f FieldDef) string {
	switch f.FieldType {
	case "string":
		if f.Sensitive {
			return "dschema.StringAttribute{Computed: true, Sensitive: true}"
		}
		return "dschema.StringAttribute{Computed: true}"
	case "bool":
		return "dschema.BoolAttribute{Computed: true}"
	case "int64":
		return "dschema.Int64Attribute{Computed: true}"
	case "float64":
		return "dschema.Float64Attribute{Computed: true}"
	case "list_string":
		return "dschema.ListAttribute{Computed: true, ElementType: types.StringType}"
	case "map_string":
		return "dschema.MapAttribute{Computed: true, ElementType: types.StringType}"
	case "list_object":
		var nested strings.Builder
		nested.WriteString("dschema.ListNestedAttribute{Computed: true, NestedObject: dschema.NestedAttributeObject{Attributes: map[string]dschema.Attribute{")
		for _, of := range f.ObjFields {
			nested.WriteString(fmt.Sprintf("\"%s\": %s,", of.TFName, datasourceAttrDef(of)))
		}
		nested.WriteString("}}}")
		return nested.String()
	}
	return "dschema.StringAttribute{Computed: true}"
}

func datasourceComputedAttrDef(f FieldDef) string {
	switch f.FieldType {
	case "string":
		if f.Sensitive {
			return "dschema.StringAttribute{Computed: true, Sensitive: true}"
		}
		return "dschema.StringAttribute{Computed: true}"
	case "bool":
		return "dschema.BoolAttribute{Computed: true}"
	case "int64":
		return "dschema.Int64Attribute{Computed: true}"
	default:
		return "dschema.StringAttribute{Computed: true}"
	}
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
	sb.WriteString("\tpd := &KvindoProviderData{Client: client.New(endpoint, token)}\n\tresp.DataSourceData = pd\n\tresp.ResourceData = pd\n}\n\n")

	// Resources
	sb.WriteString("func (p *KvindoProvider) Resources(_ context.Context) []func() resource.Resource {\n\treturn []func() resource.Resource{\n")
	for _, r := range resources {
		sb.WriteString(fmt.Sprintf("\t\tNew%sResource,\n", structName(r.Name)))
	}
	sb.WriteString("\t}\n}\n\n")

	// DataSources
	sb.WriteString("func (p *KvindoProvider) DataSources(_ context.Context) []func() datasource.DataSource {\n\treturn []func() datasource.DataSource{\n")
	for _, r := range resources {
		sb.WriteString(fmt.Sprintf("\t\tNew%sDataSource,\n", structName(r.Name)))
	}
	sb.WriteString("\t}\n}\n")

	return sb.String()
}
