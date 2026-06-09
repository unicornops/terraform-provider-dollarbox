package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var _ datasource.DataSource = &NamespaceDataSource{}
var _ datasource.DataSourceWithConfigure = &NamespaceDataSource{}

func NewNamespaceDataSource() datasource.DataSource {
	return &NamespaceDataSource{}
}

type NamespaceDataSource struct {
	client *APIClient
}

func (d *NamespaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (d *NamespaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Fetches a DollarBox namespace by ID or slug.",
		Attributes:          namespaceLookupDataSourceAttributes(),
	}
}

func namespaceLookupDataSourceAttributes() map[string]datasourceschema.Attribute {
	attrs := namespaceDataSourceAttributes()
	attrs["id"] = datasourceschema.StringAttribute{
		MarkdownDescription: "DollarBox namespace ID.",
		Optional:            true,
		Computed:            true,
	}
	attrs["slug"] = datasourceschema.StringAttribute{
		MarkdownDescription: "Namespace slug. May be used instead of `id` for lookup.",
		Optional:            true,
		Computed:            true,
	}
	return attrs
}

func namespaceDataSourceAttributes() map[string]datasourceschema.Attribute {
	return map[string]datasourceschema.Attribute{
		"id": datasourceschema.StringAttribute{
			MarkdownDescription: "DollarBox namespace ID.",
			Computed:            true,
		},
		"slug": datasourceschema.StringAttribute{
			MarkdownDescription: "Namespace slug.",
			Computed:            true,
		},
		"allocated_containers": datasourceschema.Int64Attribute{
			MarkdownDescription: "Number of container slots allocated to this namespace.",
			Computed:            true,
		},
		"allocated_volume_gb": datasourceschema.Int64Attribute{
			MarkdownDescription: "Volume capacity in GB allocated to this namespace.",
			Computed:            true,
		},
		"status": datasourceschema.StringAttribute{
			MarkdownDescription: "Current namespace status.",
			Computed:            true,
		},
		"k8s_namespace": datasourceschema.StringAttribute{
			MarkdownDescription: "Backing Kubernetes namespace name.",
			Computed:            true,
		},
		"created_at": datasourceschema.StringAttribute{
			MarkdownDescription: "Creation timestamp.",
			Computed:            true,
		},
		"updated_at": datasourceschema.StringAttribute{
			MarkdownDescription: "Last update timestamp.",
			Computed:            true,
		},
	}
}

func (d *NamespaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *APIClient, got %T. Please report this to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *NamespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config namespaceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := ""
	if !config.ID.IsNull() && !config.ID.IsUnknown() {
		id = strings.TrimSpace(config.ID.ValueString())
	}
	slug := ""
	if !config.Slug.IsNull() && !config.Slug.IsUnknown() {
		slug = strings.TrimSpace(config.Slug.ValueString())
	}
	if id == "" && slug == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Missing Namespace Lookup",
			"Set id or slug to look up a DollarBox namespace.",
		)
		return
	}

	var namespace apiNamespace
	var err error
	if id != "" {
		namespace, err = d.client.GetNamespace(ctx, id)
		if err != nil {
			addAPIError(&resp.Diagnostics, "Read DollarBox Namespace", err)
			return
		}
		if slug != "" && namespace.Slug != slug {
			resp.Diagnostics.AddAttributeError(
				path.Root("slug"),
				"Namespace Lookup Mismatch",
				fmt.Sprintf("Namespace %s has slug %q, not %q.", id, namespace.Slug, slug),
			)
			return
		}
	} else {
		namespaces, listErr := d.client.ListNamespaces(ctx)
		if listErr != nil {
			addAPIError(&resp.Diagnostics, "List DollarBox Namespaces", listErr)
			return
		}
		var ok bool
		namespace, ok = findNamespaceBySlug(namespaces, slug)
		if !ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("slug"),
				"Namespace Not Found",
				fmt.Sprintf("No DollarBox namespace with slug %q was found in the selected organisation.", slug),
			)
			return
		}
	}

	state := namespaceModelFromAPI(namespace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
