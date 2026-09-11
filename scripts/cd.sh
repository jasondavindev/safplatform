#!/bin/bash
set -exuo pipefail

app_name="$(echo "${GITHUB_REPOSITORY}" | cut -d / -f 2)"
values_file=".idp/${app_name}/values.yaml"
image_tag="${IMAGE_TAG:-}"

if [[ ! -f "${values_file}" ]]; then
    echo "values file not found: ${values_file}"
    exit 1
fi

if [[ -z "${image_tag}" ]]; then
    echo "Invalid image tag"
    exit 1
fi

IMAGE_TAG_VALUE="${image_tag}" yq -i \
    '.global.image.tag = strenv(IMAGE_TAG_VALUE) | .image.tag style="double"' \
    "${values_file}"

if git diff --quiet -- "${values_file}"; then
    echo "values.yaml is up to date ${image_tag}"
    exit 0
fi

git config user.name "github-actions[bot]"
git config user.email "github-actions[bot]@users.noreply.github.com"

git add "${values_file}"
git commit -m "chore(cd): bump ${app_name} image tag to ${image_tag} [skip ci]"

branch="${GITHUB_REF_NAME:-$(git rev-parse --abbrev-ref HEAD)}"

git pull --rebase origin "${branch}"

if git push --force origin "HEAD:${branch}"; then
    exit 0
fi

exit 1
