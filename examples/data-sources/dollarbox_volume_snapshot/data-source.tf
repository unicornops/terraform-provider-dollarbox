data "dollarbox_volume_snapshot" "daily" {
  namespace_id = 42
  pvc_name     = "app-data"
  id           = "11111111-2222-3333-4444-555555555555"
}
