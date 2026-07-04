package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func securityGroupSpecSchema(t *testing.T) map[string]schema.Attribute {
	t.Helper()
	r := NewSecurityGroupResource().(*SecurityGroupResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	spec, ok := resp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("spec is not a SingleNestedAttribute")
	}
	return spec.Attributes
}

func TestSecurityGroupSpec_HasDefaultActionField(t *testing.T) {
	spec := securityGroupSpecSchema(t)
	attr, ok := spec["default_action"]
	if !ok {
		t.Fatal("expected spec attribute \"default_action\" in security_group schema")
	}
	sa, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("default_action is not a StringAttribute")
	}
	if !sa.Optional || !sa.Computed {
		t.Errorf("default_action should be Optional+Computed (server defaults to \"deny\" when omitted), got Optional=%v Computed=%v", sa.Optional, sa.Computed)
	}
}

func baseSecurityGroupPlan() SecurityGroupResourceModel {
	return SecurityGroupResourceModel{
		ID:       types.StringValue("01sg"),
		Metadata: metadataModel{Name: types.StringValue("my-sg"), Description: types.StringNull(), FolderID: types.StringNull(), Labels: types.MapNull(types.StringType)},
		Spec: SecurityGroupSpecModel{
			DefaultAction: types.StringNull(),
			Egress:        types.ListNull(types.ObjectType{AttrTypes: objAttrTypes(securityGroupEgressObjFields)}),
			Ingress:       types.ListNull(types.ObjectType{AttrTypes: objAttrTypes(securityGroupIngressObjFields)}),
		},
	}
}

func TestBuildSecurityGroupRequestMap_DefaultActionNull(t *testing.T) {
	m := buildSecurityGroupRequestMap(context.Background(), baseSecurityGroupPlan())
	spec := m["spec"].(map[string]interface{})
	if _, ok := spec["defaultAction"]; ok {
		t.Error("defaultAction should be omitted when null")
	}
}

func TestBuildSecurityGroupRequestMap_DefaultActionAllow(t *testing.T) {
	plan := baseSecurityGroupPlan()
	plan.Spec.DefaultAction = types.StringValue("allow")
	m := buildSecurityGroupRequestMap(context.Background(), plan)
	spec := m["spec"].(map[string]interface{})
	if spec["defaultAction"] != "allow" {
		t.Errorf("expected spec.defaultAction=\"allow\", got %v", spec["defaultAction"])
	}
}

func TestPopulateSecurityGroupState_DefaultAction(t *testing.T) {
	data := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01sg", "name": "my-sg"},
		"spec":     map[string]interface{}{"defaultAction": "allow"},
	}
	var state SecurityGroupResourceModel
	if err := populateSecurityGroupState(context.Background(), data, &state); err != nil {
		t.Fatalf("populateSecurityGroupState error: %v", err)
	}
	if state.Spec.DefaultAction.ValueString() != "allow" {
		t.Errorf("spec.default_action: got %q, want \"allow\"", state.Spec.DefaultAction.ValueString())
	}
}
