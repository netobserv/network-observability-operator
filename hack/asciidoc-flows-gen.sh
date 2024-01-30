#!/bin/sh

ADOC=docs/flows-format.adoc
SOURCE=controllers/consoleplugin/config/static-frontend-config.yaml

# Header
echo "// Automatically generated by '$0'. Do not edit, or make the NETOBSERV team aware of the editions." > $ADOC
cat ./hack/flows-format-header.adoc >> $ADOC
echo -e "\n" >> $ADOC

echo -e '[cols="1,1,3,1,1",options="header"]' >> $ADOC
echo -e '|===' >> $ADOC
echo -e '| Name | Type | Description | Filter ID | Loki label' >> $ADOC

nbfields=$(yq '.fields | length' $SOURCE)

for i in $(seq 0 $(( $nbfields-1 )) ); do
  entry=$(yq ".fields | sort_by(.name) | .[$i]" $SOURCE)
  name=$(printf "$entry" | yq ".name")
  type=$(printf "$entry" | yq ".type")
  desc=$(printf "$entry" | yq ".description")
  filter=$(printf "$entry" | yq ".filter")
  if [[ "$filter" == "null" ]]; then
    filter=$(yq ".columns[] | select(.field == \"$name\").filter" $SOURCE | sed 's/null//')
    if [[ "$filter" == "" ]]; then
      filter="n/a"
    else
      filter="\`$filter\`"
    fi
  fi
  isLabel=$(printf "$entry" | yq ".lokiLabel")
  if [[ $isLabel == "true" ]]; then
    isLabel="yes"
  else
    isLabel="no"
  fi
  echo -e "| \`$name\`" >> $ADOC
  echo -e "| $type" >> $ADOC
  echo -e "| $desc" >> $ADOC
  echo -e "| $filter" >> $ADOC
  echo -e "| $isLabel" >> $ADOC
done

echo -e '|===' >> $ADOC
