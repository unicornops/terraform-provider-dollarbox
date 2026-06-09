resource "dollarbox_namespace" "dev" {
  slug                 = "dev"
  allocated_containers = 2
  allocated_volume_gb  = 10
}
