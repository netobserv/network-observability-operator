#!/usr/bin/env bash
set -euo pipefail

version="$1"

if [[ $version == "" ]]; then
  echo "Missing CRD version."
  exit 1
fi

crd="./config/crd/bases/flows.netobserv.io_flowcollectors.yaml"
csv="./config/csv/bases/transformed-csv.yaml"

crdRoot=".spec.versions[] | select(.name==\"$version\").schema.openAPIV3Schema.properties.spec.properties"
csvRoot=".spec.customresourcedefinitions.owned[] | select(.version==\"$version\").specDescriptors"

# Load the whole CRD property tree and CSV spec-descriptors once, instead of re-parsing
# both files on every node of the recursive walk (previously thousands of `yq` invocations).
crdJsonFile=$(mktemp)
csvJsonFile=$(mktemp)
trap 'rm -f "$crdJsonFile" "$csvJsonFile"' EXIT
./bin/yq -o=json e "$crdRoot" "$crd" >"$crdJsonFile"
./bin/yq -o=json e "$csvRoot" "$csv" >"$csvJsonFile"

mutations=$(jq -n -r --slurpfile crdArr "$crdJsonFile" --slurpfile csvArr "$csvJsonFile" '
  $crdArr[0] as $crd | $csvArr[0] as $csv |
  def dispName($path):
    ($path | sub(".*\\."; "")) as $last
    | ($last | gsub("(?<a>[a-z])(?<b>[A-Z])"; "\(.a) \(.b)") | ascii_downcase) as $lower
    | ($lower[0:1] | ascii_upcase) + $lower[1:];

  def ignored($path):
    ($path | test("utoscaler$"))
    or ($path | test("\\.resources$"))
    or ($path | test("\\.tls$"))
    or ($path | test("\\.statusTls$"));

  def walkProps($node; $prefix):
    ($node // {}) | to_entries[] |
    ($prefix + (if $prefix == "" then "" else "." end) + .key) as $path |
    if ignored($path) then empty
    else
      .value as $schema
      | (first($csv[] | select(.path == $path)) // null) as $entry
      | (($entry["x-descriptors"] // []) | last // "" | test(":hidden$")) as $hidden
      | if $hidden then empty
        else
          (if $entry == null then
            {type: "add", path: $path, displayName: dispName($path)}
          elif ($entry | has("displayName") | not) then
            {type: "set", path: $path, displayName: dispName($path)}
          else empty
          end),
          (if $schema.type == "object" and ($schema | has("properties")) then
            walkProps($schema.properties; $path)
          else empty
          end)
        end
    end;

  walkProps($crd; "") | [.type, .path, .displayName] | @tsv
')

exprParts=()
while IFS=$'\t' read -r mtype mpath mdisp; do
  [[ -z $mtype ]] && continue
  if [[ $mtype == "add" ]]; then
    exprParts+=("($csvRoot) += {\"path\":\"$mpath\",\"displayName\":\"$mdisp\"}")
  else
    exprParts+=("($csvRoot[] | select(.path==\"$mpath\").displayName) = \"$mdisp\"")
  fi
done <<<"$mutations"

if [[ ${#exprParts[@]} -gt 0 ]]; then
  bigExpr="${exprParts[0]}"
  for part in "${exprParts[@]:1}"; do
    bigExpr="$bigExpr | $part"
  done
  ./bin/yq -i "$bigExpr" "$csv"
fi
