package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const snapshotAcceptanceTimeout = 10 * time.Minute

type snapshotAcceptanceFixture struct {
	NamespaceID         string
	KubernetesNamespace string
	SourcePVCName       string
	TargetPVCName       string
	KubeconfigPath      string
}

func TestAccSnapshotResources(t *testing.T) {
	testAccSkipUnlessEnabled(t)
	fixture := newSnapshotAcceptanceFixture(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotResourcesDestroyed(fixture.NamespaceID, fixture.SourcePVCName),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig(fixture.NamespaceID, fixture.SourcePVCName, 7, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dollarbox_snapshot_policy.test", "retention_days", "7"),
					resource.TestCheckResourceAttr("dollarbox_snapshot_policy.test", "status", "active"),
					resource.TestCheckResourceAttrSet("dollarbox_snapshot_policy.test", "billed_gb"),
				),
			},
			{
				Config: testAccSnapshotConfig(fixture.NamespaceID, fixture.SourcePVCName, 5, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dollarbox_snapshot_policy.test", "retention_days", "5"),
					resource.TestCheckResourceAttr("dollarbox_volume_snapshot.test", "name", "terraform-acceptance"),
					resource.TestCheckResourceAttr("dollarbox_volume_snapshot.test", "labels.managed-by", "terraform"),
					resource.TestCheckResourceAttr("dollarbox_volume_snapshot.test", "kind", "manual"),
					resource.TestCheckResourceAttr("dollarbox_volume_snapshot.test", "status", "ready"),
					resource.TestCheckResourceAttrSet("data.dollarbox_volume_snapshot.test", "id"),
					resource.TestCheckResourceAttrSet("data.dollarbox_volume_snapshots.test", "snapshots.0.id"),
				),
			},
			{
				ResourceName:            "dollarbox_snapshot_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status", "next_snapshot_at", "last_snapshot_at", "updated_at"},
			},
			{
				ResourceName: "dollarbox_volume_snapshot.test",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["dollarbox_volume_snapshot.test"]
					if !ok || rs.Primary == nil {
						return "", errors.New("volume snapshot state not found")
					}
					return fmt.Sprintf("%s/%s/%s", fixture.NamespaceID, fixture.SourcePVCName, rs.Primary.ID), nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status", "updated_at"},
			},
			{
				Config: testAccSnapshotRestoreConfig(fixture.NamespaceID, fixture.SourcePVCName, fixture.TargetPVCName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dollarbox_snapshot_restore.test", "target_pvc_name", fixture.TargetPVCName),
					resource.TestCheckResourceAttr("dollarbox_snapshot_restore.test", "status", "bound"),
					resource.TestCheckResourceAttrSet("dollarbox_snapshot_restore.test", "id"),
				),
			},
		},
	})
}

