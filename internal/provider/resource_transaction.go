package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

type TransactionResourceModel struct {
	ID       types.String         `tfsdk:"id"`
	Metadata metadataModel        `tfsdk:"metadata"`
	Spec     TransactionSpecModel `tfsdk:"spec"`
	Status   types.Object         `tfsdk:"status"`
}

// TransactionSpecModel holds the transaction's spec: the delete-on-destroy flag plus one map
// of nested sub-resources per transactable type.
type TransactionSpecModel struct {
	DeleteResourcesOnTransactionDelete             types.Bool `tfsdk:"delete_resources_on_transaction_delete"`
	Folders                                        types.Map  `tfsdk:"folders"`
	SshKeys                                        types.Map  `tfsdk:"ssh_keys"`
	S3Buckets                                      types.Map  `tfsdk:"s3_buckets"`
	S3UserAccessPolicies                           types.Map  `tfsdk:"s3_user_access_policies"`
	S3Users                                        types.Map  `tfsdk:"s3_users"`
	Volumes                                        types.Map  `tfsdk:"volumes"`
	VolumeAttachments                              types.Map  `tfsdk:"volume_attachments"`
	AccessPolicies                                 types.Map  `tfsdk:"access_policies"`
	HostingProviders                               types.Map  `tfsdk:"hosting_providers"`
	SshPrivateKeys                                 types.Map  `tfsdk:"ssh_private_keys"`
	Certificates                                   types.Map  `tfsdk:"certificates"`
	VpcSubnets                                     types.Map  `tfsdk:"vpc_subnets"`
	VpcPeerings                                    types.Map  `tfsdk:"vpc_peerings"`
	VpcPeeringExternalPeers                        types.Map  `tfsdk:"vpc_peering_external_peers"`
	RouteTables                                    types.Map  `tfsdk:"route_tables"`
	RouteTableRoutes                               types.Map  `tfsdk:"route_table_routes"`
	RouteTableAttachments                          types.Map  `tfsdk:"route_table_attachments"`
	ImageSchedules                                 types.Map  `tfsdk:"image_schedules"`
	LoadbalancerTargetGroups                       types.Map  `tfsdk:"loadbalancer_target_groups"`
	LoadbalancerTargetGroupStaticTargets           types.Map  `tfsdk:"loadbalancer_target_group_static_targets"`
	LoadbalancerTargetGroupServiceDiscoveryTargets types.Map  `tfsdk:"loadbalancer_target_group_service_discovery_targets"`
	LoadbalancerHttpListeners                      types.Map  `tfsdk:"loadbalancer_http_listeners"`
	LoadbalancerHttpsListeners                     types.Map  `tfsdk:"loadbalancer_https_listeners"`
	LoadbalancerTlsListeners                       types.Map  `tfsdk:"loadbalancer_tls_listeners"`
	LoadbalancerTcpListeners                       types.Map  `tfsdk:"loadbalancer_tcp_listeners"`
	LoadbalancerUdpListeners                       types.Map  `tfsdk:"loadbalancer_udp_listeners"`
	LoadbalancerHttpListenerRules                  types.Map  `tfsdk:"loadbalancer_http_listener_rules"`
	LoadbalancerHttpsListenerRules                 types.Map  `tfsdk:"loadbalancer_https_listener_rules"`
	LoadbalancerTlsListenerRules                   types.Map  `tfsdk:"loadbalancer_tls_listener_rules"`
	LoadbalancerTcpListenerRules                   types.Map  `tfsdk:"loadbalancer_tcp_listener_rules"`
	LoadbalancerUdpListenerRules                   types.Map  `tfsdk:"loadbalancer_udp_listener_rules"`
	KubernetesNodeGroups                           types.Map  `tfsdk:"kubernetes_node_groups"`
	KubernetesUserRoles                            types.Map  `tfsdk:"kubernetes_user_roles"`
	OpenVpns                                       types.Map  `tfsdk:"open_vpns"`
	PostgresqlParametersSets                       types.Map  `tfsdk:"postgresql_parameters_sets"`
	SupportPlans                                   types.Map  `tfsdk:"support_plans"`
	SupportTickets                                 types.Map  `tfsdk:"support_tickets"`
	SupportTicketComments                          types.Map  `tfsdk:"support_ticket_comments"`
	GitlabRunners                                  types.Map  `tfsdk:"gitlab_runners"`
	OpenVpnUserSettings                            types.Map  `tfsdk:"open_vpn_user_settings"`
	Users                                          types.Map  `tfsdk:"users"`
	UserTokens                                     types.Map  `tfsdk:"user_tokens"`
	FloatingIps                                    types.Map  `tfsdk:"floating_ips"`
	Vpcs                                           types.Map  `tfsdk:"vpcs"`
	VpcPeeringPeers                                types.Map  `tfsdk:"vpc_peering_peers"`
	Loadbalancers                                  types.Map  `tfsdk:"loadbalancers"`
	Kubernetes                                     types.Map  `tfsdk:"kubernetes"`
	KubernetesUsers                                types.Map  `tfsdk:"kubernetes_users"`
	PostgresqlStandalones                          types.Map  `tfsdk:"postgresql_standalones"`
	OpenVpnUsers                                   types.Map  `tfsdk:"open_vpn_users"`
	BillingAccounts                                types.Map  `tfsdk:"billing_accounts"`
	Quotas                                         types.Map  `tfsdk:"quotas"`
	QuotaChangeRequests                            types.Map  `tfsdk:"quota_change_requests"`
	Gitlabs                                        types.Map  `tfsdk:"gitlabs"`
	Ollamas                                        types.Map  `tfsdk:"ollamas"`
	SupportTicketCommentAttachments                types.Map  `tfsdk:"support_ticket_comment_attachments"`
	Images                                         types.Map  `tfsdk:"images"`
	Vms                                            types.Map  `tfsdk:"vms"`
	SecurityGroups                                 types.Map  `tfsdk:"security_groups"`
}

