No external API integration: Phase 28 adds zero new dependencies and integrates no external API, SDK, or
service — every capability (the walk/write split, mtime capture, the diff algorithm, the resolution
writes, remove-from-library) is satisfied by Go stdlib or by primitives already present and tested in this
repository, verified against `go.mod` and `frontend/package.json` in `28-RESEARCH.md`'s Package Legitimacy
Audit. The one out-of-repo surface this phase touches, GitHub Actions (`build.yml`), is existing CI
infrastructure this phase exercises rather than integrates against.
