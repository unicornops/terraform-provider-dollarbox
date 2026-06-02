package provider

import (
	"context"
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