// ---------------------------------------------------------------------------
// Attr types
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Resource
// ---------------------------------------------------------------------------

type TransactionResource struct{ client *client.Client }

func NewTransactionResource() resource.Resource { return &TransactionResource{} }

func (r *TransactionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transaction"
}

func (r *TransactionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: txnSchemaAttrs()}
}

func (r *TransactionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData))
		return
	}
	r.client = pd.Client
}

// ---------------------------------------------------------------------------
// Request body builder
// ---------------------------------------------------------------------------

func buildTxnBody(ctx context.Context, plan TransactionResourceModel, includeFolders, includeBuckets, includePolicies, includeUsers bool) map[string]interface{} {
	metadata := map[string]interface{}{
		"id":   plan.ID.ValueString(),
		"name": plan.Metadata.Name.ValueString(),
	}
	spec := map[string]interface{}{}
	m := map[string]interface{}{"metadata": metadata, "spec": spec}
	if !plan.Metadata.Description.IsNull() && !plan.Metadata.Description.IsUnknown() {
		metadata["description"] = plan.Metadata.Description.ValueString()
	}
	if !plan.Metadata.FolderID.IsNull() && !plan.Metadata.FolderID.IsUnknown() && plan.Metadata.FolderID.ValueString() != "" {
		metadata["folderId"] = plan.Metadata.FolderID.ValueString()
	}
	if !plan.Metadata.DeleteProtection.IsNull() && !plan.Metadata.DeleteProtection.IsUnknown() {
		metadata["deleteProtection"] = plan.Metadata.DeleteProtection.ValueBool()
	}
	if !plan.Metadata.Labels.IsNull() && !plan.Metadata.Labels.IsUnknown() {
		metadata["labels"] = strMapFromTF(ctx, plan.Metadata.Labels)
	}
	if !plan.Spec.DeleteResourcesOnTransactionDelete.IsNull() && !plan.Spec.DeleteResourcesOnTransactionDelete.IsUnknown() {
		spec["deleteResourcesOnTransactionDelete"] = plan.Spec.DeleteResourcesOnTransactionDelete.ValueBool()
	}
	txnBuildSubResources(ctx, &plan, spec, includeFolders, includeBuckets, includePolicies, includeUsers)
	return m
}

