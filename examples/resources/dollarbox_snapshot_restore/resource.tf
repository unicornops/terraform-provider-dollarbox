resource "dollarbox_snapshot_restore" "recovered" {
  namespace_id    = 42
  source_pvc_name = "app-data"
  snapshot_id     = "11111111-2222-3333-4444-555555555555"
  target_pvc_name = "app-data-restored"
}
