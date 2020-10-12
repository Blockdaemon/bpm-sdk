# protocol-docs

Terraform for deploying protocol documentation to k8s/ns1.

The project contains multiple workspaces: dev, stg, prod.
Example:
```shell script
make plan TF_WORKSPACE=dev
make apply TF_WORKSPACE=dev
```

## Requirements

### Google Cloud Platform

The `google` provider is configured to use application default credentials.

If you're not logged already, run this:

```shell script
gcloud auth application-default login
```

### Kubernetes (GKE)

The Terraform module also manages a handful of Kubernetes resources.

Set up your kubeconfig to include the relevant GKE contexts.
Example for development:

```shell script
gcloud container clusters \
  --project blockdaemon-development --region us-west2-a \
  get-credentials app-dev 
```

## Static Secrets

**Currently deprecating static secrets! Any secrets should go in Vault.**

Terraform reads static secrets from local dirs.
The Git history does not include those secrets, so consult a maintainer for access.

Secret paths under `secrets/{env}`:
- `dockerhub-regcred.json`: Docker login for GitLab container registry

