# Go CLI live smoke test

Status: human-run only. Run against a dev enterprise. Do not use production
credentials or production devices.

## Prerequisites

- Build or install the `espercli` binary from the `go-rewrite` branch.
- Install `jq` and Android platform tools (`adb`).
- Use a dev API key and a dev device on which Remote ADB is enabled.
- Run the checklist in Bash, with `espercli`, `jq`, and `adb` on `PATH`.

```bash
set -o pipefail
command -v bash
command -v espercli
command -v jq
command -v adb
```

Expected: each command prints an executable path. Exit code: `0` for each.

## 1. Configure credentials

Set the dev environment name and enter the API key without putting it in shell
history:

```bash
export ESPER_SMOKE_ENVIRONMENT=develop
read -rsp "Dev Esper API key: " ESPER_SMOKE_API_KEY; printf '\n'
espercli configure \
  --environment "$ESPER_SMOKE_ENVIRONMENT" <<<"$ESPER_SMOKE_API_KEY"
printf 'exit=%s\n' "$?"
```

Expected stdout:

```text
Configuration saved.
exit=0
```

Expected stderr: the `API key:` prompt. The key is supplied through stdin so it
does not appear in shell history or the `espercli` process argument list.

Verify persistence and redaction:

```bash
espercli configure show
printf 'exit=%s\n' "$?"
```

Expected shape: two lines named `environment` and `api_key`; the environment is
`develop`, and the API key is masked except for its final four characters. The
unredacted key must not appear. Exit code: `0`.

## 2. Set enterprise context

```bash
read -rp "Dev enterprise ID: " ESPER_SMOKE_ENTERPRISE_ID
espercli context set enterprise "$ESPER_SMOKE_ENTERPRISE_ID"
espercli context get enterprise
printf 'exit=%s\n' "$?"
```

Expected shape: `enterprise: <the supplied ID>` from both context commands.
Exit code: `0`.

## 3. List devices and select one

Human-readable list envelope:

```bash
espercli device list --limit 5
printf 'exit=%s\n' "$?"
```

Expected shape: a key/value block containing `count`, `next`, `previous`, and
the serialized `results` collection. Exit code: `0`.

Select the first dev device from the raw JSON envelope:

```bash
export ESPER_SMOKE_DEVICE_ID="$(
  espercli device list --limit 1 --json |
    jq -er '.results[0].id'
)"
printf 'device=%s\n' "$ESPER_SMOKE_DEVICE_ID"
```

Expected shape: `device=<non-empty device ID>`. Exit code: `0`. Stop if `jq`
reports no device.

## 4. Show the selected device

```bash
espercli device get "$ESPER_SMOKE_DEVICE_ID"
printf 'exit=%s\n' "$?"
```

Expected shape: a key/value block containing at least the selected device ID,
name, and state. Exit code: `0`.

## 5. Verify pagination with `--all`

```bash
espercli device list --limit 2 --all
printf 'exit=%s\n' "$?"
```

Expected shape: one merged human-readable table containing all accessible dev
devices, not one envelope per page. Exit code: `0`.

## 6. Verify `--json` through `jq`

```bash
espercli device list --limit 2 --all --json |
  jq -e 'type == "array" and all(.[]; has("id"))'
printf 'exit=%s\n' "$?"
```

Expected stdout: `true`, followed by `exit=0`. The CLI side of the pipe emits
only one valid JSON array and no status text.

## 7. Send a non-destructive ping

Create an immediate `UPDATE_HEARTBEAT` command for the selected device. The
active enterprise context supplies `enterprise_id`.

```bash
ESPER_SMOKE_PING_BODY="$(
  jq -nc --arg device "$ESPER_SMOKE_DEVICE_ID" \
    '{command_type:"DEVICE", devices:[$device], command:"UPDATE_HEARTBEAT", schedule:"IMMEDIATE"}'
)"
espercli command create --body "$ESPER_SMOKE_PING_BODY" --json |
  jq -e 'type == "object"'
printf 'exit=%s\n' "$?"
```

Expected stdout: `true`, followed by `exit=0`. The API response is a JSON
command-request object, commonly containing an `id` and status fields. This is
the checklist's non-destructive command request. Secure ADB later creates its
own temporary relay session.

## 8. Verify destructive confirmation without `--yes`

Answer `n`; this must stop before an API request is sent.

```bash
set +e
printf 'n\n' | espercli device-request delete "$ESPER_SMOKE_DEVICE_ID"
ESPER_SMOKE_CANCEL_EXIT="$?"
set -e
printf 'exit=%s\n' "$ESPER_SMOKE_CANCEL_EXIT"
test "$ESPER_SMOKE_CANCEL_EXIT" -eq 2
```

Expected stderr shape: a confirmation naming the device request path and one
target. Expected final line: `exit=2`. The command must report cancellation and
must not delete or unenroll the device.

## 9. Connect Secure ADB

Store the selected device as active, then start the relay in terminal A:

```bash
espercli context set device "$ESPER_SMOKE_DEVICE_ID"
set +e
espercli --verbose secureadb connect
ESPER_SMOKE_SECUREADB_EXIT="$?"
set -e
printf 'secureadb-exit=%s\n' "$ESPER_SMOKE_SECUREADB_EXIT"
```

Expected terminal A shape:

```text
context: using active device ... for device_id
context: using active enterprise ... for enterprise_id
Secure ADB relay ready. Run:
adb connect 127.0.0.1:<ephemeral-port>
```

The command remains running while it waits for and bridges one local ADB
connection. In terminal B, copy the exact endpoint printed by terminal A:

```bash
adb connect 127.0.0.1:<ephemeral-port>
adb devices
```

Expected terminal B shape: `connected to 127.0.0.1:<port>` and an `adb devices`
entry for that endpoint. Exit code: `0` for each command.

After confirming connectivity, disconnect in terminal B and press Ctrl+C in
terminal A if it has not exited:

```bash
adb disconnect 127.0.0.1:<ephemeral-port>
printf 'disconnect-exit=%s\n' "$?"
```

Expected terminal B final line: `disconnect-exit=0`. Expected terminal A final
shape: `Session duration: ...`, `Bytes transferred: ...`, and
`secureadb-exit=0` when interrupted after the ADB bridge is established. If the
command is interrupted before the local ADB client connects,
`secureadb-exit=4` is expected and does not validate the relay.

## 10. Remove shell secrets

```bash
unset ESPER_SMOKE_API_KEY ESPER_SMOKE_PING_BODY
```

Expected output: none. Exit code: `0`.
