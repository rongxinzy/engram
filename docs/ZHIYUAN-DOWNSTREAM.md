# ZhiYuan downstream maintenance

This fork tracks official Engram. Its initial maintenance base is upstream
`71929313f1fea3f4d4b5178ab95f482525c4ad9f`, following release 3.6.0. Engram's
MIT license and upstream attribution remain unchanged. This is not a replacement
agent engine; the application uses Engram's local HTTP storage API.

## Exact memory owners

ZhiYuan Agent already stores personal and companion memory under opaque owner
identifiers. Upstream accepts these on writes but rejects them during read
project resolution. The downstream validator additionally accepts exactly:

- `personal://zhiyuan-agent/user`
- `companion://zhiyuan-agent/<id>`, with one nonempty segment of at most 128 ASCII
  lowercase letters, digits, hyphens or underscores

These identifiers never resolve a filesystem path. They keep the existing
canonicalization and known-project check; an unknown owner still returns 404.
Other schemes/authorities, ordinary paths, nested segments, queries, fragments,
percent escapes, whitespace and control characters remain invalid. There is no
all-project fallback, stored-owner rewrite, enrollment change or auth change.
The application's own current-owner authorization remains required.

The initial regression suite exercises actual HTTP writes/search/recent for two
companions, personal memory and an ordinary workspace, plus invalid and unknown
identifiers. The same tests fail on the unchanged upstream main. This small
extension does not import the old CJK production implementation: the old fork's
26-query corpus, short/mixed query cases and a 10,000-observation latency baseline
already pass on official 3.6.0. Keep CJK regressions as acceptance evidence and
port old behavior only when a concrete gap is reproduced.

## Publication boundary

Maintenance and release work is tracked in this fork's issues and pull requests;
it does not imply an upstream contribution or upstream approval. Source commits,
tests and a successful native application run are required before publishing a
pinned runtime. A local build is not a release. Multi-platform archives and
checksums must come from the same reviewed source revision, and a downstream
binary must report its downstream version. Existing application integrity and
platform acceptance checks continue to apply.
