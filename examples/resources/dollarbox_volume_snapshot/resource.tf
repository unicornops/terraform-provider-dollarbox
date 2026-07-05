resource "dollarbox_volume_snapshot" "before_upgrade" {
  namespace_id = 42
  pvc_name     = "app-data"
  name         = "before-upgrade"
  labels = {
    environment = "production"
  }
}
