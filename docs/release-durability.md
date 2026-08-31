# Release durability record

The first protocol release was intentionally preserved as a non-durable
historical record after the repository's immutable-release setting was
enabled.

```text
v0.1.0
release_id: 380039937
release_api_immutable: false
classification: NON_DURABLE_RELEASE
tag_ref_object_type: commit
tag_target: 6ceb4e01f2b0adc1de913112584fb20a4fb6393d
```

The `v0.1.0` tag, release, and assets are not deleted or rewritten. The
repository ruleset is an additional tag guard, but it is not treated as a
substitute for GitHub's immutable-release API state.

The official repository endpoint
`PUT /repos/{owner}/{repo}/immutable-releases` was enabled before the next
release. The corrective release workflow creates a draft, uploads all assets,
publishes once, then fails closed unless the release REST response says
`immutable=true`, the tag is an annotated tag whose object targets the exact
merge commit, and GitHub's asset IDs, sizes, and SHA-256 digests match the
locally computed values in the same Actions run.

`v0.1.1` was published after immutable releases were enabled and is durable
according to the release REST response, even though run `33438082755` was
recorded as failed after the exact checks because its final summary formatter
indexed the already-extracted asset array as `.assets[]`. That tag, release,
and assets are preserved unchanged and classified as `DURABLE_RELEASE_WITH_FAILED_REPORT_STEP`.

The summary formatter was corrected in the next PR. Since `v0.1.1` already
existed, the next unused version `v0.1.2` was issued without modifying the
older tag or release:

```text
v0.1.2
release_id: 380045288
release_api_immutable: true
publish_and_verify_run: 33438395905
tag_ref_object_type: tag
annotated_tag_object: cc311eb5b14f54b5e467f9402c6f7a32e9b5b3dc
tag_target: 5534a6814fdd9c8c6ddd957d5c26fd5c15258e02
asset_1: 538516119 gooo-foundation-rotation-evidence-v0.1.2.tar.gz 4767 sha256:d259722c20cb3e525575d4bfb1488424ab4d23eed964fe68334345711635fe6d
asset_2: 538516120 gooo-foundation-rotation-linux-amd64 4604245 sha256:82bd491f106abbe729242bcb0bb8b5a47d8dd5d9ceaa35be73722e4fc90d6b36
verification_artifact_checks: release_immutable annotated_tag asset_digests_verified exact_asset_count
```

The publish-and-verify run completed successfully, and its durable-release
verification artifact records all four checks as true. Live issuer and
`meta-ontology-go` PR #619 integration remain `UNKNOWN`.
