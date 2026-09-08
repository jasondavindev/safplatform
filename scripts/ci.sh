#!/bin/bash
set -aexuo pipefail

app_name="$(echo "${GITHUB_REPOSITORY}" | cut -d / -f 2)"

APP_NAME="${app_name}"
IMAGE_NAME="${CONTAINER_REGISTRY}/${app_name}"
IMAGE_TAG="${GITHUB_SHA}"
FULL_IMAGE_NAME=${IMAGE_NAME}:${IMAGE_TAG}

set +x
echo "${REPO_PASSWORD}" | docker login "${CONTAINER_REGISTRY%%/*}" -u "${REPO_USER}" --password-stdin
set -x

builder="${APP_NAME}-ci"
if ! docker buildx inspect "${builder}" >/dev/null 2>&1; then
    docker buildx create --name "${builder}" --driver docker-container --bootstrap >/dev/null
fi

docker buildx build \
    --builder "${builder}" \
    --cache-from "type=registry,ref=${IMAGE_NAME}:cache" \
    --cache-to "type=registry,ref=${IMAGE_NAME}:cache,mode=max,ignore-error=true" \
    --tag "${FULL_IMAGE_NAME}" \
    --push \
    .

if [[ -n "${GITHUB_ENV:-}" ]]; then
    {
        echo "APP_NAME=${APP_NAME}"
        echo "IMAGE_NAME=${IMAGE_NAME}"
        echo "IMAGE_TAG=${IMAGE_TAG}"
    } >>"${GITHUB_ENV}"
fi

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
    {
        echo "full_image_name=${FULL_IMAGE_NAME}"
        echo "image_tag=${IMAGE_TAG}"
    } >>"${GITHUB_OUTPUT}"
fi
