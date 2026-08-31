# Bootstrap record

This repository was created as a new public repository on `main` with one
direct bootstrap commit. The bootstrap commit intentionally predates the
conformance workflow.

```text
repository: kimjooyoon/gooo-foundation-rotation
default_branch: main
bootstrap_mode: direct-main
bootstrap_direct_main: 2
post_ci_bootstrap_direct_main: 0
bootstrap_workflow_present: false
workflow_correction_count: 1
correction: after PR #1 reported no checks, add the minimal CI bootstrap
           workflow directly on main before retrying the functional PR
```

The correction count is recorded explicitly. No bootstrap workflow is
retroactively claimed, and no failed tag or release is deleted or rewritten.
The missing-base-workflow condition is retained as a future
`repository-bootstrap` capability gap.
