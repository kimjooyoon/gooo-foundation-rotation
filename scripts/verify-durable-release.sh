#!/usr/bin/env bash
set -Eeuo pipefail

repo="${RELEASE_REPOSITORY:?RELEASE_REPOSITORY is required}"
tag="${RELEASE_TAG:?RELEASE_TAG is required}"
expected_commit="${RELEASE_COMMIT:?RELEASE_COMMIT is required}"
binary_name="${RELEASE_ASSET_BINARY:?RELEASE_ASSET_BINARY is required}"
evidence_name="${RELEASE_ASSET_EVIDENCE:?RELEASE_ASSET_EVIDENCE is required}"

release="$(gh api "repos/${repo}/releases/tags/${tag}")"
jq -e --arg tag "${tag}" '.immutable == true and .draft == false and .prerelease == false and .tag_name == $tag' <<<"${release}" >/dev/null
release_id="$(jq -r '.id' <<<"${release}")"
test "${release_id}" -gt 0

tag_ref="$(gh api "repos/${repo}/git/ref/tags/${tag}")"
jq -e '.object.type == "tag" and (.object.sha | length) == 40' <<<"${tag_ref}" >/dev/null
tag_object_sha="$(jq -r '.object.sha' <<<"${tag_ref}")"
tag_object="$(gh api "repos/${repo}/git/tags/${tag_object_sha}")"
jq -e --arg commit "${expected_commit}" '.object.type == "commit" and .object.sha == $commit and .tag != ""' <<<"${tag_object}" >/dev/null

assets="$(jq -c '.assets' <<<"${release}")"
test "$(jq 'length' <<<"${assets}")" -eq 2
binary_size="$(stat -c '%s' "${binary_name}")"
binary_digest="sha256:$(sha256sum "${binary_name}" | awk '{print $1}')"
evidence_size="$(stat -c '%s' "${evidence_name}")"
evidence_digest="sha256:$(sha256sum "${evidence_name}" | awk '{print $1}')"
jq -e --arg name "${binary_name}" --arg digest "${binary_digest}" --argjson size "${binary_size}" \
  'any(.[]; .name == $name and .size == $size and .digest == $digest and (.id | type) == "number" and .id > 0)' <<<"${assets}" >/dev/null
jq -e --arg name "${evidence_name}" --arg digest "${evidence_digest}" --argjson size "${evidence_size}" \
  'any(.[]; .name == $name and .size == $size and .digest == $digest and (.id | type) == "number" and .id > 0)' <<<"${assets}" >/dev/null

jq -S -n \
  --arg schema 'gooo/foundation-authorization/durable-release-verification/v1' \
  --arg repository "${repo}" --arg tag "${tag}" --arg release_id "${release_id}" \
  --arg tag_object_sha "${tag_object_sha}" --arg commit "${expected_commit}" \
  --argjson assets "${assets}" \
  '{schema:$schema,repository:$repository,tag:$tag,release_id:($release_id|tonumber),immutable:true,tag_ref:{object_type:"tag",object_sha:$tag_object_sha},annotated_tag:{target_type:"commit",target_sha:$commit},assets:$assets,checks:{release_immutable:true,annotated_tag:true,exact_asset_count:($assets|length==2),asset_digests_verified:true}}' > durable-release-verification.json

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "# Durable release verification"
    echo "decision=CLOSED"
    echo "release_immutable=true"
    echo "tag=${tag} annotated_target=${expected_commit}"
    jq -r '.[] | "asset_id=\(.id) name=\(.name) size=\(.size) digest=\(.digest)"' <<<"${assets}"
  } >> "${GITHUB_STEP_SUMMARY}"
fi
echo "durable release verification passed: ${tag} release_id=${release_id}"