func newSnapshotAcceptanceFixture(t *testing.T) snapshotAcceptanceFixture {
	t.Helper()
	testAccPreCheck(t)()
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Fatal("kubectl must be installed for snapshot acceptance tests")
	}
	skip, err := testAccSkipUnlessKubectlEnabled()
	if err != nil {
		t.Fatalf("check kubectl support: %v", err)
	}
	if skip {
		t.Skip("DollarBox test org does not have kubectl credentials enabled")
	}

	client := testAccAPIClient()
	ctx, cancel := context.WithTimeout(context.Background(), snapshotAcceptanceTimeout)
	defer cancel()
	namespace, err := client.CreateNamespace(ctx, namespacePayload{
		Slug:                strings.ToLower(testAccName("snapshots")),
		AllocatedContainers: 0,
		AllocatedVolumeGB:   2,
	})
	if err != nil {
		t.Fatalf("create snapshot acceptance namespace: %v", err)
	}
	namespaceID := strconv.FormatInt(namespace.ID, 10)
	t.Cleanup(func() { cleanupSnapshotAcceptanceNamespace(t, client, namespaceID) })
	namespace = waitForAcceptanceNamespaceActive(t, client, namespaceID)

	credential, err := client.CreateKubectlCredential(ctx)
	if err != nil {
		t.Fatalf("create snapshot acceptance kubectl credential: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteKubectlCredential(context.Background(), strconv.FormatInt(credential.ID, 10)); err != nil && !isNotFoundError(err) {
			t.Errorf("delete snapshot acceptance kubectl credential: %v", err)
		}
	})
	if strings.TrimSpace(credential.Kubeconfig) == "" {
		t.Fatal("snapshot acceptance kubectl credential did not include kubeconfig")
	}
	kubeconfigPath := filepath.Join(t.TempDir(), "kubeconfig")
	if err := os.WriteFile(kubeconfigPath, []byte(credential.Kubeconfig), 0o600); err != nil {
		t.Fatalf("write snapshot acceptance kubeconfig: %v", err)
	}

	fixture := snapshotAcceptanceFixture{
		NamespaceID:         namespaceID,
		KubernetesNamespace: namespace.K8sNamespace,
		SourcePVCName:       strings.ToLower(testAccName("snapshot-source")),
		TargetPVCName:       strings.ToLower(testAccName("snapshot-restore")),
		KubeconfigPath:      kubeconfigPath,
	}
	t.Cleanup(func() { deleteAcceptancePVC(t, fixture, fixture.SourcePVCName) })
	t.Cleanup(func() { deleteAcceptancePVC(t, fixture, fixture.TargetPVCName) })
	createAcceptancePVC(t, fixture)
	return fixture
}

func waitForAcceptanceNamespaceActive(t *testing.T, client *APIClient, namespaceID string) apiNamespace {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), snapshotAcceptanceTimeout)
	defer cancel()
	for {
		namespace, err := client.GetNamespace(ctx, namespaceID)
		if err != nil {
			t.Fatalf("read snapshot acceptance namespace: %v", err)
		}
		switch namespace.Status {
		case "active":
			return namespace
		case "failed":
			t.Fatalf("snapshot acceptance namespace provisioning failed")
		}
		select {
		case <-ctx.Done():
			t.Fatalf("wait for snapshot acceptance namespace activation: %v", ctx.Err())
		case <-time.After(5 * time.Second):
		}
	}
}

