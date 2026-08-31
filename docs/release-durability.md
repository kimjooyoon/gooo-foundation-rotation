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

The summary formatter is corrected in the next PR. Since `v0.1.1` already
exists, the next unused version `v0.1.2` will be the first release whose
publish-and-verify Actions run is fully green. Live issuer and
`meta-ontology-go` PR #619 integration remain `UNKNOWN`.
