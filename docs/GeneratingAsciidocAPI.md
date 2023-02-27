# Generating AsciiDoc API reference

## Setup docsgen repo

1. Clone https://github.com/jboxman-rh/openshift-apidocs-gen
2. run `npm install -g`

## Run it

The doc generator needs to talk with a running cluster, with the desired CRDs installed.

1. Deploy netobserv CRD on your cluster (e.g. deploy just the operator; no need to install everything, just the CRD is needed).
2. run `hack/asciidoc-gen.sh`

# Generate AsciiDoc for flows JSON format reference

The flows JSON format is documented in the Console plugin repository (TODO: link), in markdown. For downstream consummption, we pull it from there, do some post-processing edition, and convert it to AsciiDoc using [Kramdoc](https://matthewsetter.com/technical-documentation/asciidoc/convert-markdown-to-asciidoc-with-kramdoc/).

## Install Kramdoc

```bash
gem install kramdown-asciidoc
```

## Run the script


From console plugin repo:

```bash
make generate-doc 
```

Then, from this repo (operator):

```bash
hack/asciidoc-flows-gen.sh 
```
