# Phase 2 packets

Inventory source: `spec/openapi/*.yaml` at `go-rewrite` HEAD `534cd83`.
Operations are grouped by their generated top-level noun after generation-collision
and side-family grammar rules. Related nouns are combined into resource-domain
packets so the 328 in-scope operations form 25 independently reviewable packets.

Excluded from Phase 2:

- `device` (6 operations): completed as the Phase 1 reference.
- `remoteadb`, `remote-adb-connection` (4 operations): reserved for the
  hand-written `secureadb` module in Phase 3.
- `configure`, `context`, and completions: hand-written modules with no OpenAPI
  operations in this inventory.

## Ordered packet list

| # | Packet | Operations | Top-level nouns | State |
|---:|---|---:|---|---|
| 1 | enterprise | 4 | `enterprise`, `enterprise-policy`, `enterprise-report` | Complete |
| 2 | group | 18 | `device-group`, `group`, `group-eventfeed`, `group-report`, `group-thumbnail`, `legacy-device-group-command`, `sub-group` | Complete |
| 3 | app-application | 53 | `app`, `app-info`, `app-instance`, `app-version`, `app-vpp`, `appleappstore`, `application`, `application-minimal`, `device-app`, `esper-app-version`, `install`, `installdevice`, `itunesapp`, `preferred-region`, `product`, `tenant-app`, `tenant-app-version`, `tenant-esper-app`, `version`, `webclip` | Complete |
| 4 | blueprint | 15 | `blueprint`, `blueprint-revision`, `blueprint-version`, `revision` | Complete |
| 5 | pipeline | 37 | `pipeline`, `pipeline-run`, `stage`, `stage-run`, `target`, `target-bulk`, `target-list`, `target-run` | Complete |
| 6 | token-user | 33 | `authn-user`, `dep-token`, `dep-token-based-on`, `dep-token-upload`, `different-user`, `invite`, `own-user`, `personal-access-token`, `renew-token`, `tenant-user`, `tenant-user-invite`, `tenant-vpptoken`, `token-info`, `user`, `user-delete`, `user-info`, `webtoken`, `webtoken-instance` | Complete |
| 7 | alarm-alert | 13 | `alarm-rule`, `alarm-history`, `alert-channel` | Complete |
| 8 | apns | 5 | `apns-cert`, `apns-csr` | Complete |
| 9 | command-operation | 37 | `command`, `command-history`, `command-inbox`, `command-request`, `command-request-status`, `command-status`, `device-operation`, `operation`, `operation-list`, `pipeline-command`, `pipeline-operation`, `stat`, `status` | Complete |
| 10 | connection | 2 | `connection`, `custom-connection` | Complete |
| 11 | content | 7 | `content`, `download`, `remote-file` | Complete |
| 12 | converge | 3 | `converge` | Complete |
| 13 | custom-action | 6 | `custom-action`, `script` | Complete |
| 14 | dep-sync | 3 | `dep-sync-request` | Complete |
| 15 | device-support | 15 | `device-eventfeed`, `device-google-account-emm-managed`, `device-google-account-policy`, `device-heartbeat`, `device-heartbeat-list`, `device-request`, `devicestate`, `foundation-version-list`, `google-account`, `rv-activity-feed` | Complete |
| 16 | directory-record | 5 | `directory-record` | Complete |
| 17 | emm | 9 | `emm`, `emm-account`, `emm-detail`, `emm-enrollment-begin`, `emm-enrollment-complete`, `emm-instance`, `emm-web-token` | Complete |
| 18 | foundry | 6 | `foundry-build`, `foundry-device-model`, `foundry-event` | Complete |
| 19 | geofence | 9 | `create-apply-geo-fence`, `geofence`, `the-geofence` | Blocked |
| 20 | policy | 5 | `policy` | Blocked |
| 21 | provisioning-profile | 5 | `provisioning-profile`, `provisioning-profile-version` | Complete |
| 22 | report-telemetry | 19 | `device-location`, `device-report`, `device-tile-report`, `event-feed`, `report-info`, `report-status`, `report-type`, `specific-location`, `status-metric`, `subscription`, `subscription-report`, `telemetry-graph-data` | Blocked |
| 23 | role-scope | 7 | `role`, `scope` | Blocked |
| 24 | seamless | 2 | `seamless` | Blocked |
| 25 | tile-ui | 10 | `tile-icon`, `tile-icon-apply`, `tile-icon-unapply`, `wallpaper` | Blocked |
| | **Total** | **328** | | |

## Checkpoint summaries

Progress summaries are appended here after packets 5, 10, 15, 20, and 25.

### Packets 1-5

- `enterprise` - Complete: all 4 operations have schema-shaped success and API-error golden coverage.
- `group` - Complete: all 18 operations have schema-shaped success/API-error golden coverage, including two `--all` flows and the route/body collision rule.
- `app-application` - Complete: all 53 operations have explicit schema-shaped success/API-error fixtures and goldens, with operation-specific pagination coverage.
- `blueprint` - Complete: all 15 operations have schema-shaped success and API-error goldens, including required optional-body inputs, recursive scopes, multipart, pagination, and destructive coverage.
- `pipeline` - Complete: all 37 operations have schema-shaped success/API-error goldens, nine apps-envelope pagination flows, recursive scopes, required-body validation, and six destructive refusals.

