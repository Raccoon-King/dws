# DWS Helm Chart Makefile

.PHONY: help lint template install upgrade uninstall test clean package

CHART_NAME := dws
NAMESPACE := dws-system
RELEASE_NAME := dws

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

lint: ## Lint the Helm chart
	helm lint .

template: ## Generate Kubernetes manifests from the chart
	helm template $(RELEASE_NAME) . > manifests.yaml
	@echo "Manifests generated in manifests.yaml"

template-dev: ## Generate manifests with dev values
	helm template $(RELEASE_NAME) . -f values-dev.yaml > manifests-dev.yaml
	@echo "Dev manifests generated in manifests-dev.yaml"

template-prod: ## Generate manifests with prod values
	helm template $(RELEASE_NAME) . -f values-prod.yaml > manifests-prod.yaml
	@echo "Production manifests generated in manifests-prod.yaml"

dry-run: ## Perform a dry run installation
	helm install $(RELEASE_NAME) . --namespace $(NAMESPACE) --dry-run --debug

dry-run-dev: ## Perform a dry run with dev values
	helm install $(RELEASE_NAME) . --namespace $(NAMESPACE)-dev -f values-dev.yaml --dry-run --debug

dry-run-prod: ## Perform a dry run with prod values
	helm install $(RELEASE_NAME) . --namespace $(NAMESPACE)-prod -f values-prod.yaml --dry-run --debug

install: ## Install the chart (requires existing namespace)
	helm install $(RELEASE_NAME) . --namespace $(NAMESPACE)

install-dev: ## Install with development values (requires existing namespace)
	helm install $(RELEASE_NAME)-dev . --namespace $(NAMESPACE)-dev -f values-dev.yaml

install-prod: ## Install with production values (requires existing namespace)
	helm install $(RELEASE_NAME)-prod . --namespace $(NAMESPACE)-prod -f values-prod.yaml

upgrade: ## Upgrade the release
	helm upgrade $(RELEASE_NAME) . --namespace $(NAMESPACE)

upgrade-dev: ## Upgrade dev release
	helm upgrade $(RELEASE_NAME)-dev . --namespace $(NAMESPACE)-dev -f values-dev.yaml

upgrade-prod: ## Upgrade prod release
	helm upgrade $(RELEASE_NAME)-prod . --namespace $(NAMESPACE)-prod -f values-prod.yaml

uninstall: ## Uninstall the release
	helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE)

uninstall-dev: ## Uninstall dev release
	helm uninstall $(RELEASE_NAME)-dev --namespace $(NAMESPACE)-dev

uninstall-prod: ## Uninstall prod release
	helm uninstall $(RELEASE_NAME)-prod --namespace $(NAMESPACE)-prod

test: ## Run Helm tests
	helm test $(RELEASE_NAME) --namespace $(NAMESPACE)

test-dev: ## Run tests on dev deployment
	helm test $(RELEASE_NAME)-dev --namespace $(NAMESPACE)-dev

test-prod: ## Run tests on prod deployment
	helm test $(RELEASE_NAME)-prod --namespace $(NAMESPACE)-prod

status: ## Show release status
	helm status $(RELEASE_NAME) --namespace $(NAMESPACE)

list: ## List all releases
	helm list --all-namespaces

history: ## Show release history
	helm history $(RELEASE_NAME) --namespace $(NAMESPACE)

rollback: ## Rollback to previous version
	helm rollback $(RELEASE_NAME) --namespace $(NAMESPACE)

package: ## Package the chart
	helm package .
	@echo "Chart packaged as $(CHART_NAME)-*.tgz"

dependency-update: ## Update chart dependencies
	helm dependency update

clean: ## Clean generated files
	rm -f manifests*.yaml
	rm -f $(CHART_NAME)-*.tgz

validate: lint template ## Validate the chart (lint + template generation)
	@echo "Chart validation completed successfully"

# Development helpers
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	kubectl create namespace $(NAMESPACE)-dev --dry-run=client -o yaml | kubectl apply -f -

dev-teardown: ## Tear down development environment
	@echo "Tearing down development environment..."
	kubectl delete namespace $(NAMESPACE)-dev --ignore-not-found=true

# CI/CD helpers
ci-lint: ## Run CI linting
	helm lint . --strict

ci-template: ## Generate all templates for CI
	$(MAKE) template template-dev template-prod

ci-test: ci-lint ci-template ## Run all CI tests
	@echo "All CI tests passed"

# Iron Bank specific targets
ironbank-template: ## Generate templates with Iron Bank images
	helm template $(RELEASE_NAME) . \
		--set image.registry=registry1.dso.mil \
		--set image.repository=ironbank/opensource/dws \
		> manifests-ironbank.yaml
	@echo "Iron Bank manifests generated in manifests-ironbank.yaml"

# Custom registry targets
custom-registry-template: ## Generate templates with custom registry (see values-custom-registry.yaml)
	helm template $(RELEASE_NAME) . -f values-custom-registry.yaml > manifests-custom.yaml
	@echo "Custom registry manifests generated in manifests-custom.yaml"

custom-registry-install: ## Install with custom registry values (requires existing namespace)
	helm install $(RELEASE_NAME) . --namespace $(NAMESPACE) -f values-custom-registry.yaml