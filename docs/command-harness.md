# Command Harness

The harness is designed for repeatable runs by CI or Luna. It never uses live
credentials in command arguments. The CLI reads credentials from the normal
state file or `ESPER_CREDS_FILE`.

## Modes

### Offline

Checks every generated operation and every hand-written command by resolving its
Cobra path and rendering help. Aliases are counted separately. No network or
credentials are used.

```bash
./scripts/test-commands.sh offline
```

Report: `dist/command-harness-offline.json`.

### Live read-only

Runs generated GET operations once with bounded inputs. It never adds `--all`
and never invokes POST, PUT, PATCH, DELETE, `--yes`, or Secure ADB. Missing
resource IDs are reported as `SKIPPED`. Collection routes without a `--limit`
parameter are also skipped because the harness will not issue an unbounded list
request.

```bash
export HARNESS_ENTERPRISE=<dev-enterprise-id>
export HARNESS_DEVICE=<owned-dev-device-id>
./scripts/test-commands.sh live-readonly
```

Optional dependent IDs:

```bash
export HARNESS_GROUP=<safe-group-id>
export HARNESS_APP=<safe-application-id>
```

Report: `dist/command-harness-live-readonly.json`.

### Live mutations

Mutation runs are disabled by default. They require:

- `--allow-mutations` through the wrapper.
- A disposable enterprise ID in `HARNESS_ENTERPRISE`.
- An owned disposable device ID in `HARNESS_DEVICE`.
- `HARNESS_MUTATION_CONFIRM=I_UNDERSTAND_THIS_TENANT_IS_DISPOSABLE`.
- A scenario with at least one cleanup step.

The example removes the disposable device and clears context afterward:

```bash
export HARNESS_ENTERPRISE=<disposable-enterprise-id>
export HARNESS_DEVICE=<disposable-device-id>
export HARNESS_MUTATION_CONFIRM=I_UNDERSTAND_THIS_TENANT_IS_DISPOSABLE
./scripts/test-commands.sh live-mutations
```

Mutation scenarios are JSON. Each step must declare `expected_exit`; every
non-GET generated operation must set `mutation: true`, and destructive generated
operations must include `--yes`. Cleanup runs in reverse order even when a setup
or mutation step fails. The example is intentionally not run by CI.

## Luna

Luna can run the same wrapper without prompts:

```bash
./scripts/test-commands.sh offline
./scripts/test-commands.sh live-readonly
```

For mutations, Luna must receive the disposable IDs and confirmation through
environment variables. Do not pass an API key, token, or request body secret on
the command line. The harness emits tab-separated progress and writes a JSON
report under `dist/`.
