#!/bin/sh

# Make sure all fields used in static frontend config are also defined in fields.yaml

FIELDS=docs/fields.yaml
CONFIG=controllers/consoleplugin/config/static-frontend-config.yaml

nbcols=$(yq '.columns | length' $CONFIG)

missing=()

check_field() {
  local name=$1
  if [[ $(yq ".fields[] | select(.name==\"$name\")" $FIELDS) == "" ]]; then
    missing+=($name)
  fi
}

for i in $(seq 0 $(( $nbcols-1 )) ); do
  entry=$(yq ".columns[$i]" $CONFIG)
  field=$(printf "$entry" | yq ".field")
  if [[ "$field" != "null" ]]; then
    check_field $field
  else
    fields=$(printf "$entry" | yq ".fields")
    if [[ "$fields" != "null" ]]; then
      n=$(printf "$entry" | yq ".fields | length")
      for f in $(seq 0 $(( $n-1 )) ); do
        check_field $(printf "$entry" | yq ".fields[$f]")
      done
    fi
  fi
done

if (( ${#missing[@]} != 0 )); then
  missing=$(printf ", %s" "${missing[@]}")
  echo "Missing fields: ${missing:2}"
  exit -1
fi
