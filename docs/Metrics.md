# Metrics in the NetObserv Operator

Configuration of metrics to be collected are stored in the metrics_definitions folder.
These are defined in yaml files according to the format handled by the flp confgenerator.
The flp confgenerator was modified to produce output that can be easily consumed by the NetObserv Operator.
The flp confgenerator was further modified so that it may be called as a module, and provides its output as a data structure returned from a function rather than a yaml file.
All metrics that may be produced are included in the metrics_definitions library, and they are associated with tags.
A parameter is added to the Operator CRD to specify tags of metrics to not produce.

On each iteration of the Operator, the Operator checks whether the CRD has been modified.
If the CRD has changed, the Operator reconciles the state of the cluster to the specification in the CRD.

The implementation of the Operator specifies the flp Network Transform enrichment (in particular, kubernetes features).
The actual metrics to produce are taken from the metrics_definitions, based on the enrichment defined in the Operator.
The Operator allocates the extract_aggregate and encode_prom Stage structures for the flp pipeline,
and extract_aggregate and encode_prom entries are filled in using the results from the confgenerator.
The configuration is placed into a configMap.
Flp is then deployed using this combined configuration.
The configuration is not changed during runtime.
In order to change the configuration (e.g. exclude a different set of metrics), flp must be redeployed.

Note that there are 2 data paths in flp. Data that is ingested is enriched and is then passed directly to Loki.
In addition, after the enrichment, we derive metrics (from the metrics_definitions), aggregate them, and report to prometheus.
The metrics_definitions does not impact the data that is sent to Loki.

In the metrics_definitions yaml files, there are tags associated with each metric.
A user may specify to skip metrics that have a particular tag.
This is specified by a field in the CRD.
These tags are then specified to the confgenerator module to produce metrics that are not associated with the specified tag.

## Parameters added to CRD to support metrics
Note: These parameters may be changed between interations, in which case the Operator redeploys flp.
- ignoreMetrics (list of tags to specify which metrics to ignore)


