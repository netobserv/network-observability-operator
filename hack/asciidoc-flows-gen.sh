#!/bin/sh

ADOC=docs/flows-format.adoc
FE_SOURCE=controllers/consoleplugin/config/static-frontend-config.yaml
LOKI_LABEL_SOURCE=pkg/helper/loki/loki-labels.json
CARDINALITY_SOURCE=pkg/helper/cardinality/cardinality.json
OTEL_SOURCE=pkg/helper/otel/otel-config.json

# Header
echo "// Automatically generated by '$0'. Do not edit, or make the NETOBSERV team aware of the editions." > $ADOC
cat ./hack/flows-format-header.adoc >> $ADOC
echo -e "\n" >> $ADOC

echo -e '[cols="1,1,3,1,1,1,1",options="header"]' >> $ADOC
echo -e '|===' >> $ADOC
echo -e '| Name | Type | Description | Filter ID | Loki label | Cardinality | OpenTelemetry' >> $ADOC

nbfields=$(yq '.fields | length' $FE_SOURCE)
lokiLabels=$(cat $LOKI_LABEL_SOURCE)
cardinalityMap=$(cat $CARDINALITY_SOURCE)
otelMap=$(cat $OTEL_SOURCE)
errors=""

for i in $(seq 0 $(( $nbfields-1 )) ); do
  frontEntry=$(yq ".fields | sort_by(.name) | .[$i]" $FE_SOURCE)
  name=$(printf "$frontEntry" | yq ".name")
  type=$(printf "$frontEntry" | yq ".docType")
  if [[ "$type" == "null" ]]; then
    type=$(printf "$frontEntry" | yq ".type")
  fi
  desc=$(printf "$frontEntry" | yq ".description")
  filter=$(printf "$frontEntry" | yq ".filter")
  if [[ "$filter" == "null" ]]; then
    filter=$(yq ".columns[] | select(.field == \"$name\").filter" $FE_SOURCE | sed 's/null//')
    if [[ "$filter" == "" ]]; then
      filter="n/a"
    else
      filter="\`$filter\`"
    fi
  else 
    filter="\`$filter\`"
  fi
  isLabel=$(printf "$lokiLabels" | jq "map(select(any(match(\"^$name$\"))))")
  if [[ $isLabel == "[]" ]]; then
    isLabel="no"
  else
    isLabel="yes"
  fi
  cardWarn=$(printf "$cardinalityMap" | jq -r ".$name")
  if [[ "$cardWarn" == "null" ]]; then
      errors="$errors\nmissing cardinality for field $name; check cardinality.json"
  fi
  otel=$(printf "$otelMap" | jq -r ".$name")
  if [[ "$otel" == "null" ]]; then
      otel="n/a"
  fi
  echo -e "| \`$name\`" >> $ADOC
  echo -e "| $type" >> $ADOC
  echo -e "| $desc" >> $ADOC
  echo -e "| $filter" >> $ADOC
  echo -e "| $isLabel" >> $ADOC
  echo -e "| $cardWarn" >> $ADOC
  echo -e "| $otel" >> $ADOC
done

echo -e '|===' >> $ADOC

if [[ $errors != "" ]]; then
  echo -e $errors
  exit 1
fi
