TF_WORKSPACE ?= dev
export TF_WORKSPACE

.PHONY: plan
plan:
	terraform plan -var-file=vars-$(TF_WORKSPACE).tfvars

.PHONY: apply
apply:
	terraform apply -var-file=vars-$(TF_WORKSPACE).tfvars

