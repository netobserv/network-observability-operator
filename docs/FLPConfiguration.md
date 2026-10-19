# FLP YAML configuration

The operator generates static and dynamic flowlogs-pipeline (FLP) configuration
as YAML under the ConfigMap key `config.yaml`. The processor's `--config` argument
and its `dynamicParameters.fileName` reference use that key. There is no separate
generated JSON configuration or format selection in FlowCollector.

The static configuration digest is calculated from the generated YAML, preserving
the existing pod rollout behavior when static settings change. Dynamic settings
are read from their ConfigMap at startup and through FLP's existing watcher.

## Manual configuration

Removing a ConfigMap's operator ownership prevents the operator from overwriting
it. Keep the `config.yaml` key and the existing configuration schema when editing
its contents. Static changes require restarting the processor pods. Dynamic
changes are applied only for stages that support hot reload.

## Delivery and upgrades

Deploy this operator change only with an FLP image containing YAML support in
both its initial dynamic configuration reader and ConfigMap update reader. An
operator Go module update does not replace the deployed FLP binary.

ConfigMap names remain unchanged. When upgrading from a JSON-producing operator,
managed ConfigMaps change from `config.json` to `config.yaml` and processor pods
roll out with the new argument. Old pods still expecting `config.json` cannot
reload the new dynamic key, and restarting an old pod during that interval can
fail to find its configuration. Upgrade and rollback must be exercised before
release; this change does not promise uninterrupted coexistence of old and new
processor revisions.

For manually managed ConfigMaps without operator ownership, migrate the data key
and contents to `config.yaml` before rolling out the new processor specification.
Preserve custom settings: the operator intentionally does not convert or replace
unowned ConfigMaps. A rollback to the old operator similarly requires restoring
the old filename and JSON format for such manually managed configuration.