// idStrOf returns a non-empty string id from an attr.Value, or "".
func idStrOf(v attr.Value) string {
	if s, ok := v.(types.String); ok && !s.IsNull() && !s.IsUnknown() {
		return s.ValueString()
	}
	return ""
}

// assignSubIDsMap settles the id of every transaction sub-resource entry. Users set the id (when
// they need cross-referencing) on the entry's metadata.id; the root id is computed. This picks the
// effective id — metadata.id, else a root id, else a freshly minted ULID — and writes it to BOTH
// the root id (which buildXRequestMap reads to populate the request's metadata.id) and metadata.id
// (where the user set it and what the API echoes back), so the two always agree.
func assignSubIDsMap(m types.Map, attrTypes map[string]attr.Type) types.Map {
	if m.IsNull() || m.IsUnknown() {
		return m
	}
	elems := m.Elements()
	if len(elems) == 0 {
		return m
	}
	metaType, _ := attrTypes["metadata"].(types.ObjectType)
	updated := make(map[string]attr.Value, len(elems))
	for key, e := range elems {
		obj, ok := e.(types.Object)
		if !ok {
			updated[key] = e
			continue
		}
		attrs := obj.Attributes()

		var metaAttrs map[string]attr.Value
		if mo, ok := attrs["metadata"].(types.Object); ok && !mo.IsNull() && !mo.IsUnknown() {
			metaAttrs = mo.Attributes()
		}
		effective := ""
		if metaAttrs != nil {
			effective = idStrOf(metaAttrs["id"])
		}
		if effective == "" {
			effective = idStrOf(attrs["id"])
		}
		if effective == "" {
			effective = newULID()
		}

		newAttrs := make(map[string]attr.Value, len(attrs))
		for k, v := range attrs {
			newAttrs[k] = v
		}
		newAttrs["id"] = types.StringValue(effective)
		if metaAttrs != nil && metaType.AttrTypes != nil {
			nm := make(map[string]attr.Value, len(metaAttrs))
			for k, v := range metaAttrs {
				nm[k] = v
			}
			nm["id"] = types.StringValue(effective)
			if mo, d := types.ObjectValue(metaType.AttrTypes, nm); !d.HasError() {
				newAttrs["metadata"] = mo
			}
		}
		newObj, diags := types.ObjectValue(attrTypes, newAttrs)
		if diags.HasError() {
			updated[key] = e
			continue
		}
		updated[key] = newObj
	}
	result, _ := types.MapValue(types.ObjectType{AttrTypes: attrTypes}, updated)
	return result
}

func ulidOrNew(id types.String) string {
	if id.IsNull() || id.IsUnknown() || id.ValueString() == "" {
		return newULID()
	}
	return id.ValueString()
}

func strMapFromTF(ctx context.Context, m types.Map) map[string]string {
	var result map[string]string
	m.ElementsAs(ctx, &result, false)
	return result
}

// ---------------------------------------------------------------------------
// State population
// ---------------------------------------------------------------------------

func (r *TransactionResource) populateState(ctx context.Context, data map[string]interface{}, state *TransactionResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	state.Spec.DeleteResourcesOnTransactionDelete = getBool(getSpec(data), "deleteResourcesOnTransactionDelete")
	state.Status = simpleStateInfoObj(data)
	txnPopulateSubResources(ctx, data, state)
	return nil
}

// findMapKeyByID returns the map key whose object has the given "id" value, or ("", false).
func findMapKeyByID(m types.Map, id string) (string, bool) {
	if m.IsNull() || m.IsUnknown() || id == "" {
		return "", false
	}
	for k, v := range m.Elements() {
		obj, ok := v.(types.Object)
		if !ok {
			continue
		}
		if idAttr, ok2 := obj.Attributes()["id"].(types.String); ok2 && idAttr.ValueString() == id {
			return k, true
		}
	}
	return "", false
}

