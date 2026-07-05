package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &VolumeSnapshotResource{}
var _ resource.ResourceWithConfigure = &VolumeSnapshotResource{}
var _ resource.ResourceWithImportState = &VolumeSnapshotResource{}

func NewVolumeSnapshotResource() resource.Resource { return &VolumeSnapshotResource{} }

type VolumeSnapshotResource struct{ client *APIClient }

func (r *VolumeSnapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_snapshot"
}

func (r *VolumeSnapshotResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "Creates the single manual snapshot permitted for a DollarBox Longhorn PVC. Managed daily policy snapshots should be referenced through the volume snapshot data sources.",
		Attributes: map[string]resourceschema.Attribute{
			"id":                 resourceschema.StringAttribute{MarkdownDescription: "Snapshot UUID.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"namespace_id":       resourceschema.Int64Attribute{MarkdownDescription: "DollarBox namespace ID containing the PVC.", Required: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"pvc_name":           resourceschema.StringAttribute{MarkdownDescription: "Source Kubernetes PVC name.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":               resourceschema.StringAttribute{MarkdownDescription: "Optional display name, up to 100 characters. Changing it replaces the snapshot.", Optional: true, Computed: true, Validators: []validator.String{stringvalidator.LengthAtMost(100)}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"labels":             resourceschema.MapAttribute{MarkdownDescription: "Optional string labels. Changing labels replaces the snapshot.", ElementType: types.StringType, Optional: true, Computed: true, PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()}},
			"kind":               resourceschema.StringAttribute{MarkdownDescription: "Snapshot kind, such as manual or scheduled.", Computed: true},
			"status":             resourceschema.StringAttribute{MarkdownDescription: "Current snapshot status.", Computed: true},
			"restore_size_bytes": resourceschema.Int64Attribute{MarkdownDescription: "Required restore size in bytes, when reported by CSI.", Computed: true},
			"ready_at":           resourceschema.StringAttribute{MarkdownDescription: "Time the snapshot became ready, when available.", Computed: true},
			"error":              resourceschema.StringAttribute{MarkdownDescription: "Latest snapshot error, when present.", Computed: true},
			"created_at":         resourceschema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
			"updated_at":         resourceschema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},
		},
		Blocks: map[string]resourceschema.Block{"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true, Delete: true})},
	}
}

func (r *VolumeSnapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VolumeSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeSnapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	labels, diags := expandStringMap(ctx, plan.Labels)
	resp.Diagnostics.Append(diags...)
	timeout, timeoutDiags := plan.Timeouts.Create(ctx, snapshotOperationTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	operationCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	namespaceID := strconv.FormatInt(plan.NamespaceID.ValueInt64(), 10)
	snapshot, err := r.client.CreateVolumeSnapshot(operationCtx, namespaceID, plan.PVCName.ValueString(), snapshotCreatePayload{Name: plan.Name.ValueString(), Labels: labels})
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Volume Snapshot", err)
		return
	}
	resp.Diagnostics.Append(plan.updateFromAPI(ctx, snapshot)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	snapshot, err = waitForVolumeSnapshotReady(operationCtx, r.client, namespaceID, plan.PVCName.ValueString(), snapshot)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Volume Snapshot", err)
		return
	}
	resp.Diagnostics.Append(plan.updateFromAPI(ctx, snapshot)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VolumeSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeSnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	snapshot, err := r.client.GetVolumeSnapshot(ctx, strconv.FormatInt(state.NamespaceID.ValueInt64(), 10), state.PVCName.ValueString(), state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Read DollarBox Volume Snapshot", err)
		return
	}
	if snapshot.Status == "deleted" {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(state.updateFromAPI(ctx, snapshot)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VolumeSnapshotResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unexpected DollarBox Volume Snapshot Update", "DollarBox volume snapshots are immutable. Terraform should plan replacement when snapshot inputs change.")
}

func (r *VolumeSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeSnapshotResourceModel
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
	snapshot, err := r.client.DeleteVolumeSnapshot(operationCtx, namespaceID, state.PVCName.ValueString(), state.ID.ValueString())
	if isNotFoundError(err) {
		return
	}
	if err == nil {
		err = waitForVolumeSnapshotDeleted(operationCtx, r.client, namespaceID, state.PVCName.ValueString(), snapshot)
	}
	if err != nil {
		addAPIError(&resp.Diagnostics, "Delete DollarBox Volume Snapshot", err)
	}
}

func (r *VolumeSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseSnapshotImportID(req.ID, 3)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Volume Snapshot Import ID", "Use {namespace_id}/{pvc_name}/{snapshot_id}: "+err.Error())
		return
	}
	namespaceID, _ := strconv.ParseInt(parts[0], 10, 64)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace_id"), namespaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pvc_name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
}
