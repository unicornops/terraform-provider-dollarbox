package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VolumeSnapshotDataSource{}
var _ datasource.DataSourceWithConfigure = &VolumeSnapshotDataSource{}

func NewVolumeSnapshotDataSource() datasource.DataSource { return &VolumeSnapshotDataSource{} }

type VolumeSnapshotDataSource struct{ client *APIClient }

func (d *VolumeSnapshotDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_snapshot"
}

func (d *VolumeSnapshotDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Fetches a manual or managed daily DollarBox volume snapshot.",
		Attributes:          snapshotDataSourceAttributes(true),
	}
}

func snapshotDataSourceAttributes(lookup bool) map[string]datasourceschema.Attribute {
	attrs := map[string]datasourceschema.Attribute{
		"namespace_id":       datasourceschema.Int64Attribute{MarkdownDescription: "DollarBox namespace ID containing the PVC.", Computed: true},
		"pvc_name":           datasourceschema.StringAttribute{MarkdownDescription: "Source Kubernetes PVC name.", Computed: true},
		"id":                 datasourceschema.StringAttribute{MarkdownDescription: "Snapshot UUID.", Computed: true},
		"name":               datasourceschema.StringAttribute{MarkdownDescription: "Snapshot display name.", Computed: true},
		"labels":             datasourceschema.MapAttribute{MarkdownDescription: "Snapshot labels.", ElementType: types.StringType, Computed: true},
		"kind":               datasourceschema.StringAttribute{MarkdownDescription: "Snapshot kind, such as manual or scheduled.", Computed: true},
		"status":             datasourceschema.StringAttribute{MarkdownDescription: "Current snapshot status.", Computed: true},
		"restore_size_bytes": datasourceschema.Int64Attribute{MarkdownDescription: "Required restore size in bytes, when reported by CSI.", Computed: true},
		"ready_at":           datasourceschema.StringAttribute{MarkdownDescription: "Time the snapshot became ready, when available.", Computed: true},
		"error":              datasourceschema.StringAttribute{MarkdownDescription: "Latest snapshot error, when present.", Computed: true},
		"created_at":         datasourceschema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
		"updated_at":         datasourceschema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},
	}
	if lookup {
		attrs["namespace_id"] = datasourceschema.Int64Attribute{MarkdownDescription: "DollarBox namespace ID containing the PVC.", Required: true}
		attrs["pvc_name"] = datasourceschema.StringAttribute{MarkdownDescription: "Source Kubernetes PVC name.", Required: true}
		attrs["id"] = datasourceschema.StringAttribute{MarkdownDescription: "Snapshot UUID.", Required: true}
	}
	return attrs
}

func (d *VolumeSnapshotDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *APIClient, got %T. Please report this to the provider developers.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *VolumeSnapshotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeSnapshotDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	snapshot, err := d.client.GetVolumeSnapshot(ctx, strconv.FormatInt(config.NamespaceID.ValueInt64(), 10), config.PVCName.ValueString(), config.ID.ValueString())
	if err != nil {
		addAPIError(&resp.Diagnostics, "Read DollarBox Volume Snapshot", err)
		return
	}
	state, diags := volumeSnapshotDataModelFromAPI(ctx, config.NamespaceID, config.PVCName, snapshot)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
