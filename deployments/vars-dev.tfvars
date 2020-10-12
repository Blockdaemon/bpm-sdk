gcp_project                = "blockdaemon-development"
gcp_bucket_name            = "dev-bpm-sdk-docs"
k8s_context                = "gke_blockdaemon-development_us-west2-a_app-dev"
k8s_namespace              = "bpm-sdk-docs"
k8s_image_pull_secret_name = "regcred"

nginx_ingress_loadbalancer = "dev.api.blockdaemon.com"

dns_env_subdomain = "dev."
dns_domain        = "sdk.bpm.docs.blockdaemon.com"

redoc_image = "redocly/redoc:v2.0.0-rc.41"