func cleanupSnapshotAcceptanceNamespace(t *testing.T, client *APIClient, namespaceID string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), snapshotAcceptanceTimeout)
	defer cancel()
	if err := client.DeleteNamespace(ctx, namespaceID); err != nil && !isNotFoundError(err) {
		t.Errorf("delete snapshot acceptance namespace: %v", err)
		return
	}
	for {
		namespace, err := client.GetNamespace(ctx, namespaceID)
		if isNotFoundError(err) || (err == nil && namespace.Status == "deleted") {
			return
		}
		if err != nil {
			t.Errorf("read snapshot acceptance namespace during cleanup: %v", err)
			return
		}
		select {
		case <-ctx.Done():
			t.Errorf("wait for snapshot acceptance namespace deletion: %v", ctx.Err())
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func createAcceptancePVC(t *testing.T, fixture snapshotAcceptanceFixture) {
	t.Helper()
	manifest, err := json.Marshal(map[string]any{
		"apiVersion": "v1",
		"kind":       "PersistentVolumeClaim",
		"metadata":   map[string]any{"name": fixture.SourcePVCName},
		"spec": map[string]any{
			"accessModes":      []string{"ReadWriteOnce"},
			"storageClassName": "longhorn",
			"resources": map[string]any{
				"requests": map[string]string{"storage": "1Gi"},
			},
		},
	})
	if err != nil {
		t.Fatalf("encode snapshot acceptance PVC: %v", err)
	}
	if output, err := runAcceptanceKubectl(t, fixture, manifest, "create", "-f", "-"); err != nil {
		t.Fatalf("create snapshot acceptance PVC: %v\n%s", err, output)
	}
	if output, err := runAcceptanceKubectl(t, fixture, nil, "wait", "--for=jsonpath={.status.phase}=Bound", "persistentvolumeclaim/"+fixture.SourcePVCName, "--timeout=5m"); err != nil {
		t.Fatalf("wait for snapshot acceptance PVC: %v\n%s", err, output)
	}
}

func deleteAcceptancePVC(t *testing.T, fixture snapshotAcceptanceFixture, pvcName string) {
	t.Helper()
	if output, err := runAcceptanceKubectl(t, fixture, nil, "delete", "persistentvolumeclaim/"+pvcName, "--ignore-not-found=true", "--wait=true", "--timeout=5m"); err != nil {
		t.Errorf("delete snapshot acceptance PVC %s: %v\n%s", pvcName, err, output)
	}
}

func runAcceptanceKubectl(t *testing.T, fixture snapshotAcceptanceFixture, stdin []byte, args ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	baseArgs := []string{"--kubeconfig", fixture.KubeconfigPath, "--namespace", fixture.KubernetesNamespace}
	cmd := exec.CommandContext(ctx, "kubectl", append(baseArgs, args...)...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	return cmd.CombinedOutput()
}

func testAccCheckSnapshotResourcesDestroyed(namespaceID, pvcName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccAPIClient()
		ctx := context.Background()
		for _, rs := range state.RootModule().Resources {
			if rs.Primary == nil {
				continue
			}
			switch rs.Type {
			case "dollarbox_snapshot_policy":
				policy, err := client.GetSnapshotPolicy(ctx, namespaceID, pvcName)
				if err != nil && !isNotFoundError(err) {
					return fmt.Errorf("check snapshot policy destroy: %w", err)
				}
				if err == nil && policy.Status != "disabled" {
					return fmt.Errorf("snapshot policy still has status %q", policy.Status)
				}
			case "dollarbox_volume_snapshot":
				snapshot, err := client.GetVolumeSnapshot(ctx, namespaceID, pvcName, rs.Primary.ID)
				if err != nil && !isNotFoundError(err) {
					return fmt.Errorf("check volume snapshot destroy: %w", err)
				}
				if err == nil && snapshot.Status != "deleted" {
					return fmt.Errorf("volume snapshot still has status %q", snapshot.Status)
				}
			}
		}
		return nil
	}
}

func testAccSnapshotConfig(namespaceID, pvcName string, retentionDays int, includeSnapshot bool) string {
	config := testAccProviderConfig() + fmt.Sprintf(`
resource "dollarbox_snapshot_policy" "test" {
  namespace_id   = %[1]s
  pvc_name       = %[2]q
  retention_days = %[3]d
}
`, namespaceID, pvcName, retentionDays)
	if !includeSnapshot {
		return config
	}
	return config + fmt.Sprintf(`
resource "dollarbox_volume_snapshot" "test" {
  namespace_id = %[1]s
  pvc_name     = %[2]q
  name         = "terraform-acceptance"
  labels = {
    managed-by = "terraform"
  }
  depends_on = [dollarbox_snapshot_policy.test]
}

data "dollarbox_volume_snapshot" "test" {
  namespace_id = %[1]s
  pvc_name     = %[2]q
  id           = dollarbox_volume_snapshot.test.id
}

data "dollarbox_volume_snapshots" "test" {
  namespace_id = %[1]s
  pvc_name     = %[2]q
  depends_on   = [dollarbox_volume_snapshot.test]
}
`, namespaceID, pvcName)
}

func testAccSnapshotRestoreConfig(namespaceID, pvcName, targetPVCName string) string {
	return testAccSnapshotConfig(namespaceID, pvcName, 7, true) + fmt.Sprintf(`
resource "dollarbox_snapshot_restore" "test" {
  namespace_id    = %[1]s
  source_pvc_name = %[2]q
  snapshot_id     = dollarbox_volume_snapshot.test.id
  target_pvc_name = %[3]q
}
`, namespaceID, pvcName, targetPVCName)
}
