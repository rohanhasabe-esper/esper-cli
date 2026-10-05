# Esper CLI maintenance

This guide covers syncing the CLI's API definitions with Esper API docs and
adding CLI tools (commands). Run the CLI commands below from this repository's
root unless a different working directory is stated.

## 1. Sources and configuration

The CLI uses a reviewed, checked-in API bundle. It does not download API docs
or generate commands when an installed CLI runs.

| File or directory | Purpose |
| --- | --- |
| Private `github.com/esper-code/esper-api-docs` checkout | Upstream `openapi.yaml` and referenced API documentation files |
| `https://api.esper.io/page-data/shared/oas-openapi.yaml.json` | Deployed public OpenAPI snapshot, which determines the public operation boundary |
| `tools/specbundle/bundle.sh` | Bundles the upstream checkout, fetches or reads the public snapshot, and replaces the local bundle |
| `tools/specbundle/main.mjs` | Merges sources and overlays, filters public operations, partitions generations, and adds CLI annotations |
| `spec/openapi/*.yaml` | Resolved, JSON-formatted YAML consumed by Go code generation; also the default reviewed overlay during a refresh |
| `spec/openapi/manifest.json` | Public snapshot metadata, operation keys, counts, and exclusions |
| `spec/redocly.yaml` | Redocly lint configuration and generation roots |
| `internal/commandpolicy/policy.go` | Explicit CLI exclusions, replacements, and naming exceptions |
| `tools/codegen/main.go` | Generates the command metadata |
| `internal/cmd/generated/runner.go` | Builds Cobra commands and executes them using shared runtime behavior |
| `tools/skillgen/main.go` | Generates the agent guide in `SKILL.md` |

Syncing these files changes API coverage and command metadata. It does not
change a user's tenant credentials or active context. Those are managed through
`espercli configure` and `espercli context`.

## 2. Sync with esper-api-docs

### Prerequisites

- An authenticated checkout of `esper-code/esper-api-docs`.
- Go 1.23 or newer, Node.js, npm/npx, and curl.
- A reviewed upstream revision and public snapshot. Record the upstream commit
  and snapshot provenance in the change description.

Update the API-docs checkout using that repository's contribution instructions.
Inspect its changes before bringing them into this CLI. In that checkout, use
`git rev-parse HEAD` to record the exact revision; the CLI bundler does not
automatically record the checkout's Git revision in its manifest.

### Refresh the bundle

From the CLI repository root:

```sh
tools/specbundle/bundle.sh /absolute/path/to/esper-api-docs \
  https://api.esper.io/page-data/shared/oas-openapi.yaml.json
npx --yes @redocly/cli@1.34.5 lint --config spec/redocly.yaml
```

For a reproducible refresh, pass a saved public snapshot instead of its URL:

```sh
tools/specbundle/bundle.sh /absolute/path/to/esper-api-docs \
  /absolute/path/to/reviewed-public-oas.json
```

The script replaces `spec/openapi/*.yaml` and `manifest.json`. Its optional
third argument supplies a different canonical overlay directory; the default
is the existing `spec/openapi` directory.

**Review existing overlays carefully.** For overlapping operations, the current
merge preserves the existing operation definition when merging the public
snapshot, not just its `x-esper-*` annotations. A fresh public snapshot therefore
does not guarantee that an existing request or response schema was updated.
Compare changed upstream/public contracts with the overlay and reconcile stale
definitions through the maintained source or overlay-generation logic before
accepting the result. Do not assume that rerunning the bundler fixes drift.

Review the bundle diff for added/removed operations, request and response
schemas, required fields, authentication, pagination, and command annotations.
Check manifest public operation keys and exclusion reasons. Update
`spec/openapi/README.md` when its recorded source revision or source description
changes. If a refresh creates a new generation file, also register its root in
`spec/redocly.yaml` and check generator generation support.

### Regenerate commands and guidance

```sh
go run ./tools/codegen
go run ./tools/skillgen
```

Never manually edit `internal/cmd/generated/zz_generated_commands.go` or
generated `SKILL.md`. Do not manually invent manifest operation keys to make a
private or undocumented endpoint pass the public contract check.

## 3. How API docs become CLI tools

The generation flow is:

```text
esper-api-docs/openapi.yaml + public OpenAPI snapshot + reviewed overlays
  -> tools/specbundle/bundle.sh and main.mjs
  -> spec/openapi/<generation>.yaml + manifest.json
  -> tools/codegen
  -> internal/cmd/generated/zz_generated_commands.go
  -> generated.AddCommands in internal/cmd/root.go
  -> Cobra help, flags, request execution, discovery, and completion
  -> tools/skillgen -> SKILL.md
```

