build:
	@go build -o terraform-provider-lambda
	@terraform init

clean:
	@rm -rf $(HOME)/Library/Caches/terraform-provider-lambda
	@rm -rf ./terraform.tfstate
	@rm -rf ./terraform.tfstate.backup

apply: build
	@terraform apply -auto-approve
