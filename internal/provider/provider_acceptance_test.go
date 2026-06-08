package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dollarbox": providerserver.NewProtocol6WithError(New("test")()),
}

func TestAccOrgDataSource(t *testing.T) {
	testAccSkipUnlessEnabled(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dollarbox_org.test", "slug", testAccOrg()),
					resource.TestCheckResourceAttrSet("data.dollarbox_org.test", "name"),
					resource.TestCheckResourceAttrSet("data.dollarbox_org.test", "status"),
					resource.TestCheckResourceAttrSet("data.dollarbox_org.test", "created_at"),
				),
			},
		},
	})
}

func TestAccContainerResource(t *testing.T) {
	testAccSkipUnlessEnabled(t)
	testAccSkipUnlessBillableResourceEnabled(t, "container")

	name := testAccName("container")

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroy("dollarbox_container", func(ctx context.Context, client *APIClient, id string) error {
			_, err := client.GetContainer(ctx, id)
			return err
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerResourceConfig(name, "nginx:1.27"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dollarbox_container.test", "name", name),
					resource.TestCheckResourceAttr("dollarbox_container.test", "image", "nginx:1.27"),
					resource.TestCheckResourceAttrSet("dollarbox_container.test", "id"),
					resource.TestCheckResourceAttrSet("dollarbox_container.test", "status"),
					resource.TestCheckResourceAttrSet("dollarbox_container.test", "created_at"),
				),
			},
			{
				Config: testAccContainerResourceConfig(name, "nginx:1.27-alpine"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dollarbox_container.test", "name", name),
					resource.TestCheckResourceAttr("dollarbox_container.test", "image", "nginx:1.27-alpine"),
					resource.TestCheckResourceAttrSet("dollarbox_container.test", "updated_at"),
				),
			},
			{
				ResourceName:      "dollarbox_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVolumeResource(t *testing.T) {
	testAccSkipUnlessEnabled(t)
	testAccSkipUnlessBillableResourceEnabled(t, "volume")

	name := testAccName("volume")

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroy("dollarbox_volume", func(ctx context.Context, client *APIClient, id string) error {
			_, err := client.GetVolume(ctx, id)
			return err
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dollarbox_volume.test", "name", name),
					resource.TestCheckResourceAttr("dollarbox_volume.test", "size_gb", "10"),
					resource.TestCheckResourceAttrSet("dollarbox_volume.test", "id"),
					resource.TestCheckResourceAttrSet("dollarbox_volume.test", "status"),
					resource.TestCheckResourceAttrSet("dollarbox_volume.test", "created_at"),
				),
			},
			{
				ResourceName:      "dollarbox_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccInvitationResource(t *testing.T) {
	testAccSkipUnlessEnabled(t)

	email := fmt.Sprintf("%s@example.invalid", testAccName("invitation"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroy("dollarbox_invitation", func(ctx context.Context, client *APIClient, id string) error {
			_, err := client.GetInvitation(ctx, id)
			return err
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccInvitationResourceConfig(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dollarbox_invitation.test", "email", email),
					resource.TestCheckResourceAttr("dollarbox_invitation.test", "role", "member"),
					resource.TestCheckResourceAttr("dollarbox_invitation.test", "accepted", "false"),
					resource.TestCheckResourceAttrSet("dollarbox_invitation.test", "id"),
					resource.TestCheckResourceAttrSet("dollarbox_invitation.test", "created_at"),
				),
			},
			{
				ResourceName:      "dollarbox_invitation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKubectlCredentialResource(t *testing.T) {
	testAccSkipUnlessEnabled(t)
	testAccPreCheck(t)()
	skip, err := testAccSkipUnlessKubectlEnabled()
	if err != nil {
		t.Fatal(err)
	}
	if skip {
		t.Skip("DollarBox test org does not have kubectl credentials enabled")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroy("dollarbox_kubectl_credential", func(ctx context.Context, client *APIClient, id string) error {
			_, err := client.GetKubectlCredential(ctx, id)
			return err
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccKubectlCredentialResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dollarbox_kubectl_credential.test", "id"),
					resource.TestCheckResourceAttr("dollarbox_kubectl_credential.test", "org", testAccOrg()),
					resource.TestCheckResourceAttrSet("dollarbox_kubectl_credential.test", "sa_name"),
					resource.TestCheckResourceAttrSet("dollarbox_kubectl_credential.test", "kubeconfig"),
					resource.TestCheckResourceAttrSet("dollarbox_kubectl_credential.test", "created_at"),
				),
			},
			{
				ResourceName:            "dollarbox_kubectl_credential.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kubeconfig"},
			},
		},
	})
}

func testAccSkipUnlessEnabled(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests skipped unless TF_ACC is set")
	}
}

func testAccPreCheck(t *testing.T) func() {
	return func() {
		t.Helper()

		if strings.TrimSpace(os.Getenv("DOLLARBOX_TOKEN")) == "" {
			t.Fatal("DOLLARBOX_TOKEN must be set for acceptance tests")
		}
		if strings.TrimSpace(os.Getenv("DOLLARBOX_ORG")) == "" {
			t.Fatal("DOLLARBOX_ORG must be set for acceptance tests")
		}
	}
}

func testAccSkipUnlessKubectlEnabled() (bool, error) {
	org, err := testAccAPIClient().GetOrg(context.Background(), testAccOrg())
	if err != nil {
		return false, fmt.Errorf("check kubectl support: %w", err)
	}
	return !org.KubectlEnabled, nil
}

func testAccSkipUnlessBillableResourceEnabled(t *testing.T, resourceName string) {
	t.Helper()

	testAccPreCheck(t)()
	client := testAccAPIClient()
	ctx := context.Background()

	switch resourceName {
	case "container":
		container, err := client.CreateContainer(ctx, containerPayload{
			Name:    testAccName("preflight-container"),
			Image:   "nginx:1.27",
			Env:     map[string]string{},
			Command: []string{},
		})
		if err != nil {
			testAccSkipIfBillingUnavailable(t, resourceName, err)
			t.Fatalf("check %s acceptance preflight: %v", resourceName, err)
		}
		if err := client.DeleteContainer(ctx, fmt.Sprint(container.ID)); err != nil && !isNotFoundError(err) {
			t.Fatalf("clean up %s acceptance preflight: %v", resourceName, err)
		}
	case "volume":
		volume, err := client.CreateVolume(ctx, volumePayload{
			Name:   testAccName("preflight-volume"),
			SizeGB: 10,
		})
		if err != nil {
			testAccSkipIfBillingUnavailable(t, resourceName, err)
			t.Fatalf("check %s acceptance preflight: %v", resourceName, err)
		}
		if err := client.DeleteVolume(ctx, fmt.Sprint(volume.ID)); err != nil && !isNotFoundError(err) {
			t.Fatalf("clean up %s acceptance preflight: %v", resourceName, err)
		}
	default:
		t.Fatalf("unsupported billable resource preflight %q", resourceName)
	}
}

func testAccSkipIfBillingUnavailable(t *testing.T, resourceName string, err error) {
	t.Helper()

	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Code == "billing_error" {
		detail := apiErr.Message
		if detail == "" {
			detail = apiErr.Error()
		}
		t.Skipf("DollarBox test org cannot create %s resources: %s", resourceName, detail)
	}
}

func testAccCheckResourceDestroy(resourceType string, fetch func(context.Context, *APIClient, string) error) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccAPIClient()
		ctx := context.Background()

		for _, rs := range state.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			if rs.Primary == nil || rs.Primary.ID == "" {
				continue
			}

			err := fetch(ctx, client, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("%s %s still exists", resourceType, rs.Primary.ID)
			}
			if !isNotFoundError(err) {
				return fmt.Errorf("check %s %s destroy: %w", resourceType, rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccAPIClient() *APIClient {
	return NewAPIClient(ClientConfig{
		Endpoint: testAccEndpoint(),
		Token:    strings.TrimSpace(os.Getenv("DOLLARBOX_TOKEN")),
		Org:      testAccOrg(),
	})
}

func testAccEndpoint() string {
	if endpoint := strings.TrimSpace(os.Getenv("DOLLARBOX_ENDPOINT")); endpoint != "" {
		return strings.TrimRight(endpoint, "/")
	}
	return defaultEndpoint
}

func testAccOrg() string {
	return strings.TrimSpace(os.Getenv("DOLLARBOX_ORG"))
}

func testAccProviderConfig() string {
	return fmt.Sprintf(`
provider "dollarbox" {
  endpoint = %[1]q
  org      = %[2]q
}
`, testAccEndpoint(), testAccOrg())
}

func testAccOrgDataSourceConfig() string {
	return testAccProviderConfig() + `
data "dollarbox_org" "test" {}
`
}

func testAccContainerResourceConfig(name, image string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "dollarbox_container" "test" {
  name  = %[1]q
  image = %[2]q
}
`, name, image)
}

func testAccVolumeResourceConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "dollarbox_volume" "test" {
  name    = %[1]q
  size_gb = 10
}
`, name)
}

func testAccInvitationResourceConfig(email string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "dollarbox_invitation" "test" {
  email = %[1]q
  role  = "member"
}
`, email)
}

func testAccKubectlCredentialResourceConfig() string {
	return testAccProviderConfig() + `
resource "dollarbox_kubectl_credential" "test" {}
`
}

func testAccName(prefix string) string {
	return fmt.Sprintf("tfacc-%s-%s", prefix, acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
}
