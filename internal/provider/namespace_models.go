package provider

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type namespaceModel struct {
	ID                  types.String `tfsdk:"id"`
	Slug                types.String `tfsdk:"slug"`
	AllocatedContainers types.Int64  `tfsdk:"allocated_containers"`
	AllocatedVolumeGB   types.Int64  `tfsdk:"allocated_volume_gb"`
	Status              types.String `tfsdk:"status"`
	K8sNamespace        types.String `tfsdk:"k8s_namespace"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func namespaceModelFromAPI(namespace apiNamespace) namespaceModel {
	return namespaceModel{
		ID:                  types.StringValue(strconv.FormatInt(namespace.ID, 10)),
		Slug:                types.StringValue(namespace.Slug),
		AllocatedContainers: types.Int64Value(namespace.AllocatedContainers),
		AllocatedVolumeGB:   types.Int64Value(namespace.AllocatedVolumeGB),
		Status:              stringValueOrNull(namespace.Status),
		K8sNamespace:        stringValueOrNull(namespace.K8sNamespace),
		CreatedAt:           stringValueOrNull(namespace.CreatedAt),
		UpdatedAt:           stringValueOrNull(namespace.UpdatedAt),
	}
}

func updateNamespaceModelFromAPI(model *namespaceModel, namespace apiNamespace) {
	*model = namespaceModelFromAPI(namespace)
}

func findNamespaceBySlug(namespaces []apiNamespace, slug string) (apiNamespace, bool) {
	slug = strings.TrimSpace(slug)
	for _, namespace := range namespaces {
		if namespace.Slug == slug {
			return namespace, true
		}
	}
	return apiNamespace{}, false
}
