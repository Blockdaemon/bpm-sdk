gcp_project                = "blockdaemon-development"
gcp_bucket_name            = "prod-bpm-sdk-docs"
k8s_context                = "gke_blockdaemon-production_us-central1-a_app-cluster-prod"
k8s_namespace              = "bpm-sdk-docs"
k8s_image_pull_secret_name = "regcred"

nginx_ingress_loadbalancer = "api.blockdaemon.com"

dns_env_subdomain = ""
dns_domain        = "sdk.bpm.docs.blockdaemon.com"

redoc_image = "redocly/redoc:v2.0.0-rc.41"

