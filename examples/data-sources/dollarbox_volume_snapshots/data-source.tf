data "dollarbox_volume_snapshots" "data" {
  namespace_id = 42
  pvc_name     = "app-data"
}
