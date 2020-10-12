data "ns1_zone" "my_zone" {
  zone = var.dns_zone
}

resource "ns1_record" "bpm_ns1_record" {
  zone   = data.ns1_zone.my_zone.zone
  domain = local.fqdn
  type   = "CNAME"
  ttl    = 3600

  answers {
    answer = var.nginx_ingress_loadbalancer
  }
}


