# Generating AsciiDoc API reference

## Setup docsgen repo

1. Clone https://github.com/jboxman-rh/openshift-apidocs-gen
2. run `npm install -g`

## Run it

The doc generator needs to talk with a running cluster, with the desired CRDs installed.

1. Deploy netobserv CRD on your cluster (e.g. deploy just the operator; no need to install everything, just the CRD is needed).
2. run `hack/asciidoc-gen.sh`

# Generate AsciiDoc for flows JSON format reference

## Run the script

```bash
hack/asciidoc-flows-gen.sh 
```
