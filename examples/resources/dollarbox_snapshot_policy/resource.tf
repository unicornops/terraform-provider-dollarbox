resource "dollarbox_snapshot_policy" "data" {
  namespace_id   = 42
  pvc_name       = "app-data"
  retention_days = 7
}
