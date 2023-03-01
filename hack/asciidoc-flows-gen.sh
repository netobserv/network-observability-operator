#!/bin/sh

ADOC=docs/flows-format.adoc
MD_SOURCE=https://raw.githubusercontent.com/netobserv/network-observability-console-plugin/main/web/docs

# Header
echo "// Automatically generated by '$0'. Do not edit." > $ADOC
cat ./hack/flows-format-header.adoc >> $ADOC
echo -e "\n" >> $ADOC

# Labels
kramdoc -o - <(curl -fsSL ${MD_SOURCE}/interfaces/Labels.md) \
  | sed -r 's/^= /== /' \
  | sed -r 's/Interface: //' \
  | sed -r '/Properties/d' \
  | sed -r 's~xref:.*FlowDirection\.adoc\[`FlowDirection`\]~<<Enumeration: FlowDirection,`FlowDirection`>>~' >> $ADOC

# Fields
echo -e "\n" >> $ADOC
kramdoc -o - <(curl -fsSL ${MD_SOURCE}/interfaces/Fields.md) \
  | sed -r 's/^= /== /' \
  | sed -r 's/Interface: //' \
  | sed -r '/Properties/d' >> $ADOC

# FlowDirection enum
echo -e "\n" >> $ADOC
kramdoc -o - <(curl -fsSL ${MD_SOURCE}/enums/FlowDirection.md) \
  | sed -r 's/^= /== /' \
  | sed -r '/Enumeration Members/d' >> $ADOC

sed -i -r "s/^=== (.+)/\1::/" $ADOC
