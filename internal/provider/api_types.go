package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const apiBasePath = "/api/v1"

type apiContainer struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Env         map[string]string `json:"env"`
	Command     []string          `json:"command"`
	Status      string            `json:"status"`
	IPv6Address string            `json:"ipv6_address"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

type containerPayload struct {
	Name    string            `json:"name,omitempty"`
	Image   string            `json:"image"`
	Env     map[string]string `json:"env"`
	Command []string          `json:"command"`
}

type apiVolume struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	SizeGB       int64  `json:"size_gb"`
	Status       string `json:"status"`
	StorageClass string `json:"storage_class"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type volumePayload struct {
	Name   string `json:"name"`
	SizeGB int64  `json:"size_gb"`
}

type apiInvitation struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Accepted  bool   `json:"accepted"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

type invitationPayload struct {
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
}

type apiKubectlCredential struct {
	ID               int64   `json:"id"`
	Org              string  `json:"org"`
	SAName           string  `json:"sa_name"`
	CreatedAt        string  `json:"created_at"`
	RotatedAt        string  `json:"rotated_at"`
	LastDownloadedAt *string `json:"last_downloaded_at"`
	Kubeconfig       string  `json:"kubeconfig,omitempty"`
}

type apiOrg struct {
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	BillingEmail   string `json:"billing_email"`
	Status         string `json:"status"`
	BillingMode    string `json:"billing_mode"`
	KubectlEnabled bool   `json:"kubectl_enabled"`
	APIEnabled     bool   `json:"api_enabled"`
	CreatedAt      string `json:"created_at"`
}

type orgPayload struct {
	Name         string `json:"name"`
	BillingEmail string `json:"billing_email"`
}

type apiMember struct {
	ID       string
	Email    string
	Role     string
	JoinedAt string
}

type memberPayload struct {
	Role string `json:"role,omitempty"`
}

type apiNamespace struct {
	ID                  int64  `json:"id"`
	Slug                string `json:"slug"`
	AllocatedContainers int64  `json:"allocated_containers"`
	AllocatedVolumeGB   int64  `json:"allocated_volume_gb"`
	Status              string `json:"status"`
	K8sNamespace        string `json:"k8s_namespace"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type namespacePayload struct {
	Slug                string `json:"slug,omitempty"`
	AllocatedContainers int64  `json:"allocated_containers"`
	AllocatedVolumeGB   int64  `json:"allocated_volume_gb"`
}

type paginatedContainerList struct {
	Next    string         `json:"next"`
	Results []apiContainer `json:"results"`
}

type paginatedVolumeList struct {
	Next    string      `json:"next"`
	Results []apiVolume `json:"results"`
}

type paginatedInvitationList struct {
	Next    string          `json:"next"`
	Results []apiInvitation `json:"results"`
}

type paginatedKubectlCredentialList struct {
	Next    string                 `json:"next"`
	Results []apiKubectlCredential `json:"results"`
}

type paginatedMemberList struct {
	Next    string      `json:"next"`
	Results []apiMember `json:"results"`
}

type paginatedNamespaceList struct {
	Next    string         `json:"next"`
	Results []apiNamespace `json:"results"`
}

type paginatedOrgList struct {
	Next    string   `json:"next"`
	Results []apiOrg `json:"results"`
}

