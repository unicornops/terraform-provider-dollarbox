package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VolumeSnapshotsDataSource{}
var _ datasource.DataSourceWithConfigure = &VolumeSnapshotsDataSource{}

func NewVolumeSnapshotsDataSource() datasource.DataSource { return &VolumeSnapshotsDataSource{} }

type VolumeSnapshotsDataSource struct{ client *APIClient }

type volumeSnapshotsDataSourceModel struct {
	NamespaceID types.Int64               `tfsdk:"namespace_id"`
	PVCName     types.String              `tfsdk:"pvc_name"`
	Snapshots   []volumeSnapshotDataModel `tfsdk:"snapshots"`
}

func (d *VolumeSnapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_snapshots"
}

func (d *VolumeSnapshotsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists all manual and managed daily snapshots for a DollarBox volume, following API cursor pagination.",
		Attributes: map[string]datasourceschema.Attribute{
			"namespace_id": datasourceschema.Int64Attribute{MarkdownDescription: "DollarBox namespace ID containing the PVC.", Required: true},
			"pvc_name":     datasourceschema.StringAttribute{MarkdownDescription: "Source Kubernetes PVC name.", Required: true},
			"snapshots": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Snapshots for the selected PVC.",
				Computed:            true,
				NestedObject:        datasourceschema.NestedAttributeObject{Attributes: snapshotDataSourceAttributes(false)},
			},
		},
	}
}

func (d *VolumeSnapshotsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VolumeSnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeSnapshotsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	snapshots, err := d.client.ListVolumeSnapshots(ctx, strconv.FormatInt(config.NamespaceID.ValueInt64(), 10), config.PVCName.ValueString())
	if err != nil {
		addAPIError(&resp.Diagnostics, "List DollarBox Volume Snapshots", err)
		return
	}
	state := volumeSnapshotsDataSourceModel{NamespaceID: config.NamespaceID, PVCName: config.PVCName, Snapshots: make([]volumeSnapshotDataModel, 0, len(snapshots))}
	for _, snapshot := range snapshots {
		model, diags := volumeSnapshotDataModelFromAPI(ctx, config.NamespaceID, config.PVCName, snapshot)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Snapshots = append(state.Snapshots, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
