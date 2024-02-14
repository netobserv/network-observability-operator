# Generating AsciiDoc API reference

## One-time setup docsgen repo

1. Clone https://github.com/jboxman-rh/openshift-apidocs-gen
2. run `npm install -g`

## Run it

```bash
# If you haven't already, start any k8s cluster
kind create cluster

# Make sure the CRD has all the desired doc within
make generate

make install
hack/asciidoc-gen.sh
```

# Generate AsciiDoc for flows JSON format reference

## Run the script

```bash
hack/asciidoc-flows-gen.sh 
```
