package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type snapshotPolicyResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	NamespaceID      types.Int64    `tfsdk:"namespace_id"`
	PVCName          types.String   `tfsdk:"pvc_name"`
	RetentionDays    types.Int64    `tfsdk:"retention_days"`
	ProtectedGB      types.Int64    `tfsdk:"protected_gb"`
	BilledGB         types.Int64    `tfsdk:"billed_gb"`
	MonthlyCostCents types.Int64    `tfsdk:"monthly_cost_cents"`
	Status           types.String   `tfsdk:"status"`
	NextSnapshotAt   types.String   `tfsdk:"next_snapshot_at"`
	LastSnapshotAt   types.String   `tfsdk:"last_snapshot_at"`
	Error            types.String   `tfsdk:"error"`
	CreatedAt        types.String   `tfsdk:"created_at"`
	UpdatedAt        types.String   `tfsdk:"updated_at"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

type volumeSnapshotResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	NamespaceID      types.Int64    `tfsdk:"namespace_id"`
	PVCName          types.String   `tfsdk:"pvc_name"`
	Name             types.String   `tfsdk:"name"`
	Labels           types.Map      `tfsdk:"labels"`
	Kind             types.String   `tfsdk:"kind"`
	Status           types.String   `tfsdk:"status"`
	RestoreSizeBytes types.Int64    `tfsdk:"restore_size_bytes"`
	ReadyAt          types.String   `tfsdk:"ready_at"`
	Error            types.String   `tfsdk:"error"`
	CreatedAt        types.String   `tfsdk:"created_at"`
	UpdatedAt        types.String   `tfsdk:"updated_at"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

type volumeSnapshotDataModel struct {
	ID               types.String `tfsdk:"id"`
	NamespaceID      types.Int64  `tfsdk:"namespace_id"`
	PVCName          types.String `tfsdk:"pvc_name"`
	Name             types.String `tfsdk:"name"`
	Labels           types.Map    `tfsdk:"labels"`
	Kind             types.String `tfsdk:"kind"`
	Status           types.String `tfsdk:"status"`
	RestoreSizeBytes types.Int64  `tfsdk:"restore_size_bytes"`
	ReadyAt          types.String `tfsdk:"ready_at"`
	Error            types.String `tfsdk:"error"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

type snapshotRestoreResourceModel struct {
	ID            types.String   `tfsdk:"id"`
	NamespaceID   types.Int64    `tfsdk:"namespace_id"`
	SourcePVCName types.String   `tfsdk:"source_pvc_name"`
	SnapshotID    types.String   `tfsdk:"snapshot_id"`
	TargetPVCName types.String   `tfsdk:"target_pvc_name"`
	Status        types.String   `tfsdk:"status"`
	Error         types.String   `tfsdk:"error"`
	CreatedAt     types.String   `tfsdk:"created_at"`
	UpdatedAt     types.String   `tfsdk:"updated_at"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

func (m *snapshotPolicyResourceModel) updateFromAPI(policy apiSnapshotPolicy) {
	m.PVCName = types.StringValue(policy.PVCName)
	m.RetentionDays = types.Int64Value(policy.RetentionDays)
	m.ProtectedGB = types.Int64Value(policy.ProtectedGB)
	m.BilledGB = types.Int64Value(policy.BilledGB)
	m.MonthlyCostCents = types.Int64Value(policy.MonthlyCostCents)
	m.Status = types.StringValue(policy.Status)
	m.NextSnapshotAt = stringPointerValueOrNull(policy.NextSnapshotAt)
	m.LastSnapshotAt = stringPointerValueOrNull(policy.LastSnapshotAt)
	m.Error = stringValueOrNull(policy.Error)
	m.CreatedAt = stringValueOrNull(policy.CreatedAt)
	m.UpdatedAt = stringValueOrNull(policy.UpdatedAt)
}

func (m *volumeSnapshotResourceModel) updateFromAPI(ctx context.Context, snapshot apiVolumeSnapshot) diag.Diagnostics {
	m.ID = types.StringValue(snapshot.ID)
	m.Name = types.StringValue(snapshot.Name)
	m.Kind = types.StringValue(snapshot.Kind)
	m.Status = types.StringValue(snapshot.Status)
	m.RestoreSizeBytes = int64PointerValueOrNull(snapshot.RestoreSizeBytes)
	m.ReadyAt = stringPointerValueOrNull(snapshot.ReadyAt)
	m.Error = stringValueOrNull(snapshot.Error)
	m.CreatedAt = stringValueOrNull(snapshot.CreatedAt)
	m.UpdatedAt = stringValueOrNull(snapshot.UpdatedAt)
	if snapshot.Labels == nil {
		snapshot.Labels = map[string]string{}
	}
	labels, diags := types.MapValueFrom(ctx, types.StringType, snapshot.Labels)
	m.Labels = labels
	return diags
}

func volumeSnapshotDataModelFromAPI(
	ctx context.Context,
	namespaceID types.Int64,
	pvcName types.String,
	snapshot apiVolumeSnapshot,
) (volumeSnapshotDataModel, diag.Diagnostics) {
	resourceModel := volumeSnapshotResourceModel{NamespaceID: namespaceID, PVCName: pvcName}
	diags := resourceModel.updateFromAPI(ctx, snapshot)
	return volumeSnapshotDataModel{
		ID:               resourceModel.ID,
		NamespaceID:      namespaceID,
		PVCName:          pvcName,
		Name:             resourceModel.Name,
		Labels:           resourceModel.Labels,
		Kind:             resourceModel.Kind,
		Status:           resourceModel.Status,
		RestoreSizeBytes: resourceModel.RestoreSizeBytes,
		ReadyAt:          resourceModel.ReadyAt,
		Error:            resourceModel.Error,
		CreatedAt:        resourceModel.CreatedAt,
		UpdatedAt:        resourceModel.UpdatedAt,
	}, diags
}

func (m *snapshotRestoreResourceModel) updateFromAPI(restore apiSnapshotRestore) {
	m.ID = types.StringValue(restore.ID)
	m.SnapshotID = types.StringValue(restore.SnapshotID)
	m.TargetPVCName = types.StringValue(restore.PVCName)
	m.Status = types.StringValue(restore.Status)
	m.Error = stringValueOrNull(restore.Error)
	m.CreatedAt = stringValueOrNull(restore.CreatedAt)
	m.UpdatedAt = stringValueOrNull(restore.UpdatedAt)
}

func stringPointerValueOrNull(value *string) types.String {
	if value == nil || *value == "" {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func int64PointerValueOrNull(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}
