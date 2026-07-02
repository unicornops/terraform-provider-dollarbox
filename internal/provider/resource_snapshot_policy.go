package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SnapshotPolicyResource{}
var _ resource.ResourceWithConfigure = &SnapshotPolicyResource{}
var _ resource.ResourceWithImportState = &SnapshotPolicyResource{}

func NewSnapshotPolicyResource() resource.Resource { return &SnapshotPolicyResource{} }

type SnapshotPolicyResource struct{ client *APIClient }

func (r *SnapshotPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot_policy"
}

func (r *SnapshotPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "Enables daily snapshot protection for a DollarBox Longhorn PVC. Protection costs €0.10 per protected source GB per month. Billing is based on source PVC capacity, begins before activation completes, and decreases only after retained CSI snapshots are deleted.",
		Attributes: map[string]resourceschema.Attribute{
			"id":                 resourceschema.StringAttribute{MarkdownDescription: "Compound policy ID in `namespace_id/pvc_name` form.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"namespace_id":       resourceschema.Int64Attribute{MarkdownDescription: "DollarBox namespace ID containing the PVC.", Required: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"pvc_name":           resourceschema.StringAttribute{MarkdownDescription: "Kubernetes PVC name to protect.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"retention_days":     resourceschema.Int64Attribute{MarkdownDescription: "Number of daily snapshots to retain, from 1 through 7.", Optional: true, Computed: true, Default: int64default.StaticInt64(7), Validators: []validator.Int64{int64validator.Between(1, 7)}},
			"protected_gb":       resourceschema.Int64Attribute{MarkdownDescription: "Source PVC capacity protected in GB.", Computed: true},
			"billed_gb":          resourceschema.Int64Attribute{MarkdownDescription: "Protected source capacity currently billed in GB.", Computed: true},
			"monthly_cost_cents": resourceschema.Int64Attribute{MarkdownDescription: "Current monthly policy cost in euro cents.", Computed: true},
			"status":             resourceschema.StringAttribute{MarkdownDescription: "Current snapshot policy status.", Computed: true},
			"next_snapshot_at":   resourceschema.StringAttribute{MarkdownDescription: "Scheduled time of the next snapshot, when available.", Computed: true},
			"last_snapshot_at":   resourceschema.StringAttribute{MarkdownDescription: "Time of the most recent snapshot, when available.", Computed: true},
			"error":              resourceschema.StringAttribute{MarkdownDescription: "Latest policy error, when present.", Computed: true},
			"created_at":         resourceschema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
			"updated_at":         resourceschema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},
		},
		Blocks: map[string]resourceschema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func (r *SnapshotPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *APIClient, got %T. Please report this to the provider developers.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *SnapshotPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan snapshotPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, diags := plan.Timeouts.Create(ctx, snapshotOperationTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	operationCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	policy, err := r.client.PutSnapshotPolicy(operationCtx, strconv.FormatInt(plan.NamespaceID.ValueInt64(), 10), plan.PVCName.ValueString(), snapshotPolicyPayload{RetentionDays: plan.RetentionDays.ValueInt64()})
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Snapshot Policy", err)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%d/%s", plan.NamespaceID.ValueInt64(), plan.PVCName.ValueString()))
	plan.updateFromAPI(policy)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policy, err = waitForSnapshotPolicyActive(operationCtx, r.client, strconv.FormatInt(plan.NamespaceID.ValueInt64(), 10), plan.PVCName.ValueString(), policy)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Snapshot Policy", err)
		return
	}
	plan.updateFromAPI(policy)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SnapshotPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state snapshotPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policy, err := r.client.GetSnapshotPolicy(ctx, strconv.FormatInt(state.NamespaceID.ValueInt64(), 10), state.PVCName.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Read DollarBox Snapshot Policy", err)
		return
	}
	if policy.Status == "disabled" {
		resp.State.RemoveResource(ctx)
		return
	}
	state.updateFromAPI(policy)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SnapshotPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan snapshotPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, diags := plan.Timeouts.Update(ctx, snapshotOperationTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	operationCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	policy, err := r.client.PutSnapshotPolicy(operationCtx, strconv.FormatInt(plan.NamespaceID.ValueInt64(), 10), plan.PVCName.ValueString(), snapshotPolicyPayload{RetentionDays: plan.RetentionDays.ValueInt64()})
	if err == nil {
		policy, err = waitForSnapshotPolicyActive(operationCtx, r.client, strconv.FormatInt(plan.NamespaceID.ValueInt64(), 10), plan.PVCName.ValueString(), policy)
	}
	if err != nil {
		addAPIError(&resp.Diagnostics, "Update DollarBox Snapshot Policy", err)
		return
	}
	plan.updateFromAPI(policy)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SnapshotPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state snapshotPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, diags := state.Timeouts.Delete(ctx, snapshotOperationTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	operationCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	namespaceID := strconv.FormatInt(state.NamespaceID.ValueInt64(), 10)
	policy, err := r.client.DeleteSnapshotPolicy(operationCtx, namespaceID, state.PVCName.ValueString())
	if isNotFoundError(err) {
		return
	}
	if err == nil {
		err = waitForSnapshotPolicyDeleted(operationCtx, r.client, namespaceID, state.PVCName.ValueString(), policy)
	}
	if err != nil {
		addAPIError(&resp.Diagnostics, "Delete DollarBox Snapshot Policy", err)
	}
}

func (r *SnapshotPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseSnapshotImportID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Snapshot Policy Import ID", "Use {namespace_id}/{pvc_name}: "+err.Error())
		return
	}
	namespaceID, _ := strconv.ParseInt(parts[0], 10, 64)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace_id"), namespaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pvc_name"), parts[1])...)
}
