##@ FLP

.PHONY: generate-flp-configuration
generate-flp-configuration: ## Generate metrics configuration
	@echo -e "\n==> Generate metrics configuration"
	cp -r network_definitions /tmp/network_definitions
	$(OCI_BIN) run --entrypoint /app/confgenerator -v /tmp/network_definitions:/app/network_definitions -v /tmp/destDir:/app/destDir quay.io/netobserv/flowlogs-pipeline:latest --srcFolder /app/network_definitions --destConfFile /app/destDir/flowlogs-pipeline.conf.yaml
	# TODO: add --destDocFile and --destGrafanaJsonnetFolder
	cp /tmp/destDir/flowlogs-pipeline.conf.yaml config/flp/flowlogs-pipeline.conf.yaml
	@echo -e "\n===> flowlogs-configuration.yaml available in config/flp/flowlogs-configuration.yaml \n"
	# TODO: delete container

