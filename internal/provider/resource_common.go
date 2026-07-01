package provider

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
	"github.com/oklog/ulid/v2"
)

// newULID generates a new random ULID string (26-char Crockford base32, lowercase).
// The API always stores ULIDs in lowercase; generating lowercase avoids case-mismatch drift
// on terraform plan after resources are created (state has generated IDs, API returns lowercase).
func newULID() string {
	return strings.ToLower(ulid.MustNew(ulid.Now(), rand.Reader).String())
}

// KvindoProviderData holds the configured API client.
type KvindoProviderData struct {
	Client *client.Client
}

// commonSchemaAttributes returns the schema attributes common to all resources.
func commonSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"description": schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"folder_id": schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"delete_protection": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"labels": schema.MapAttribute{
			Optional:    true,
			Computed:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.Map{
				mapplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// isStableState reports whether a server "state" value is the terminal "stable" value.
// "stable" is the only settled state; scheduling / reconciling / schedulingfailed are
// all transient or in-flight (the reconciler will move them).
func isStableState(s string) bool {
	return s == "stable"
}

// volatileInfoModifier is the plan modifier for the "info" block. The "state" field inside
// info is VOLATILE — the server transitions it during reconciliation (scheduling ->
// reconciling -> stable / schedulingfailed). That makes a plain UseStateForUnknown wrong:
// it would freeze a stale, non-terminal value into the plan, and the apply (which polls to
// completion) then reads a different value, producing "Provider produced inconsistent
// result after apply".
//
// The fix is terminal-gated freezing, keyed off the PRIOR "state" value:
//
//   - prior == "stable": the resource is settled. Freeze the whole object to prior state,
//     exactly like UseStateForUnknown. No spurious "(known after apply)" diff on idle plans,
//     and safe because re-reading a settled resource yields the same values.
//
//   - prior != "stable" (in-flight, failed, or a value left behind by an interrupted apply):
//     leave the object unknown. The apply re-resolves it and accepts whatever terminal value
//     the reconciler produces — no inconsistency error. At most one corrective apply is
//     needed, after which the resource is "stable" and falls into the freeze case above.
type volatileInfoModifier struct{}

func (m volatileInfoModifier) Description(_ context.Context) string {
	return "Reuses prior state when the resource is stable; recomputes on apply while it is in-flight."
}

func (m volatileInfoModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m volatileInfoModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// No prior state (create, or a sub-resource never saved to state, e.g. after Ctrl+C)
	// -> leave the object unknown so the post-apply read fills it in.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	// Plan value already known -> keep it.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Freeze to prior state ONLY when the prior state is the terminal "stable" value.
	sv, ok := req.StateValue.Attributes()["state"].(types.String)
	if !ok || sv.IsNull() || sv.IsUnknown() {
		return
	}
	if isStableState(sv.ValueString()) {
		resp.PlanValue = req.StateValue
	}
	// else: leave unknown -> object re-resolves on apply.
}

// volatileStateModifier is the string-attribute analogue of volatileInfoModifier, for a
// top-level "state" field that is not wrapped in an info object (the transaction's own
// state). Same terminal-gated rule: freeze to prior only when prior == "stable", otherwise
// leave "(known after apply)" so the apply re-resolves it without an inconsistency error.
type volatileStateModifier struct{}

func (m volatileStateModifier) Description(_ context.Context) string {
	return "Reuses prior state only when it is the terminal 'stable' value; otherwise recomputes on apply."
}

func (m volatileStateModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m volatileStateModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if !req.PlanValue.IsUnknown() {
		return
	}
	if isStableState(req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

// Attr-type maps for the nested objects inside the info block.
var userInfoAttrTypes = map[string]attr.Type{"id": types.StringType, "name": types.StringType}
var lcrInfoAttrTypes = map[string]attr.Type{
	"state": types.StringType, "create_time": types.StringType,
	"error_message":   types.StringType,
	"created_by_user": types.ObjectType{AttrTypes: userInfoAttrTypes},
}
var pricingInfoAttrTypes = map[string]attr.Type{
	"month": types.Float64Type, "day": types.Float64Type, "hour": types.Float64Type,
}

// simpleStateInfoAttrTypes is the attr.Type map for the base ResourceInfo object.
var simpleStateInfoAttrTypes = map[string]attr.Type{
	"state":               types.StringType,
	"create_time":         types.StringType,
	"created_by_user":     types.ObjectType{AttrTypes: userInfoAttrTypes},
	"last_change_request": types.ObjectType{AttrTypes: lcrInfoAttrTypes},
	"pricing":             types.ObjectType{AttrTypes: pricingInfoAttrTypes},
}

func buildUserInfoObj(m map[string]interface{}) types.Object {
	if m == nil {
		return types.ObjectNull(userInfoAttrTypes)
	}
	obj, _ := types.ObjectValue(userInfoAttrTypes, map[string]attr.Value{
		"id": getString(m, "id"), "name": getString(m, "name"),
	})
	return obj
}

func buildLcrInfoObj(m map[string]interface{}) types.Object {
	if m == nil {
		return types.ObjectNull(lcrInfoAttrTypes)
	}
	var cbu map[string]interface{}
	if v, ok := m["createdByUser"].(map[string]interface{}); ok {
		cbu = v
	}
	obj, _ := types.ObjectValue(lcrInfoAttrTypes, map[string]attr.Value{
		"state":           getString(m, "state"),
		"create_time":     getString(m, "createTime"),
		"error_message":   getString(m, "errorMessage"),
		"created_by_user": buildUserInfoObj(cbu),
	})
	return obj
}

func buildPricingObj(m map[string]interface{}) types.Object {
	if m == nil {
		return types.ObjectNull(pricingInfoAttrTypes)
	}
	obj, _ := types.ObjectValue(pricingInfoAttrTypes, map[string]attr.Value{
		"month": getFloat64(m, "month"), "day": getFloat64(m, "day"), "hour": getFloat64(m, "hour"),
	})
	return obj
}

// buildInfoObj reads the full ResourceInfo from data["status"] and merges resource-specific extras.
// Simple resources call simpleStateInfoObj (no extras). Resources with extra status fields
// (e.g. VM's public_ipv4) pass them via extraAttrTypes / extraVals.
func buildInfoObj(data map[string]interface{}, extraAttrTypes map[string]attr.Type, extraVals map[string]attr.Value) types.Object {
	status, _ := data["status"].(map[string]interface{})
	if status == nil {
		status = map[string]interface{}{}
	}
	var cbu, lcr, pricing map[string]interface{}
	if v, ok := status["createdByUser"].(map[string]interface{}); ok {
		cbu = v
	}
	if v, ok := status["lastChangeRequest"].(map[string]interface{}); ok {
		lcr = v
	}
	if v, ok := status["pricing"].(map[string]interface{}); ok {
		pricing = v
	}

	attrTypes := map[string]attr.Type{}
	for k, v := range simpleStateInfoAttrTypes {
		attrTypes[k] = v
	}
	for k, v := range extraAttrTypes {
		attrTypes[k] = v
	}

	vals := map[string]attr.Value{
		"state":               getString(status, "state"),
		"create_time":         getString(status, "createTime"),
		"created_by_user":     buildUserInfoObj(cbu),
		"last_change_request": buildLcrInfoObj(lcr),
		"pricing":             buildPricingObj(pricing),
	}
	for k, v := range extraVals {
		vals[k] = v
	}

	obj, _ := types.ObjectValue(attrTypes, vals)
	return obj
}

// simpleStateInfoObj builds a types.Object with the full base ResourceInfo from data["status"].
func simpleStateInfoObj(data map[string]interface{}) types.Object {
	return buildInfoObj(data, nil, nil)
}

func baseInfoSchemaAttrs() map[string]schema.Attribute {
	userAttrs := map[string]schema.Attribute{
		"id":   schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{Computed: true},
	}
	return map[string]schema.Attribute{
		"state":           schema.StringAttribute{Computed: true},
		"create_time":     schema.StringAttribute{Computed: true},
		"created_by_user": schema.SingleNestedAttribute{Computed: true, Attributes: userAttrs},
		"last_change_request": schema.SingleNestedAttribute{Computed: true, Attributes: map[string]schema.Attribute{
			"state":           schema.StringAttribute{Computed: true},
			"create_time":     schema.StringAttribute{Computed: true},
			"error_message":   schema.StringAttribute{Computed: true},
			"created_by_user": schema.SingleNestedAttribute{Computed: true, Attributes: userAttrs},
		}},
		"pricing": schema.SingleNestedAttribute{Computed: true, Attributes: map[string]schema.Attribute{
			"month": schema.Float64Attribute{Computed: true},
			"day":   schema.Float64Attribute{Computed: true},
			"hour":  schema.Float64Attribute{Computed: true},
		}},
	}
}

// commonInfoSchema returns a Computed SingleNestedAttribute for the "info" block.
// The base ResourceInfo fields (state, create_time, created_by_user, last_change_request, pricing)
// are included automatically. Pass extra resource-specific attrs to merge in.
// Access in HCL: resource.foo.info.state, resource.foo.info.public_ipv4, etc.
func commonInfoSchema(extraAttrs map[string]schema.Attribute) schema.Attribute {
	attrs := baseInfoSchemaAttrs()
	for k, v := range extraAttrs {
		attrs[k] = v
	}
	return schema.SingleNestedAttribute{
		Computed:      true,
		Attributes:    attrs,
		PlanModifiers: []planmodifier.Object{volatileInfoModifier{}},
	}
}

// CommonPriorSchema builds the PriorSchema argument for CommonInfoUpgrader.
// extraAttrs: any non-common, non-info resource attributes (e.g. vm_state, tier).
// infoAttrs:  the info block attributes AS THEY WERE before the schema change.
//
// Example — prior schema for a folder (info had only "state"):
//
//	var folderPriorSchemaV0 = CommonPriorSchema(nil, map[string]schema.Attribute{
//	    "state": schema.StringAttribute{Computed: true},
//	})
func CommonPriorSchema(extraAttrs map[string]schema.Attribute, infoAttrs map[string]schema.Attribute) *schema.Schema {
	attrs := commonSchemaAttributes()
	for k, v := range extraAttrs {
		attrs[k] = v
	}
	attrs["info"] = commonInfoSchema(infoAttrs)
	return &schema.Schema{Attributes: attrs}
}

// CommonInfoUpgrader returns a StateUpgrader that migrates state when new Computed
// fields are added to the info SingleNestedAttribute.  It copies all existing info
// values and fills missing new fields with null so the API can populate them on the
// next read.
//
// Usage — when bumping a resource's schema Version from N to N+1 because new info
// fields were added:
//
//	var myResourcePriorSchemaVN = CommonPriorSchema(extraAttrs, oldInfoAttrs)
//
//	func (r *MyResource) StateUpgraders(ctx context.Context,
//	    req resource.StateUpgradersRequest, resp *resource.StateUpgradersResponse,
//	) {
//	    resp.StateUpgraders = []resource.StateUpgrader{
//	        CommonInfoUpgrader(&myResourcePriorSchemaVN, map[string]attr.Type{
//	            "state":     types.StringType,
//	            "new_field": types.StringType, // newly added field
//	        }),
//	    }
//	}
func CommonInfoUpgrader(priorSchema *schema.Schema, newInfoAttrTypes map[string]attr.Type) resource.StateUpgrader {
	return resource.StateUpgrader{
		PriorSchema: priorSchema,
		StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
			// Decode the prior state as a raw map of tftypes.Value.
			var stateVals map[string]tftypes.Value
			if err := req.State.Raw.As(&stateVals); err != nil {
				resp.Diagnostics.AddError("State Upgrade Error",
					fmt.Sprintf("failed to decode prior state: %s", err))
				return
			}

			// Extract the old info attribute values.
			oldInfoVals := make(map[string]tftypes.Value)
			if infoVal, ok := stateVals["info"]; ok && !infoVal.IsNull() && infoVal.IsKnown() {
				_ = infoVal.As(&oldInfoVals)
			}

			// Build the new info tftypes and fill in null for any missing new fields.
			newInfoTfTypes := make(map[string]tftypes.Type, len(newInfoAttrTypes))
			for name, at := range newInfoAttrTypes {
				newInfoTfTypes[name] = at.TerraformType(ctx)
			}
			for name, at := range newInfoAttrTypes {
				if _, exists := oldInfoVals[name]; !exists {
					oldInfoVals[name] = tftypes.NewValue(at.TerraformType(ctx), nil)
				}
			}

			stateVals["info"] = tftypes.NewValue(
				tftypes.Object{AttributeTypes: newInfoTfTypes}, oldInfoVals)

			// Rebuild the root tftypes.Object type, replacing only the info attribute type.
			oldRootType := req.State.Raw.Type().(tftypes.Object)
			newRootAttrTypes := make(map[string]tftypes.Type, len(oldRootType.AttributeTypes))
			for k, t := range oldRootType.AttributeTypes {
				newRootAttrTypes[k] = t
			}
			newRootAttrTypes["info"] = tftypes.Object{AttributeTypes: newInfoTfTypes}

			resp.State.Raw = tftypes.NewValue(
				tftypes.Object{AttributeTypes: newRootAttrTypes}, stateVals)
		},
	}
}

// CommonModel holds the common fields for all resources.
type CommonModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
}

// metadataModel is the shared nested "metadata" block model, mirroring the API's
// metadata envelope. Every resource and datasource embeds this as its `metadata` block.
// id is a computed mirror of the root-level id (which Terraform requires at the root for
// import/tooling); name/description/folder_id/delete_protection/labels are the user-settable
// identity fields.
type metadataModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
}

// metadataResourceSchema returns the Required "metadata" SingleNestedAttribute for resources.
// name is Required; the rest are Optional+Computed (server may default them); id is Computed.
func metadataResourceSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Required: true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"folder_id": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"delete_protection": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"labels": schema.MapAttribute{
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

// setCommonFieldsNested fills a metadataModel from an API response (delegates to setCommonFields,
// which reads the "metadata" sub-map and resolves absent folderId to null for root resources).
func setCommonFieldsNested(ctx context.Context, data map[string]interface{}, md *metadataModel) error {
	return setCommonFields(ctx, data, &md.ID, &md.Name, &md.Description, &md.FolderID, &md.DeleteProtection, &md.Labels)
}

// buildCommonRequestMap creates the PUT request body.
// The API uses k8s-style nesting: common fields (id, name, description, folderId,
// deleteProtection, labels) go under "metadata"; type-specific fields go under "spec".
// Callers must extract m["spec"].(map[string]interface{}) and add their fields to it.
func buildCommonRequestMap(id, name string, description types.String, folderID types.String, deleteProtection types.Bool, labels types.Map, ctx context.Context) map[string]interface{} {
	metadata := map[string]interface{}{"id": id, "name": name}
	if !description.IsNull() && !description.IsUnknown() {
		metadata["description"] = description.ValueString()
	}
	if !folderID.IsNull() && !folderID.IsUnknown() {
		metadata["folderId"] = folderID.ValueString()
	}
	if !deleteProtection.IsNull() && !deleteProtection.IsUnknown() {
		metadata["deleteProtection"] = deleteProtection.ValueBool()
	}
	if !labels.IsNull() && !labels.IsUnknown() {
		labelsMap := make(map[string]string)
		diags := labels.ElementsAs(ctx, &labelsMap, false)
		if !diags.HasError() {
			metadata["labels"] = labelsMap
		}
	}
	return map[string]interface{}{"metadata": metadata, "spec": map[string]interface{}{}}
}

// setCommonFields populates common model fields from an API response map.
// The API returns common fields nested under "metadata"; this function extracts that sub-map.
func setCommonFields(ctx context.Context, data map[string]interface{}, id *types.String, name *types.String, description *types.String, folderID *types.String, deleteProtection *types.Bool, labels *types.Map) error {
	md, _ := data["metadata"].(map[string]interface{})
	if md == nil {
		md = map[string]interface{}{}
	}
	if v, ok := md["id"].(string); ok {
		*id = types.StringValue(v)
	}
	if v, ok := md["name"].(string); ok {
		*name = types.StringValue(v)
	}
	if v, ok := md["description"].(string); ok {
		*description = types.StringValue(v)
	} else {
		*description = types.StringValue("")
	}
	if v, ok := md["folderId"].(string); ok && v != "" {
		*folderID = types.StringValue(v)
	} else {
		// folderId absent = root-level resource; resolve to null so Terraform doesn't
		// see an unknown value after apply (Optional+Computed requires a known result).
		*folderID = types.StringNull()
	}
	if v, ok := md["deleteProtection"].(bool); ok {
		*deleteProtection = types.BoolValue(v)
	} else {
		*deleteProtection = types.BoolValue(false)
	}

	labelsRaw, ok := md["labels"].(map[string]interface{})
	if ok && len(labelsRaw) > 0 {
		labelsMap := make(map[string]attr.Value, len(labelsRaw))
		for k, v := range labelsRaw {
			if sv, ok := v.(string); ok {
				labelsMap[k] = types.StringValue(sv)
			}
		}
		lm, diags := types.MapValue(types.StringType, labelsMap)
		if diags.HasError() {
			return fmt.Errorf("error building labels map")
		}
		*labels = lm
	} else {
		*labels = types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	return nil
}

// getSpec extracts the "spec" sub-map from an API response (type-specific fields).
func getSpec(data map[string]interface{}) map[string]interface{} {
	if spec, ok := data["spec"].(map[string]interface{}); ok {
		return spec
	}
	return map[string]interface{}{}
}

// infoFieldRaw looks up a field in the "status" sub-object case-insensitively.
//
// Why case-insensitive: the wire keys for status fields are inconsistently cased by the
// C# serializer — IP fields arrive all-lowercase ("publicipv4") while others are camelCase
// ("windowsAdministratorPassword", "endpointUrl"), and neither matches the swagger property
// casing the generator derives keys from. An exact-match lookup would silently read empty for
// the mismatched ones. A case-insensitive match is parity-preserving (any key an exact match
// already found is still found) and lets the generator emit one uniform key per field without
// a per-field wire-key override table. Why this over a table: one helper vs ~25 hand entries.
func infoFieldRaw(data map[string]interface{}, field string) (interface{}, bool) {
	info, ok := data["status"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	if v, ok := info[field]; ok {
		return v, true
	}
	lf := strings.ToLower(field)
	for k, v := range info {
		if strings.ToLower(k) == lf {
			return v, true
		}
	}
	return nil, false
}

// getStringFromInfo extracts a string field from the "status" sub-object of an API response.
func getStringFromInfo(data map[string]interface{}, field string) types.String {
	if v, ok := infoFieldRaw(data, field); ok {
		if s, ok := v.(string); ok {
			return types.StringValue(s)
		}
	}
	return types.StringValue("")
}

// getInt64FromInfo extracts an int64 field from the "status" sub-object.
func getInt64FromInfo(data map[string]interface{}, field string) types.Int64 {
	if v, ok := infoFieldRaw(data, field); ok {
		switch n := v.(type) {
		case float64:
			return types.Int64Value(int64(n))
		case int64:
			return types.Int64Value(n)
		case int:
			return types.Int64Value(int64(n))
		}
	}
	return types.Int64Value(0)
}

// getBoolFromInfo extracts a bool field from the "status" sub-object.
func getBoolFromInfo(data map[string]interface{}, field string) types.Bool {
	if v, ok := infoFieldRaw(data, field); ok {
		if b, ok := v.(bool); ok {
			return types.BoolValue(b)
		}
	}
	return types.BoolValue(false)
}

// getStringFromResource extracts a string field from the "resource" sub-object.
// Transaction sub-resources nest their ID and info under "resource": {"id":..., "info":{...}}.
func getStringFromResource(data map[string]interface{}, field string) types.String {
	res, ok := data["resource"].(map[string]interface{})
	if !ok {
		return types.StringNull()
	}
	return getString(res, field)
}

// getStringFromResourceInfo extracts a string field from the "resource"."info" sub-object.
func getStringFromResourceInfo(data map[string]interface{}, field string) types.String {
	res, ok := data["resource"].(map[string]interface{})
	if !ok {
		return types.StringValue("")
	}
	return getStringFromInfo(res, field)
}

// getString extracts a string from a flat API response map.
// Returns Null when the field is absent or null in the API response.
func getString(data map[string]interface{}, field string) types.String {
	v, exists := data[field]
	if !exists || v == nil {
		return types.StringNull()
	}
	if s, ok := v.(string); ok {
		return types.StringValue(s)
	}
	return types.StringNull()
}

// getBool extracts a bool from a flat API response map.
// Returns Null when the field is absent or null in the API response.
func getBool(data map[string]interface{}, field string) types.Bool {
	v, exists := data[field]
	if !exists || v == nil {
		return types.BoolNull()
	}
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}

// getInt64 extracts an int64 from a flat API response map.
// Returns Null when the field is absent or null in the API response.
func getInt64(data map[string]interface{}, field string) types.Int64 {
	v, exists := data[field]
	if !exists || v == nil {
		return types.Int64Null()
	}
	switch val := v.(type) {
	case float64:
		return types.Int64Value(int64(val))
	case int64:
		return types.Int64Value(val)
	case int:
		return types.Int64Value(int64(val))
	}
	return types.Int64Null()
}

// getFloat64 extracts a float64 from a flat API response map.
// Returns Null when the field is absent or null in the API response.
func getFloat64(data map[string]interface{}, field string) types.Float64 {
	v, exists := data[field]
	if !exists || v == nil {
		return types.Float64Null()
	}
	switch val := v.(type) {
	case float64:
		return types.Float64Value(val)
	case int64:
		return types.Float64Value(float64(val))
	case int:
		return types.Float64Value(float64(val))
	}
	return types.Float64Null()
}

// getStringList extracts a list of strings from a flat API response map.
func getStringList(ctx context.Context, data map[string]interface{}, field string) types.List {
	raw, ok := data[field].([]interface{})
	if !ok {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}
	vals := make([]attr.Value, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			vals = append(vals, types.StringValue(s))
		}
	}
	return types.ListValueMust(types.StringType, vals)
}

// getStringMap extracts a map[string]string from a flat API response map.
func getStringMap(data map[string]interface{}, field string) types.Map {
	raw, ok := data[field].(map[string]interface{})
	if !ok {
		return types.MapValueMust(types.StringType, map[string]attr.Value{})
	}
	vals := make(map[string]attr.Value, len(raw))
	for k, v := range raw {
		if s, ok := v.(string); ok {
			vals[k] = types.StringValue(s)
		}
	}
	return types.MapValueMust(types.StringType, vals)
}

// stringListToInterface converts types.List of strings to []interface{} for API requests.
func stringListToInterface(ctx context.Context, list types.List) []interface{} {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var strs []string
	list.ElementsAs(ctx, &strs, false)
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

// stringMapToInterface converts types.Map of strings to map[string]interface{} for API requests.
func stringMapToInterface(ctx context.Context, m types.Map) map[string]interface{} {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	var strs map[string]string
	m.ElementsAs(ctx, &strs, false)
	result := make(map[string]interface{}, len(strs))
	for k, v := range strs {
		result[k] = v
	}
	return result
}
