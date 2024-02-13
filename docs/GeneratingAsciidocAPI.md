# Generating AsciiDoc API reference

## Setup docsgen repo

1. Clone https://github.com/jboxman-rh/openshift-apidocs-gen
2. run `npm install -g`

## Run it

The doc generator needs to talk with a running cluster, with the desired CRDs installed. It doesn't require anything fancy, you can run KIND (`kind create cluster`).

1. Deploy netobserv CRD on your cluster (e.g. `make generate install`).
2. run `hack/asciidoc-gen.sh`

# Generate AsciiDoc for flows JSON format reference

## Run the script

```bash
hack/asciidoc-flows-gen.sh 
```