The bundler filters operations by public HTTP method/path, normalizing the
`/api` prefix and trailing slash for comparison. Backend-only operations are
not enough to create a generated CLI command. GET, POST, PUT, PATCH, and DELETE
are the supported generated methods.

The bundler infers defaults and preserves reviewed `x-esper-*` annotations:

| Annotation | Controls |
| --- | --- |
| `x-esper-generation` on document `info` | API generation |
| `x-esper-noun`, `x-esper-verb` | Resource and action names |
| `x-esper-scope-parent` | Parent scope for nested resources |
| `x-esper-pagination` | Pagination behavior |
| `x-esper-response-envelope` | Response unwrapping |
| `x-esper-destructive` | Additional destructive-action confirmation |
| `x-esper-require-one-of` | Alternative required inputs |
| `x-esper-docs-slugs` | Documentation-based discovery |
| `x-esper-alias-of` | Canonical-operation alias metadata |

Path/query/header parameters and request-body schemas supply flags and input
metadata. Complex bodies use `--body` with inline JSON, `@file`, or stdin.
Newer core generations generally own the default noun/verb; older colliding
operations use explicit generation routing. Same-generation collisions need a
reviewed naming or scope decision, not a silent overwrite.

## 4. Add a new API-backed tool

1. Confirm the endpoint's method, route, auth, request/response schema, and
   deployed public support. Add or correct its definition in `esper-api-docs`
   using that repository's publishing process. The CLI bundler does not publish
   API docs. Until the operation appears in the supplied public snapshot, it
   will be filtered out of generated CLI coverage.
2. Refresh the bundle as described above. Review inferred naming, scopes,
   pagination, envelopes, and destructive behavior. If inference is wrong,
   update the maintained annotation/overlay logic in `tools/specbundle/main.mjs`
   or the upstream annotations, then regenerate. Avoid one-off edits to
   generated output that the next refresh would lose.
3. Check `internal/commandpolicy/policy.go`. A public endpoint can still be
   intentionally excluded or renamed by CLI policy. Change those decisions only
   with reviewed evidence; do not remove unrelated exclusions.
4. Regenerate command metadata and `SKILL.md`. Usually no new Cobra command is
   needed for an ordinary API operation.
5. Add request/response regression coverage following nearby tests under
   `internal/cmd` and fixtures under `spec/fixtures`. Cover method, route, body,
   relevant pagination, errors, and write approval behavior.
6. Verify help and discovery locally, then run the checks below. Any live
   verification is a separate, explicitly scoped step.

For device commands, use DSO `operation create` for supported types. Preserve
the legacy submission guard and reviewed exclusions; see the README migration
guide. When DSO enum support changes, reconcile `supportedDSOType` in the shared
runner and its coverage test with the reviewed spec and backend evidence.

## 5. Add a manually added tool or workflow

Use this path for local commands or multi-step workflows that ordinary API
generation cannot represent. Follow existing examples such as `configure`,
`context`, `discover`, and `secureadb`.

- Implement the command under `internal/cmd/<name>` using existing Cobra and
  `internal/runtime` patterns, then register it in `NewRootCommand` in
  `internal/cmd/root.go`.
- Reuse credential resolution, HTTP handling, error categories, output, and
  approval behavior. Every API write requires human approval; `--yes` only
  skips the separate destructive confirmation.
- Add behavioral tests and document any workflow-specific approval boundary.
  Secure ADB is the existing reviewed exception to generated public-operation
  coverage, not permission to expose other internal endpoints.
- Update manually added command guidance in `tools/skillgen/main.go`, its
  relevant tests, and README. Regenerate `SKILL.md` rather than editing it.
- Check whether the offline command harness needs an explicit entry for the
  new manually added command; generated operations are already metadata-driven.

## 6. Verify and release

```sh
go run ./tools/codegen
go run ./tools/skillgen --check
go run ./tools/contractcheck
go test ./...
go vet ./...
go build ./...
test -z "$(gofmt -l cmd internal tools)"
go run ./tools/command-harness --mode offline --report dist/command-harness-offline.json
git diff --check
```

Review generated diffs and confirm that a second codegen run makes no further
changes. `contractcheck` verifies public operation membership and CLI contract
metadata; the offline harness checks command help, not live API behavior.

Review `.github/workflows/ci.yml` for the current CI checks and
`.goreleaser.yaml` for packaging. CI creates snapshot artifacts; it does not
automatically publish a stable release or upload the generated Homebrew formula.
Publish through the agreed release process and include migration notes for
changed commands. Users receive API coverage changes by installing a new CLI
build, not by refreshing specs on their machines. The Python SDK is separate.
