# Install CRDs into a cluster:
#   kubectl apply -f config/crd/bases/consoleui.kubevirt.io_kubevirtconsoleuis.yaml
#
# Or with kustomize:
#   kubectl apply -k config/crd

CONTROLLER_GEN ?= $(shell command -v controller-gen 2>/dev/null)

.PHONY: manifests
manifests: ## Generate CRD manifests (requires controller-gen v0.16.x).
	@if [ -z "$(CONTROLLER_GEN)" ]; then \
		echo "Install controller-gen: go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.5"; \
		exit 1; \
	fi
	$(CONTROLLER_GEN) crd paths="./api/..." output:crd:artifacts:config=config/crd/bases

.PHONY: test
test: ## Run unit tests.
	go test ./...

.PHONY: install-crd
install-crd: ## Apply CRD to the current kubeconfig context.
	kubectl apply -f config/crd/bases/consoleui.kubevirt.io_kubevirtconsoleuis.yaml
