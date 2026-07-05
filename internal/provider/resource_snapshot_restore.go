package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var _ resource.Resource = &SnapshotRestoreResource{}
var _ resource.ResourceWithConfigure = &SnapshotRestoreResource{}

func NewSnapshotRestoreResource() resource.Resource { return &SnapshotRestoreResource{} }

type SnapshotRestoreResource struct{ client *APIClient }

func (r *SnapshotRestoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot_restore"
}

func (r *SnapshotRestoreResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "Restores a DollarBox volume snapshot into a new PVC. The API enforces namespace storage allocation. Destroying this Terraform resource only forgets the completed restore operation; it does not delete the restored PVC.",
		Attributes: map[string]resourceschema.Attribute{
			"id":              resourceschema.StringAttribute{MarkdownDescription: "Snapshot restore UUID.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"namespace_id":    resourceschema.Int64Attribute{MarkdownDescription: "DollarBox namespace ID containing the source PVC.", Required: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"source_pvc_name": resourceschema.StringAttribute{MarkdownDescription: "PVC containing the source snapshot.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"snapshot_id":     resourceschema.StringAttribute{MarkdownDescription: "Snapshot UUID to restore.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"target_pvc_name": resourceschema.StringAttribute{MarkdownDescription: "Name of the new PVC created by the restore.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"status":          resourceschema.StringAttribute{MarkdownDescription: "Current restore status.", Computed: true},
			"error":           resourceschema.StringAttribute{MarkdownDescription: "Latest restore error, when present.", Computed: true},
			"created_at":      resourceschema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
			"updated_at":      resourceschema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},
		},
		Blocks: map[string]resourceschema.Block{"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true})},
	}
}

func (r *SnapshotRestoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SnapshotRestoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan snapshotRestoreResourceModel
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
	namespaceID := strconv.FormatInt(plan.NamespaceID.ValueInt64(), 10)
	restore, err := r.client.CreateSnapshotRestore(operationCtx, namespaceID, plan.SourcePVCName.ValueString(), plan.SnapshotID.ValueString(), snapshotRestorePayload{PVCName: plan.TargetPVCName.ValueString()})
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Snapshot Restore", err)
		return
	}
	plan.updateFromAPI(restore)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	restore, err = waitForSnapshotRestoreBound(operationCtx, r.client, namespaceID, plan.SourcePVCName.ValueString(), restore)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Snapshot Restore", err)
		return
	}
	plan.updateFromAPI(restore)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SnapshotRestoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state snapshotRestoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	restore, err := r.client.GetSnapshotRestore(ctx, strconv.FormatInt(state.NamespaceID.ValueInt64(), 10), state.SourcePVCName.ValueString(), state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.Diagnostics.AddWarning("Snapshot Restore Status Unavailable", "The completed restore operation is no longer returned by the API. Terraform retained it in state to avoid creating the target PVC again.")
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
		addAPIError(&resp.Diagnostics, "Read DollarBox Snapshot Restore", err)
		return
	}
	state.updateFromAPI(restore)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SnapshotRestoreResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unexpected DollarBox Snapshot Restore Update", "Snapshot restores are immutable operations. Terraform should plan replacement when an input changes.")
}

func (r *SnapshotRestoreResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// A restore creates a PVC but the snapshot restore API has no delete operation.
	// Forgetting the operation is intentionally non-destructive.
}