// resolveMapKey returns the stable map key for a sub-resource.
// Priority: (1) existing state key matched by ID — preserves user-chosen key names across updates;
// (2) name from API response — handles new sub-resources never saved to state (e.g. after Ctrl+C);
// (3) raw ID as last resort.
func resolveMapKey(existing types.Map, id string, item map[string]interface{}) string {
	if k, found := findMapKeyByID(existing, id); found {
		return k
	}
	if name, _ := item["name"].(string); name != "" {
		return name
	}
	return id
}

// --- Original 5 populate functions (unchanged) ---

// --- New 54 populate functions ---

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

func (r *TransactionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TransactionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())

	// The framework marks Computed+Optional nested attrs as unknown during Create (no prior
	// state), even when the user explicitly set them in config. Read req.Config to retrieve
	// user-provided values for sub-resource IDs (and bucket_id/access_policy_ids for s3_users).
	var cfg TransactionResourceModel
	req.Config.Get(ctx, &cfg) // diagnostics intentionally ignored — we only need known values
	txnAssignAndRecoverIDs(ctx, &plan, &cfg)

	hasUsers := !plan.Spec.S3Users.IsNull() && !plan.Spec.S3Users.IsUnknown() && len(plan.Spec.S3Users.Elements()) > 0

	// Phase 1: create transaction with folders + buckets + policies (no s3Users).
	// s3Users require bucketId as an API-assigned ULID which isn't known until phase 1 completes.
	//
	// Retry loop: if 422 ResourceIsScheduling is returned, a sub-resource from a previous
	// failed attempt is still being provisioned. Wait for it to stabilize, then retry.
	body := buildTxnBody(ctx, plan, true, true, true, false)
	var modResp *client.ModificationResponse
	var lastPutErr error
	for attempt := 0; attempt < 50; attempt++ {
		var putErr error
		modResp, putErr = r.client.Put(ctx, "/api/v1/transaction", body)
		if putErr == nil {
			break
		}
		lastPutErr = putErr
		blockingID, blockingType := parseSchedulingConflict(putErr.Error())
		if blockingID == "" {
			resp.Diagnostics.AddError("Create Error", putErr.Error())
			return
		}
		// Wait for the conflicting resource to leave Scheduling/Reconciling state.
		if waitErr := waitForResourceStable(ctx, r.client, blockingType, blockingID); waitErr != nil {
			resp.Diagnostics.AddError("Create Error (scheduling wait)", fmt.Sprintf("waiting for %s/%s: %v", blockingType, blockingID, waitErr))
			return
		}
		// The API has a dual-type quirk: a single ULID may exist as both an s3-bucket and
		// an s3-user-access-policy. Delete from ALL resource types so both are cleared.
		// Also delete regardless of state (not just schedulingfailed): a "stable" orphan from
		// a prior failed run still blocks Phase 1 internally in the API scheduler.
		for _, rt := range []string{"s3-bucket", "s3-user-access-policy", "folder", "transaction"} {
			deleteAnyExistingResource(ctx, r.client, rt, blockingID)
		}
		// Delay to let API-internal state propagate before the next PUT.
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Create Error", "context cancelled during retry")
			return
		case <-time.After(10 * time.Second):
		}
	}
	if modResp == nil {
		errDetail := "phase 1 failed after retries"
		if lastPutErr != nil {
			errDetail = fmt.Sprintf("phase 1 failed after retries; last error: %s", lastPutErr.Error())
		}
		resp.Diagnostics.AddError("Create Error", errDetail)
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/transaction", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	txnID := modResp.ResourceId
	if txnID == "" {
		txnID = plan.ID.ValueString()
	}

	if !hasUsers {
		// No users — single phase is sufficient.
		apiData, err := r.client.Get(ctx, "/api/v1/transaction", txnID)
		if err != nil {
			resp.Diagnostics.AddError("Read After Create Error", err.Error())
			return
		}
		if err := r.populateState(ctx, apiData, &plan); err != nil {
			resp.Diagnostics.AddError("State Population Error", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		return
	}

	// Wait for all phase-1 sub-resources (buckets, policies) to leave Scheduling state.
	// The transaction request completes before sub-resources finish provisioning; phase 2
	// PUT will fail with 422 if any sub-resource is still Scheduling.
	if err := waitForSubResourcesStable(ctx, r.client, txnID); err != nil {
		cleanupTransaction(ctx, r.client, txnID)
		resp.Diagnostics.AddError("Phase1 Sub-Resource Wait Error", err.Error())
		return
	}

	// Phase 2: read back ULIDs of created buckets/policies and wire them into s3Users.
	apiData, err := r.client.Get(ctx, "/api/v1/transaction", txnID)
	if err != nil {
		resp.Diagnostics.AddError("Read After Phase1 Error", err.Error())
		return
	}

	// populateState stores API-assigned bucket/policy ULIDs in plan.S3Buckets/S3UserAccessPolicies.
	// For s3Users: phase 1 response has no users yet, so populateTxnS3Users returns plan.S3Users unchanged.
	if err := r.populateState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error (phase1)", err.Error())
		return
	}

	// Wire s3_users to use the actual Phase 1 bucket/policy IDs. Users may have unknown/null
	// bucket_id/access_policy_ids when IDs are not pre-specified in config.
	plan.Spec.S3Users = autoWireUsers(ctx, plan.Spec.S3Users, plan.Spec.S3Buckets, plan.Spec.S3UserAccessPolicies)

	// Phase 2 PUT: include all sub-resources with wired-up user bucket/policy references.
	body2 := buildTxnBody(ctx, plan, true, true, true, true)
	modResp2, err := r.client.Put(ctx, "/api/v1/transaction", body2)
	if err != nil {
		cleanupTransaction(ctx, r.client, txnID)
		resp.Diagnostics.AddError("Create Phase2 Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/transaction", modResp2.RequestId); err != nil {
		cleanupTransaction(ctx, r.client, txnID)
		resp.Diagnostics.AddError("Create Phase2 Poll Error", err.Error())
		return
	}

	apiData2, err := r.client.Get(ctx, "/api/v1/transaction", txnID)
	if err != nil {
		resp.Diagnostics.AddError("Read After Phase2 Error", err.Error())
		return
	}
	if err := r.populateState(ctx, apiData2, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error (phase2)", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TransactionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TransactionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/transaction", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := r.populateState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *TransactionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TransactionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state TransactionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	var cfgU TransactionResourceModel
	req.Config.Get(ctx, &cfgU)
	txnAssignAndRecoverIDs(ctx, &plan, &cfgU)

	body := buildTxnBody(ctx, plan, true, true, true, true)
	modResp, err := r.client.Put(ctx, "/api/v1/transaction", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/transaction", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/transaction", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if err := r.populateState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TransactionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TransactionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/transaction", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/transaction", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *TransactionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state TransactionResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/transaction", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := r.populateState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ---------------------------------------------------------------------------
// Helpers for two-phase create
// ---------------------------------------------------------------------------

// parseSchedulingConflict extracts the ULID and resource type from a ResourceIsScheduling or
// ResourceIsReconciling 422 error so the caller can wait and retry.
func parseSchedulingConflict(errMsg string) (ulid, resType string) {
	if !strings.Contains(errMsg, "ResourceIsScheduling") && !strings.Contains(errMsg, "is in state Scheduling") &&
		!strings.Contains(errMsg, "Reconciling state") && !strings.Contains(errMsg, "ResourceIsReconciling") {
		return "", ""
	}
	// Map API type names to path segments
	typeMap := map[string]string{
		"OrganizationS3UserAccessPolicy": "s3-user-access-policy",
		"OrganizationS3Bucket":           "s3-bucket",
		"OrganizationFolder":             "folder",
		"OrganizationTransaction":        "transaction",
	}
	for apiType, path := range typeMap {
		if strings.Contains(errMsg, apiType) {
			resType = path
			break
		}
	}
	// Extract ULID: text inside parentheses after the type name.
	start := strings.Index(errMsg, " (")
	end := strings.Index(errMsg, ") is in state Scheduling")
	if end < 0 {
		end = strings.Index(errMsg, ") is in state Reconciling")
	}
	if start < 0 || end < 0 || end <= start {
		return "", ""
	}
	ulid = strings.TrimSpace(errMsg[start+2 : end])
	return ulid, resType
}

// waitForResourceStable polls a specific resource until its info.state leaves "scheduling".
func waitForResourceStable(ctx context.Context, c *client.Client, resType, id string) error {
	deadline := time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		apiData, err := c.Get(ctx, "/api/v1/"+resType, id)
		if err != nil || apiData == nil {
			return nil
		}
		info, _ := apiData["info"].(map[string]interface{})
		state, _ := info["state"].(string)
		s := strings.ToLower(state)
		if s != "scheduling" && s != "reconciling" {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
	return fmt.Errorf("timed out waiting for %s/%s to leave scheduling state", resType, id)
}

// deleteAnyExistingResource deletes a resource regardless of its state, silently ignoring
// "not found" responses. Used in the Phase 1 retry loop to clear orphaned blocking resources.
func deleteAnyExistingResource(ctx context.Context, c *client.Client, resType, id string) {
	apiData, err := c.Get(ctx, "/api/v1/"+resType, id)
	if err != nil || apiData == nil {
		return
	}
	_ = waitForResourceStable(ctx, c, resType, id)
	for i := 0; i < 10; i++ {
		delResp, delErr := c.Delete(ctx, "/api/v1/"+resType, id)
		if delErr == nil && delResp != nil {
			_ = c.PollUntilDone(ctx, "/api/v1/"+resType, delResp.RequestId)
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
		}
	}
}

// cleanupTransaction explicitly deletes all sub-resources of a transaction before deleting
// the transaction container.
func cleanupTransaction(ctx context.Context, c *client.Client, txnID string) {
	apiData, err := c.Get(ctx, "/api/v1/transaction", txnID)
	if err == nil && apiData != nil {
		for _, key := range []string{"s3Buckets", "s3UserAccessPolicies", "folders"} {
			resType := map[string]string{
				"s3Buckets": "s3-bucket", "s3UserAccessPolicies": "s3-user-access-policy", "folders": "folder",
			}[key]
			items, _ := apiData[key].([]interface{})
			for _, raw := range items {
				m, _ := raw.(map[string]interface{})
				if m == nil {
					continue
				}
				id, _ := m["id"].(string)
				if id != "" {
					for _, rt := range []string{resType, "s3-bucket", "s3-user-access-policy"} {
						deleteAnyExistingResource(ctx, c, rt, id)
					}
				}
			}
		}
	}
	if delResp, delErr := c.Delete(ctx, "/api/v1/transaction", txnID); delErr == nil {
		_ = c.PollUntilDone(ctx, "/api/v1/transaction", delResp.RequestId)
	}
}

// deleteIfSchedulingFailed deletes a resource that is in schedulingfailed state.
func deleteIfSchedulingFailed(ctx context.Context, c *client.Client, resType, id string) {
	apiData, err := c.Get(ctx, "/api/v1/"+resType, id)
	if err != nil || apiData == nil {
		return
	}
	info, _ := apiData["info"].(map[string]interface{})
	state, _ := info["state"].(string)
	if strings.ToLower(state) != "schedulingfailed" {
		return
	}
	for i := 0; i < 10; i++ {
		delResp, delErr := c.Delete(ctx, "/api/v1/"+resType, id)
		if delErr == nil && delResp != nil {
			_ = c.PollUntilDone(ctx, "/api/v1/"+resType, delResp.RequestId)
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
		}
	}
}

// waitForSubResourcesStable polls all bucket and policy sub-resources of the given transaction
// until none of them is in "scheduling" state.
func waitForSubResourcesStable(ctx context.Context, c *client.Client, txnID string) error {
	deadline := time.Now().Add(30 * time.Minute)
	type subRes struct{ resType, id string }
	for time.Now().Before(deadline) {
		apiData, err := c.Get(ctx, "/api/v1/transaction", txnID)
		if err != nil {
			return err
		}
		var pending []subRes
		for _, key := range []string{"s3Buckets", "s3UserAccessPolicies", "folders"} {
			resType := map[string]string{
				"s3Buckets": "s3-bucket", "s3UserAccessPolicies": "s3-user-access-policy", "folders": "folder",
			}[key]
			items, _ := apiData[key].([]interface{})
			for _, raw := range items {
				m, _ := raw.(map[string]interface{})
				if m == nil {
					continue
				}
				id, _ := m["id"].(string)
				info, _ := m["info"].(map[string]interface{})
				state, _ := info["state"].(string)
				switch strings.ToLower(state) {
				case "scheduling", "reconciling":
					pending = append(pending, subRes{resType, id})
				case "deleted", "schedulingfailed":
					return fmt.Errorf("sub-resource %s/%s entered failed state %q — provisioning failed", resType, id, state)
				}
			}
		}
		if len(pending) == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
	return fmt.Errorf("timed out waiting for transaction sub-resources to leave scheduling state")
}

// fillUnknownIDsFromConfig copies known values from cfgMap into planMap at matching keys
// for the "id" field and any extraFields.
func fillUnknownIDsFromConfig(_ context.Context, planMap, cfgMap types.Map, attrTypes map[string]attr.Type) types.Map {
	if cfgMap.IsNull() || cfgMap.IsUnknown() || planMap.IsNull() || planMap.IsUnknown() {
		return planMap
	}
	planElems := planMap.Elements()
	cfgElems := cfgMap.Elements()
	if len(planElems) == 0 {
		return planMap
	}
	specType, _ := attrTypes["spec"].(types.ObjectType)
	updated := make(map[string]attr.Value, len(planElems))
	for key, pe := range planElems {
		planObj, ok := pe.(types.Object)
		if !ok {
			updated[key] = pe
			continue
		}
		cfgObj, ok2 := cfgElems[key].(types.Object)
		if !ok2 {
			updated[key] = pe
			continue
		}
		newAttrs := make(map[string]attr.Value, len(planObj.Attributes()))
		for k, v := range planObj.Attributes() {
			newAttrs[k] = v
		}
		// Root id (mirrors metadata.id; metadata also recovered below).
		overrideIfConfigKnown(newAttrs, cfgObj.Attributes(), "id")
		// metadata.id, and spec ref fields, recovered from the nested config objects.
		planMeta, _ := newAttrs["metadata"].(types.Object)
		cfgMeta, _ := cfgObj.Attributes()["metadata"].(types.Object)
		if !planMeta.IsNull() && !cfgMeta.IsNull() {
			metaType, _ := attrTypes["metadata"].(types.ObjectType)
			ma := make(map[string]attr.Value, len(planMeta.Attributes()))
			for k, v := range planMeta.Attributes() {
				ma[k] = v
			}
			overrideIfConfigKnown(ma, cfgMeta.Attributes(), "id")
			if mo, d := types.ObjectValue(metaType.AttrTypes, ma); !d.HasError() {
				newAttrs["metadata"] = mo
			}
		}
		// Recover EVERY user-set spec field from config. The framework marks Optional+Computed
		// nested attrs unknown during create, so any cross-reference the user put on a spec field
		// (vpc_id, bucket_id, loadbalancer_id, access_policy_ids, ...) referencing another
		// sub-resource in the same transaction would otherwise be dropped. overrideIfConfigKnown
		// only copies known config values, so genuinely-computed fields stay unknown.
		planSpec, _ := newAttrs["spec"].(types.Object)
		cfgSpec, _ := cfgObj.Attributes()["spec"].(types.Object)
		if !planSpec.IsNull() && !cfgSpec.IsNull() {
			sa := make(map[string]attr.Value, len(planSpec.Attributes()))
			for k, v := range planSpec.Attributes() {
				sa[k] = v
			}
			for k := range sa {
				overrideIfConfigKnown(sa, cfgSpec.Attributes(), k)
			}
			if so, d := types.ObjectValue(specType.AttrTypes, sa); !d.HasError() {
				newAttrs["spec"] = so
			}
		}
		newObj, diags := types.ObjectValue(attrTypes, newAttrs)
		if diags.HasError() {
			updated[key] = pe
			continue
		}
		updated[key] = newObj
	}
	result, _ := types.MapValue(types.ObjectType{AttrTypes: attrTypes}, updated)
	return result
}

// overrideIfConfigKnown copies cfgAttrs[key] into planAttrs[key] whenever the config has a
// known non-null non-empty value.
func overrideIfConfigKnown(planAttrs, cfgAttrs map[string]attr.Value, key string) {
	cfgVal, hasCfg := cfgAttrs[key]
	if !hasCfg {
		return
	}
	if _, hasPlan := planAttrs[key]; !hasPlan {
		return
	}

	var cfgIsKnown bool
	switch cv := cfgVal.(type) {
	case types.String:
		cfgIsKnown = !cv.IsNull() && !cv.IsUnknown() && cv.ValueString() != ""
	case types.List:
		cfgIsKnown = !cv.IsNull() && !cv.IsUnknown()
	default:
		return
	}
	if !cfgIsKnown {
		return
	}

	// id-like fields hold ULIDs, which the API stores lowercase — normalize so plan == state.
	// Other fields (region, names, policy_json, ...) are copied verbatim.
	lower := isIdKey(key)
	switch cv := cfgVal.(type) {
	case types.String:
		s := cv.ValueString()
		if lower {
			s = strings.ToLower(s)
		}
		planAttrs[key] = types.StringValue(s)
	case types.List:
		if !lower {
			planAttrs[key] = cfgVal
			return
		}
		elems := cv.Elements()
		lowerElems := make([]attr.Value, len(elems))
		for i, e := range elems {
			if sv, ok := e.(types.String); ok {
				lowerElems[i] = types.StringValue(strings.ToLower(sv.ValueString()))
			} else {
				lowerElems[i] = e
			}
		}
		lowerList, diags := types.ListValue(types.StringType, lowerElems)
		if diags.HasError() {
			planAttrs[key] = cfgVal
			return
		}
		planAttrs[key] = lowerList
	default:
		planAttrs[key] = cfgVal
	}
}

// isIdKey reports whether an attribute holds a ULID reference (and so should be lowercase-normalized).
func isIdKey(k string) bool {
	return k == "id" || strings.HasSuffix(k, "_id") || strings.HasSuffix(k, "_ids")
}

// autoWireUsers populates s3_users.bucket_id with actual Phase 1 bucket ID for users that
// have unknown/null bucket_id.
func autoWireUsers(_ context.Context, users, buckets, _ types.Map) types.Map {
	if users.IsNull() || users.IsUnknown() || len(users.Elements()) == 0 {
		return users
	}
	var firstBucketID types.String
	if !buckets.IsNull() && !buckets.IsUnknown() {
		for _, bVal := range buckets.Elements() {
			if bObj, ok := bVal.(types.Object); ok {
				if idVal, ok2 := bObj.Attributes()["id"].(types.String); ok2 && !idVal.IsNull() && !idVal.IsUnknown() && idVal.ValueString() != "" {
					firstBucketID = idVal
					break
				}
			}
		}
	}
	if firstBucketID.IsNull() || firstBucketID.IsUnknown() {
		return users
	}
	elemTypes := txnObjType(S3UserResourceSchemaAttrs()).AttrTypes
	userElems := users.Elements()
	updated := make(map[string]attr.Value, len(userElems))
	for key, ue := range userElems {
		uObj, ok := ue.(types.Object)
		if !ok {
			updated[key] = ue
			continue
		}
		if bv, ok2 := getSpecString(uObj, "bucket_id"); !ok2 || bv.IsNull() || bv.IsUnknown() {
			uObj = setSpecField(uObj, elemTypes, "bucket_id", firstBucketID)
		}
		updated[key] = uObj
	}
	result, _ := types.MapValue(types.ObjectType{AttrTypes: elemTypes}, updated)
	return result
}
