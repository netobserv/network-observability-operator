##@ FLP

.PHONY: generate-custom-configuration
generate-custom-configuration: ## Generate custom metrics configuration
	@echo -e "\n==> Generate metrics configuration"
	rm -rf /tmp/network_definitions || true
	rm -rf /tmp/destDir || true
	cp -r network_definitions /tmp/network_definitions
	$(OCI_BIN) run --entrypoint /app/confgenerator -v /tmp/network_definitions:/app/network_definitions -v /tmp/destDir:/app/destDir quay.io/netobserv/flowlogs-pipeline:latest --srcFolder /app/network_definitions --destConfFile /app/destDir/flowlogs-pipeline.conf.yaml
	# TODO: add --destDocFile and --destGrafanaJsonnetFolder
	cp /tmp/destDir/flowlogs-pipeline.conf.yaml config/flp/flowlogs-pipeline.conf.yaml
	@echo -e "\n===> flowlogs-configuration.yaml available in config/flp/flowlogs-pipeline.conf.yaml \n"
	# TODO: delete container

.PHONY: generate-custom-pipeline ## Generate custom pipeline
generate-custom-pipeline: ## Generate custom metrics pipeline
	@echo -e "\n==> Generate metrics pipeline"
	$(OCI_BIN) build -t quay.io/netobserv/flowlogs-pipeline:custom -f config/flp/custom.Dockerfile .
	@echo -e "\n==> Custom Metrics pipeline created in quay.io/netobserv/flowlogs-pipeline:custom \n"

.PHONY: kind-load-image
kind-load-image: ## Load image to kind
	$(eval tmpfile="/tmp/flp.tar")
	-rm $(tmpfile)
	$(OCI_BIN) save quay.io/netobserv/flowlogs-pipeline:custom -o $(tmpfile)
	$(KIND) load image-archive $(tmpfile)
	-rm $(tmpfile)