### Packets 6-10

- `token-user` - Complete: all 33 operations have schema-shaped success/API-error golden coverage, including raw DEP output, multipart, pagination, body-only inputs, recursive scopes, and six destructive refusals.
- `alarm-alert` - Complete: all 13 operations have schema-shaped success and API-error golden coverage, including body-only create/replace, enterprise auto-fill, PATCH partial input, nested history scope, pagination, and destructive refusals.
- `apns` - Complete: all 5 operations have schema-shaped success/API-error golden coverage, including corrected apps-envelope pagination, raw CSR output, multipart upload, and destructive confirmation.
- `command-operation` - Complete: all 37 operations have schema-shaped success/API-error golden coverage, including 13 pagination flows, side-family collisions, multi-parent scope selection, body input modes, cancellation bodies, bodyless 304 handling, and six destructive refusals.
- `connection` - Complete: both operations have schema-shaped success/API-error golden coverage, including tenant scope, optional filtering, body-only update input, body file/stdin handling, and internal-header exclusion.

### Packets 11-15

- `content` - Complete: all 7 operations have success/API-error goldens, including multipart and `--all` coverage.
- `converge` - Complete: all 3 operations have goldens and apps-envelope `--all` coverage.
- `custom-action` - Complete: all 6 operations have schema-shaped success/API-error golden coverage, including composed-schema body-only create, partial update inputs, apps-envelope pagination, script retrieval, and destructive refusal.
- `dep-sync` - Complete: all 3 operations have schema-shaped success/API-error golden coverage, including apps-envelope pagination and automatic `{}` for the required empty create body.
- `device-support` - Complete: all 15 operations have schema-shaped success/API-error golden coverage, including five pagination flows, explicit array-only policy input, optional Google-account body omission, route/body auto-fill, and two destructive refusals.

### Packets 16-20

- `directory-record` - Complete: all 5 operations have schema-shaped success/API-error golden coverage, including corrected limit-offset pagination, platform-backed body-only create/update inputs, exact bodyless 204 output, and destructive refusal.
- `emm` - Complete: all 9 operations have schema-shaped success/API-error golden coverage, including three pagination flows, validator-backed required inputs, transparent-proxy response envelopes, completion 201 handling, and internal auth-header exclusion.
- `foundry` - Complete: all 6 operations have schema-shaped success/API-error golden coverage, including three apps-envelope pagination flows, platform-backed response envelopes and DTOs, build approval input, partial device-model update input, and internal auth-header exclusion.
- `geofence` - Blocked: body rules and normalized same-generation aliases are undefined.
- `policy` - Blocked: required wildcard bodies, complex properties, and scope/body collisions lack rules.

### Packets 21-25

- `provisioning-profile` - Complete: all 5 operations have success/API-error goldens, multipart, pagination, scope, and destructive coverage.
- `report-telemetry` - Blocked: required array-only subscription input lacks explicit-body enforcement.
- `role-scope` - Blocked: required bodies have only optional scalar or array inputs.
- `seamless` - Blocked: both operations require bodies without required scalar input.
- `tile-ui` - Blocked: scope/body collision and required complex apply input lack locked rules.

## Blockers

- `geofence`: wildcard/required body behavior is unresolved, and normalizing
  the malformed geofence nouns creates an undefined same-generation command
  alias collision. Details are in the steering files.

- `policy`: create and updates have required wildcard bodies plus unresolved
  complex-property enforcement and `--enterprise` scope/body collisions.

- `report-telemetry`: subscription add/update require an array-only
  `email_ids` body, but the locked rules do not define when explicit `--body`
  is mandatory.

- `role-scope`: role create/update have only optional scalar body properties,
  and scope update has only optional arrays; all require bodies under an
  undefined explicit-input rule.

- `seamless`: both create and upload require bodies with no required scalar
  input, so invocation validation depends on the unresolved body rule.

- `tile-ui`: tile-icon add collides `--enterprise` between route and body, and
  tile-icon apply requires complex-only input without an enforcement rule.
  Unapply also needs its destructive overlay corrected after unblock.

## Final status

| Packet | State | Blocker |
|---|---|---|
| enterprise | Complete | - |
| group | Complete | - |
| app-application | Complete | - |
| blueprint | Complete | - |
| pipeline | Complete | - |
| token-user | Complete | - |
| alarm-alert | Complete | - |
| apns | Complete | - |
| command-operation | Complete | - |
| connection | Complete | - |
| content | Complete | - |
| converge | Complete | - |
| custom-action | Complete | - |
| dep-sync | Complete | - |
| device-support | Complete | - |
| directory-record | Complete | - |
| emm | Complete | - |
| foundry | Complete | - |
| geofence | Blocked | Body rules and normalized same-generation alias collision. |
| policy | Blocked | Wildcard/complex bodies and scope/body collisions. |
| provisioning-profile | Complete | - |
| report-telemetry | Blocked | Required array-only body input. |
| role-scope | Blocked | Required bodies with optional scalar/array inputs. |
| seamless | Blocked | Required bodies with no required scalar input. |
| tile-ui | Blocked | Scope/body collision and required complex-only input. |

Summary: 19 packets complete (276 operations); 6 packets blocked (52
operations). All remaining blockers are recorded in `inbox.md` and `spec.md`
Issues.
