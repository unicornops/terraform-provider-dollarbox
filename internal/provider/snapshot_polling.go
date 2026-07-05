package provider

import (
	"context"
	"fmt"
	"time"
)

const snapshotOperationTimeout = 30 * time.Minute

var snapshotPollInterval = 5 * time.Second

func waitForSnapshotPolicyActive(
	ctx context.Context,
	client *APIClient,
	namespaceID, pvcName string,
	policy apiSnapshotPolicy,
) (apiSnapshotPolicy, error) {
	for {
		switch policy.Status {
		case "active":
			return policy, nil
		case "billing_blocked", "error":
			return policy, fmt.Errorf("snapshot policy entered terminal status %q: %s", policy.Status, policy.Error)
		}

		if err := waitForNextSnapshotPoll(ctx); err != nil {
			return policy, fmt.Errorf("wait for snapshot policy activation (last status %q): %w", policy.Status, err)
		}
		var err error
		policy, err = client.GetSnapshotPolicy(ctx, namespaceID, pvcName)
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return policy, fmt.Errorf("read snapshot policy while waiting for activation: %w", err)
		}
	}
}

func waitForSnapshotPolicyDeleted(
	ctx context.Context,
	client *APIClient,
	namespaceID, pvcName string,
	policy apiSnapshotPolicy,
) error {
	for {
		if policy.Status == "disabled" {
			return nil
		}
		if policy.Status == "error" || policy.Error != "" {
			return fmt.Errorf("snapshot policy entered terminal status %q while retiring: %s", policy.Status, policy.Error)
		}
		if err := waitForNextSnapshotPoll(ctx); err != nil {
			return fmt.Errorf("wait for snapshot policy retirement (last status %q): %w", policy.Status, err)
		}
		var err error
		policy, err = client.GetSnapshotPolicy(ctx, namespaceID, pvcName)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return fmt.Errorf("read snapshot policy while waiting for retirement: %w", err)
		}
	}
}

func waitForVolumeSnapshotReady(
	ctx context.Context,
	client *APIClient,
	namespaceID, pvcName string,
	snapshot apiVolumeSnapshot,
) (apiVolumeSnapshot, error) {
	for {
		switch snapshot.Status {
		case "ready":
			return snapshot, nil
		case "failed", "deleted":
			return snapshot, fmt.Errorf("volume snapshot entered terminal status %q: %s", snapshot.Status, snapshot.Error)
		}

		if err := waitForNextSnapshotPoll(ctx); err != nil {
			return snapshot, fmt.Errorf("wait for volume snapshot readiness (last status %q): %w", snapshot.Status, err)
		}
		var err error
		snapshot, err = client.GetVolumeSnapshot(ctx, namespaceID, pvcName, snapshot.ID)
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return snapshot, fmt.Errorf("read volume snapshot while waiting for readiness: %w", err)
		}
	}
}

func waitForVolumeSnapshotDeleted(
	ctx context.Context,
	client *APIClient,
	namespaceID, pvcName string,
	snapshot apiVolumeSnapshot,
) error {
	for {
		if snapshot.Status == "deleted" {
			return nil
		}
		if snapshot.Status == "failed" || snapshot.Error != "" {
			return fmt.Errorf("volume snapshot entered terminal status %q while deleting: %s", snapshot.Status, snapshot.Error)
		}
		if err := waitForNextSnapshotPoll(ctx); err != nil {
			return fmt.Errorf("wait for volume snapshot deletion (last status %q): %w", snapshot.Status, err)
		}
		var err error
		snapshot, err = client.GetVolumeSnapshot(ctx, namespaceID, pvcName, snapshot.ID)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return fmt.Errorf("read volume snapshot while waiting for deletion: %w", err)
		}
	}
}

func waitForSnapshotRestoreBound(
	ctx context.Context,
	client *APIClient,
	namespaceID, sourcePVCName string,
	restore apiSnapshotRestore,
) (apiSnapshotRestore, error) {
	for {
		switch restore.Status {
		case "bound":
			return restore, nil
		case "failed":
			return restore, fmt.Errorf("snapshot restore entered terminal status %q: %s", restore.Status, restore.Error)
		}

		if err := waitForNextSnapshotPoll(ctx); err != nil {
			return restore, fmt.Errorf("wait for snapshot restore binding (last status %q): %w", restore.Status, err)
		}
		var err error
		restore, err = client.GetSnapshotRestore(ctx, namespaceID, sourcePVCName, restore.ID)
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return restore, fmt.Errorf("read snapshot restore while waiting for binding: %w", err)
		}
	}
}

func waitForNextSnapshotPoll(ctx context.Context) error {
	timer := time.NewTimer(snapshotPollInterval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
