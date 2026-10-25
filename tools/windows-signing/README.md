# Windows runtime signing

Ported under the repository MIT license from `z189yis/engram-cjk` commit
`632a97122a3c9b15f9ed10de476d105b8ed119b9`. This downstream pipeline changes the
runtime repository identity to `rongxinzy/engram`, uses Go 1.26.10 and pins
external Actions to reviewed commits. Upstream attribution remains intact.

Certum credentials remain exclusively in the central signing repository,
`rongxinzy/RongxinAI`. This repository neither stores those credentials nor
authenticates to SimplySign. Tagged builds upload unsigned Actions artifacts;
final centrally signed assets are published in this original repository.

Configure the repository secret `RUNTIME_ARTIFACT_READ_TOKEN` to read Actions
artifacts and source provenance in the central and runtime repositories.
Protect the `release` environment with main-only deployment rules.
No public `RUNTIME_SIGNER_THUMBPRINT` variable is required.

The upstream `Release` workflow excludes `v*-zhiyuan.*` tags, which are owned
only by `Fork Release`; they do not invoke the upstream Pi or Homebrew publisher.
All existing pull-request CI remains active. Build and publication both use
GoReleaser 2.18.2 on Windows so the authenticated unsigned bytes can be compared
with the publication rebuild before replacement. The namespace fix in #2 must
be merged before creating the first downstream release tag.

After a successful tagged build, dispatch the central signing workflow from
main with its build run ID. Then dispatch this repository's release workflow
from main with the successful central `signing_run_id`.

Publication checks workflow origins, main ancestry, unchanged tag identity,
manifest source identity, and SHA-256 hashes. It does not perform Authenticode,
certificate publisher, trust-chain, EKU, or timestamp verification. This is an
explicit policy choice; hashes verify byte integrity, not publisher identity.
Central signing still fails if the signing command fails.

Artifacts expire after 14 days. Expired inputs require a new build/sign run.
Never replace assets under a published tag. See the
[central operational guide](https://github.com/rongxinzy/RongxinAI/blob/main/scripts/runtime-signing/README.md).

Desktop consumers update runtime pins and SHA-256 values together. Their
download, package, and install checks retain integrity and execution checks,
but do not require runtime signature verification.

Run publication policy tests without cloud credentials:

```bash
go test ./tools/windows-signing
```
