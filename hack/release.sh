#!/bin/bash

# NOTE: this release script is provided for help purpose, but you should carefully read it and understand what it does before running it.
# Appropriate permissions will be required when pushing git tags / images
# Adapt the variables below to match your local files, desired version and remote target.
# Part of this workflow will be automated soon.

# Make sure also all repo have the correct HEAD & clean state

path_noo="../network-observability-operator"
path_gfk="../flowlogs-pipeline"
path_plg="../network-observability-console-plugin"
changelogs="../changelogs"
version="0.1.1"
since_date="2022-01-28 08:58:11"
# since-tag=v0.1.0

user=netobserv
remote=upstream

vv=v$version

cd $path_plg && \
 TAG="$vv" IMG_USER="$user" make image push && \
 git tag -a "$vv" -m "$vv" && \
 git push $remote --tags && \
 github_changelog_generator -u netobserv -p network-observability-console-plugin --since-commit "$since_date" --future-release "$vv" -o "$changelogs/plg.md" && \
#  github_changelog_generator -u netobserv -p network-observability-console-plugin --since-tag "$since-tag" --future-release "$vv" -o "$changelogs/plg.md" && \
 sed -i 's/\(NETOBSERV-[0-9]\+\)/[\1](https:\/\/issues.redhat.com\/browse\/\1)/g' "$changelogs/plg.md"

cd $path_gfk && \
 VERSION="$vv" USER="$user" make image push && \
 git tag -a "$vv" -m "$vv" && \
 git push $remote --tags && \
 github_changelog_generator -u netobserv -p goflow2-kube-enricher --since-commit "$since_date" --future-release "$vv" -o "$changelogs/gfk.md" && \
#  github_changelog_generator -u netobserv -p goflow2-kube-enricher --since-tag "$since-tag" --future-release "$vv" -o "$changelogs/gfk.md" && \
 sed -i 's/\(NETOBSERV-[0-9]\+\)/[\1](https:\/\/issues.redhat.com\/browse\/\1)/g' "$changelogs/gfk.md"

cd $path_noo && \
 VERSION="$version" IMAGE_TAG_BASE="quay.io/$user/network-observability-operator" make image-push bundle-push && \
 git tag -a "$vv" -m "$vv" && \
 git push $remote --tags && \
 github_changelog_generator -u netobserv -p network-observability-operator --since-commit "$since_date" --future-release "$vv" -o "$changelogs/noo.md" && \
#  github_changelog_generator -u netobserv -p network-observability-operator --since-tag "$since-tag" --future-release "$vv" -o "$changelogs/noo.md" && \
 sed -i 's/\(NETOBSERV-[0-9]\+\)/[\1](https:\/\/issues.redhat.com\/browse\/\1)/g' "$changelogs/noo.md"

# Then in github, create release from tag and paste the generated changelogs in all three repos
