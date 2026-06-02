resource "dollarbox_container" "web" {
  name  = "web"
  image = "nginx:1.27"

  env = {
    NGINX_PORT = "80"
  }
}
