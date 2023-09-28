#! /usr/bin/env bash

# set -o errexit
# set -o pipefail

release_tag="${RELEASE_TAG}"
pre_release=${PRE_RELEASE}
overwrite_release_assets=${OVERWRITE_RELEASE_ASSETS}
helm_merge_with_existing_indexes=${HELM_MERGE_WITH_EXISTING_INDEXES}
release_title="${RELEASE_TITLE}"

github_repo_owner="${GITHUB_REPO_OWNER}"
github_repo_name="${GITHUB_REPO_NAME}"
github_repository="${github_repo_owner}/${github_repo_name}"

opts=("-R" "${github_repository}")

release=$(gh release list ${opts[@]} | tail -n +1 | (grep -iE "\s+$release_tag\s+" || echo -n "") | awk '{print $3}')
if [[ -z $release ]]; then
	echo "going to create release, as RELEASE ($release_tag) does not exist"
	createOpts="${opts[@]}"
	if $pre_release; then
		createOpts+=("--prerelease")
	fi
	if ! [[ -z $release_title ]]; then
		createOpts+=("--title" "'$release_title'")
	fi
	createOpts+=("--notes" "'$RELEASE_NOTES'")

	echo "creating github release with cmd: \`gh release create $release_tag ${createOpts[@]}\` "
	eval gh release create "$release_tag" ${createOpts[@]} --generate-notes
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

echo "uploading packaged helm-charts with cmd: \`gh release upload $release_tag ${uploadOpts[*]} $tar_dir/*.tgz\`"
gh release upload "$release_tag" ${uploadOpts[@]} $tar_dir/*.tgz

## updating CRDs
rm -rf crds/crds-all.yml
for file in crds/*; do
	cat "$file" >>crds/crds-all.yml
	echo "---" >>crds/crds-all.yml
done

gh release upload "$release_tag" ${uploadOpts[@]} crds/*.yml

if $helm_merge_with_existing_indexes; then
	# remove entries related to the current release_tag, for all the charts
	index_file_url="https://${github_repo_owner}.github.io/${github_repo_name}/index.yaml"
	curl -f -L0 "$index_file_url" >$tar_dir/index.yaml
	echo "+++++++ current: index.yaml"
	cat $tar_dir/index.yaml

	cat $tar_dir/index.yaml | yq '. | (
    .entries = (
      .entries | map_values([
                    .[] | select(
                      (. != null) and (.appVersion != env.RELEASE_TAG)
                    )
                  ])
    )
  )' -y >$tar_dir/old-index.yaml
fi

# helm repo index --debug $tar_dir --url "https://github.com/${github_repository}/releases/download/${release_tag}" --merge $tar_dir/index.yaml
helm repo index --debug $tar_dir --url "https://github.com/${github_repository}/releases/download/${release_tag}"

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
gh release upload "$release_tag" ${uploadOpts[@]} .static-pages/index.yaml

cat >.static-pages/index.html <<EOF
<html>
  <head>
    <title>Kloudlite Helm Charts</title>
  </head>

  <body>
    <p>Hi, the file you are looking for is located <a href="${index_file_url}">here</a>.</p>
  </body>
</html>
EOF

gh release upload "$release_tag" ${uploadOpts[@]} .static-pages/index.html