func (m *apiMember) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID       json.RawMessage `json:"id"`
		Email    string          `json:"email"`
		Role     string          `json:"role"`
		JoinedAt string          `json:"joined_at"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.ID = decodeJSONString(raw.ID)
	m.Email = raw.Email
	m.Role = raw.Role
	m.JoinedAt = raw.JoinedAt
	return nil
}

func decodeJSONString(data json.RawMessage) string {
	if len(data) == 0 || string(data) == "null" {
		return ""
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		return stringValue
	}

	var intValue int64
	if err := json.Unmarshal(data, &intValue); err == nil {
		return fmt.Sprint(intValue)
	}

	return ""
}

func nextPagePath(next string) string {
	if next == "" {
		return ""
	}
	parsed, err := url.Parse(next)
	if err == nil && parsed.IsAbs() {
		return parsed.RequestURI()
	}
	return next
}

func (c *APIClient) ListContainers(ctx context.Context) ([]apiContainer, error) {
	path := apiBasePath + "/containers/"
	containers := []apiContainer{}
	for path != "" {
		var page paginatedContainerList
		if err := c.do(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, err
		}
		containers = append(containers, page.Results...)
		path = nextPagePath(page.Next)
	}
	return containers, nil
}

func (c *APIClient) CreateContainer(ctx context.Context, payload containerPayload) (apiContainer, error) {
	var container apiContainer
	err := c.do(ctx, http.MethodPost, apiBasePath+"/containers/", payload, &container)
	return container, err
}

func (c *APIClient) GetContainer(ctx context.Context, id string) (apiContainer, error) {
	var container apiContainer
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s/containers/%s/", apiBasePath, url.PathEscape(id)), nil, &container)
	return container, err
}

func (c *APIClient) UpdateContainer(ctx context.Context, id string, payload containerPayload) (apiContainer, error) {
	var container apiContainer
	err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/containers/%s/", apiBasePath, url.PathEscape(id)), payload, &container)
	return container, err
}

func (c *APIClient) DeleteContainer(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/containers/%s/", apiBasePath, url.PathEscape(id)), nil, nil)
}

func (c *APIClient) ListVolumes(ctx context.Context) ([]apiVolume, error) {
	path := apiBasePath + "/volumes/"
	volumes := []apiVolume{}
	for path != "" {
		var page paginatedVolumeList
		if err := c.do(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, err
		}
		volumes = append(volumes, page.Results...)
		path = nextPagePath(page.Next)
	}
	return volumes, nil
}

func (c *APIClient) CreateVolume(ctx context.Context, payload volumePayload) (apiVolume, error) {
	var volume apiVolume
	err := c.do(ctx, http.MethodPost, apiBasePath+"/volumes/", payload, &volume)
	return volume, err
}

func (c *APIClient) GetVolume(ctx context.Context, id string) (apiVolume, error) {
	var volume apiVolume
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s/volumes/%s/", apiBasePath, url.PathEscape(id)), nil, &volume)
	return volume, err
}

func (c *APIClient) DeleteVolume(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/volumes/%s/", apiBasePath, url.PathEscape(id)), nil, nil)
}

func (c *APIClient) ListInvitations(ctx context.Context) ([]apiInvitation, error) {
	path := apiBasePath + "/invitations/"
	invitations := []apiInvitation{}
	for path != "" {
		var page paginatedInvitationList
		if err := c.do(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, err
		}
		invitations = append(invitations, page.Results...)
		path = nextPagePath(page.Next)
	}
	return invitations, nil
}

func (c *APIClient) CreateInvitation(ctx context.Context, payload invitationPayload) (apiInvitation, error) {
	var invitation apiInvitation
	err := c.do(ctx, http.MethodPost, apiBasePath+"/invitations/", payload, &invitation)
	return invitation, err
}

func (c *APIClient) GetInvitation(ctx context.Context, id string) (apiInvitation, error) {
	var invitation apiInvitation
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s/invitations/%s/", apiBasePath, url.PathEscape(id)), nil, &invitation)
	return invitation, err
}

func (c *APIClient) DeleteInvitation(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/invitations/%s/", apiBasePath, url.PathEscape(id)), nil, nil)
}

func (c *APIClient) ListKubectlCredentials(ctx context.Context) ([]apiKubectlCredential, error) {
	path := apiBasePath + "/kubectl-credentials/"
	credentials := []apiKubectlCredential{}
	for path != "" {
		var page paginatedKubectlCredentialList
		if err := c.do(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, err
		}
		credentials = append(credentials, page.Results...)
		path = nextPagePath(page.Next)
	}
	return credentials, nil
}

func (c *APIClient) CreateKubectlCredential(ctx context.Context) (apiKubectlCredential, error) {
	var credential apiKubectlCredential
	err := c.do(ctx, http.MethodPost, apiBasePath+"/kubectl-credentials/", nil, &credential)
	return credential, err
}

func (c *APIClient) GetKubectlCredential(ctx context.Context, id string) (apiKubectlCredential, error) {
	var credential apiKubectlCredential
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s/kubectl-credentials/%s/", apiBasePath, url.PathEscape(id)), nil, &credential)
	return credential, err
}

func (c *APIClient) DeleteKubectlCredential(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/kubectl-credentials/%s/", apiBasePath, url.PathEscape(id)), nil, nil)
}

func (c *APIClient) GetOrg(ctx context.Context, slug string) (apiOrg, error) {
	var org apiOrg
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s/orgs/%s/", apiBasePath, url.PathEscape(slug)), nil, &org)
	return org, err
}

func (c *APIClient) UpdateOrg(ctx context.Context, slug string, payload orgPayload) (apiOrg, error) {
	var org apiOrg
	err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/orgs/%s/", apiBasePath, url.PathEscape(slug)), payload, &org)
	return org, err
}

func (c *APIClient) ListOrgs(ctx context.Context) ([]apiOrg, error) {
	path := apiBasePath + "/orgs/"
	orgs := []apiOrg{}
	for path != "" {
		var page paginatedOrgList
		if err := c.do(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, err
		}
		orgs = append(orgs, page.Results...)
		path = nextPagePath(page.Next)
	}
	return orgs, nil
}

func (c *APIClient) ListMembers(ctx context.Context) ([]apiMember, error) {
	path := apiBasePath + "/members/"
	members := []apiMember{}
	for path != "" {
		var page paginatedMemberList
		if err := c.do(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, err
		}
		members = append(members, page.Results...)
		path = nextPagePath(page.Next)
	}
	return members, nil
}

func (c *APIClient) GetMember(ctx context.Context, id string) (apiMember, error) {
	var member apiMember
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s/members/%s/", apiBasePath, url.PathEscape(id)), nil, &member)
	return member, err
}

func (c *APIClient) UpdateMember(ctx context.Context, id string, payload memberPayload) (apiMember, error) {
	var member apiMember
	err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/members/%s/", apiBasePath, url.PathEscape(id)), payload, &member)
	return member, err
}

func (c *APIClient) DeleteMember(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/members/%s/", apiBasePath, url.PathEscape(id)), nil, nil)
}

func (c *APIClient) CreateNamespace(ctx context.Context, payload namespacePayload) (apiNamespace, error) {
	var namespace apiNamespace
	err := c.do(ctx, http.MethodPost, apiBasePath+"/namespaces/", payload, &namespace)
	return namespace, err
}

func (c *APIClient) ListNamespaces(ctx context.Context) ([]apiNamespace, error) {
	path := apiBasePath + "/namespaces/"
	namespaces := []apiNamespace{}
	for path != "" {
		var page paginatedNamespaceList
		if err := c.do(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, err
		}
		namespaces = append(namespaces, page.Results...)
		path = nextPagePath(page.Next)
	}
	return namespaces, nil
}

func (c *APIClient) GetNamespace(ctx context.Context, id string) (apiNamespace, error) {
	var namespace apiNamespace
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s/namespaces/%s/", apiBasePath, url.PathEscape(id)), nil, &namespace)
	return namespace, err
}

func (c *APIClient) UpdateNamespace(ctx context.Context, id string, payload namespacePayload) (apiNamespace, error) {
	var namespace apiNamespace
	err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/namespaces/%s/", apiBasePath, url.PathEscape(id)), payload, &namespace)
	return namespace, err
}

func (c *APIClient) DeleteNamespace(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/namespaces/%s/", apiBasePath, url.PathEscape(id)), nil, nil)
}
