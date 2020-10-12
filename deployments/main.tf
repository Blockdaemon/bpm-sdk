terraform {
  required_version = ">= 0.13"
  backend "gcs" {
    bucket = "blockdaemon-terraform-state"
    prefix = "bpm/sdk-docs"
  }
  required_providers {
    google = {
      source = "hashicorp/google"
    }
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    ns1 = {
      source = "terraform-providers/ns1"
    }
  }
}

provider "google" {
  project = var.gcp_project
}

provider "kubernetes" {
  config_context = var.k8s_context
}

provider "ns1" {
  apikey = var.ns1_apikey
}

locals {
  fqdn = "${var.dns_env_subdomain}${var.dns_domain}"
}
