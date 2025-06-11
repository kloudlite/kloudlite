#! /usr/bin/env bash

# set -o errexit
# set -o pipefail

chart_version="${CHART_VERSION}"
pre_release=${PRE_RELEASE}
overwrite_release_assets=${OVERWRITE_RELEASE_ASSETS}
helm_merge_with_existing_indexes=${HELM_MERGE_WITH_EXISTING_INDEXES}
release_title="${RELEASE_TITLE}"

github_repo_owner="${GITHUB_REPO_OWNER}"
github_repo_name="${GITHUB_REPO_NAME}"
github_repository="${github_repo_owner}/${github_repo_name}"

opts=("-R" "${github_repository}")

echo "$chart_version" | grep -i '\-nightly$'
if [ $? -eq 0 ]; then
  overwrite_release_assets=true
fi

release=$(gh release list ${opts[@]} | tail -n +1 | (grep -iE "\s+$chart_version\s+" || echo -n "") | awk '{print $3}')
if [[ -z $release ]]; then
  echo "going to create release, as RELEASE ($chart_version) does not exist"
  createOpts="${opts[@]}"
  if $pre_release; then
    createOpts+=("--prerelease")
  fi
  if ! [[ -z $release_title ]]; then
    createOpts+=("--title" "'$release_title'")
  fi
  createOpts+=("--notes" "'$RELEASE_NOTES'")

  echo "creating github release with cmd: \`gh release create $chart_version ${createOpts[@]}\` "
  eval gh release create "$chart_version" ${createOpts[@]} --generate-notes
else
  echo "release $release exists, going to build charts, now"
fi

tar_dir=".chart-releases"

# for dir in $(ls -d charts/*); do
for dir in charts/*/; do
  echo cr package "$dir" --package-path $tar_dir
  cr package "$dir" --package-path $tar_dir
done

uploadOpts="${opts[@]}"
if $overwrite_release_assets; then
  uploadOpts+=("--clobber")
fi

echo "uploading packaged helm-charts with cmd: \`gh release upload $chart_version ${uploadOpts[*]} $tar_dir/*.tgz\`"
gh release upload "$chart_version" ${uploadOpts[@]} $tar_dir/*.tgz

## updating CRDs
rm -rf crds/crds-all.yml
for file in crds/*; do
  cat "$file" >>crds/crds-all.yml
  echo "---" >>crds/crds-all.yml
done

gh release upload "$chart_version" ${uploadOpts[@]} crds/*.yml

# remove entries related to the current release_tag, for all the charts
index_file_url="https://${github_repo_owner}.github.io/${github_repo_name}/index.yaml"
curl -f -L0 "$index_file_url" >$tar_dir/index.yaml
echo "+++++++ current: index.yaml"
cat $tar_dir/index.yaml

cat $tar_dir/index.yaml | yq '. | (
  .entries = (
    .entries | map_values([
                  .[] | select(
                    (. != null) and (.version != env.CHART_VERSION)
                  )
                ])
  )
)' -y >$tar_dir/old-index.yaml

# helm repo index --debug $tar_dir --url "https://github.com/${github_repository}/releases/download/${release_tag}" --merge $tar_dir/index.yaml
helm repo index --debug $tar_dir --url "https://github.com/${github_repository}/releases/download/${chart_version}"

cp $tar_dir/index.yaml $tar_dir/new-index.yaml
keys=$(cat $tar_dir/index.yaml | yq '.entries | keys |.[]' -r)
pushd $tar_dir
for key in $keys; do
  echo "merging for chart: $key"
  cat index.yaml | yq '.entries[$key] = ($entries|fromjson)' --arg key "$key" --arg entries "$(yq -s '.' index.yaml old-index.yaml | jq --arg key "$key" '(.[0].entries[$key] + .[1].entries[$key])' -r)" -y >new-index.yaml
  mv new-index.yaml index.yaml
done
popd

echo "+++++++ new: index.yaml"
cat $tar_dir/index.yaml

mkdir -p .static-pages
cp $tar_dir/index.yaml .static-pages/index.yaml
gh release upload "$chart_version" ${uploadOpts[@]} .static-pages/index.yaml

cat >.static-pages/index.html <<EOF
<html>
  <head>
    <title>Kloudlite Helm Charts - $chart_version </title>
    <meta name="revised" content="$(date -Iseconds)" />
  </head>

  <body>
    <p>Hi, the file you are looking for is located <a href="${index_file_url}">here</a>.</p>
  </body>
</html>
EOF

gh release upload "$chart_version" ${uploadOpts[@]} .static-pages/index.html
