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
| 2 | group | 18 | `device-group`, `group`, `group-eventfeed`, `group-report`, `group-thumbnail`, `legacy-device-group-command`, `sub-group` | Blocked |
| 3 | app-application | 53 | `app`, `app-info`, `app-instance`, `app-version`, `app-vpp`, `appleappstore`, `application`, `application-minimal`, `device-app`, `esper-app-version`, `install`, `installdevice`, `itunesapp`, `preferred-region`, `product`, `tenant-app`, `tenant-app-version`, `tenant-esper-app`, `version`, `webclip` | Blocked |
| 4 | blueprint | 15 | `blueprint`, `blueprint-revision`, `blueprint-version`, `revision` | Blocked |
| 5 | pipeline | 37 | `pipeline`, `pipeline-run`, `stage`, `stage-run`, `target`, `target-bulk`, `target-list`, `target-run` | Blocked |
| 6 | token-user | 33 | `authn-user`, `dep-token`, `dep-token-based-on`, `dep-token-upload`, `different-user`, `invite`, `own-user`, `personal-access-token`, `renew-token`, `tenant-user`, `tenant-user-invite`, `tenant-vpptoken`, `token-info`, `user`, `user-delete`, `user-info`, `webtoken`, `webtoken-instance` | Blocked |
| 7 | alarm-alert | 13 | `alarm-rule`, `alarmhistory`, `alarmrule`, `alert-channel`, `alertchannel` | Blocked |
| 8 | apns | 5 | `apns-cert`, `apns-csr` | Blocked |
| 9 | command-operation | 37 | `command`, `command-history`, `command-inbox`, `command-request`, `command-request-status`, `command-status`, `device-operation`, `operation`, `operation-list`, `stat`, `status` | Blocked |
| 10 | connection | 2 | `connection`, `custom-connection` | Blocked |
| 11 | content | 7 | `content`, `download`, `remote-file` | Complete |
| 12 | converge | 3 | `converge` | Complete |
| 13 | custom-action | 6 | `custom-action`, `script` | Blocked |
| 14 | dep-sync | 3 | `dep-sync-request` | Blocked |
| 15 | device-support | 15 | `device-eventfeed`, `device-google-account-emm-managed`, `device-google-account-policy`, `device-heartbeat`, `device-heartbeat-list`, `device-request`, `devicestate`, `foundation-version-list`, `google-account`, `rv-activity-feed` | Blocked |
| 16 | directory-record | 5 | `directory-record` | Blocked |
| 17 | emm | 9 | `emm`, `emm-account`, `emm-detail`, `emm-enrollment-begin`, `emm-enrollment-complete`, `emm-instance`, `emm-web-token` | Blocked |
| 18 | foundry | 6 | `foundry-build`, `foundry-device-model`, `foundry-event` | Blocked |
| 19 | geofence | 9 | `create-apply-geo-fence`, `geofence`, `the-geofence` | Blocked |
| 20 | policy | 5 | `policy` | Blocked |
| 21 | provisioning-profile | 5 | `provisioning-profile`, `provisioning-profile-version` | Complete |
| 22 | report-telemetry | 19 | `device-location`, `device-report`, `device-tile-report`, `event-feed`, `report-info`, `report-status`, `report-type`, `specific-location`, `status-metric`, `subscription`, `subscription-report`, `telemetry-graph-data` | Blocked |
| 23 | role-scope | 7 | `role`, `scope` | Blocked |
| 24 | seamless | 2 | `seamless` | Pending |
| 25 | tile-ui | 10 | `tile-icon`, `tile-icon-apply`, `tile-icon-unapply`, `wallpaper` | Pending |
| | **Total** | **328** | | |

## Checkpoint summaries

Progress summaries are appended here after packets 5, 10, 15, 20, and 25.

### Packets 1-5

- `enterprise` - Complete: all 4 operations have schema-shaped success and API-error golden coverage.
- `group` - Blocked: route-scope and body-property `--enterprise` collision lacks a locked rule.
- `app-application` - Blocked: merged routes require conditional query-flag enforcement not defined by conventions.
- `blueprint` - Blocked: required bodies with no required properties lack a locked input rule.
- `pipeline` - Blocked: seven required bodies with no required scalar input depend on the same missing rule.

### Packets 6-10

- `token-user` - Blocked: legacy user partial-update depends on the missing required-body input rule.
- `alarm-alert` - Blocked: required body/object inputs and scope/body flag collisions lack locked rules.
- `apns` - Blocked: empty required body and binary plist output behavior are undefined.
- `command-operation` - Blocked: required-body inputs and a scope/body flag collision lack locked rules.
- `connection` - Blocked: custom connection update has a required body with no required scalar input.

### Packets 11-15

- `content` - Complete: all 7 operations have success/API-error goldens, including multipart and `--all` coverage.
- `converge` - Complete: all 3 operations have goldens and apps-envelope `--all` coverage.
- `custom-action` - Blocked: required complex properties and body-only update input lack locked enforcement.
- `dep-sync` - Blocked: empty required JSON body behavior is undefined.
- `device-support` - Blocked: Google EMM policy update has a required body with only an optional array.

### Packets 16-20

- `directory-record` - Blocked: create/update bodies have no required properties.
- `emm` - Blocked: detail/account create bodies have no required properties.
- `foundry` - Blocked: update bodies have only optional scalar properties.
- `geofence` - Blocked: body rules and normalized same-generation aliases are undefined.
- `policy` - Blocked: required wildcard bodies, complex properties, and scope/body collisions lack rules.

## Blockers

- `group`: `device-group create` requires `--enterprise` as both a route scope
  and a scalar JSON body property, but the locked conventions define no
  collision rule. The source body is also incorrectly advertised as `*/*`.
  Details are recorded in `inbox.md` and `spec.md` Issues.

- `app-application`: merged `device-app list` routes have different required
  query flags, but the locked scoped-collection rules do not define
  route-conditional required-flag enforcement. Details are recorded in
  `inbox.md` and `spec.md` Issues.

- `blueprint`: two operations require a request body but require none of its
  properties. The locked body rules do not define whether empty input is valid
  or an explicit body/property flag is mandatory. Details are recorded in
  `inbox.md` and `spec.md` Issues.

- `pipeline`: seven operations require request bodies with no required scalar
  input, including top-level array and empty-object schemas. The locked body
  rules do not define whether explicit `--body` is mandatory. Details are
  recorded in `inbox.md` and `spec.md` Issues.

- `token-user`: legacy `user partial-update` requires a JSON body whose schema
  has no required properties. It depends on the unresolved required-body input
  rule recorded in `inbox.md` and `spec.md` Issues.

- `alarm-alert`: required body/object-only inputs lack a locked enforcement
  rule, and create routes collide `--enterprise` between scope and body.
  Details are recorded in `inbox.md` and `spec.md` Issues.

- `apns`: CSR creation requires an empty JSON body with no locked invocation
  rule, and its binary plist response has no defined output behavior. Details
  are recorded in `inbox.md` and `spec.md` Issues.

- `command-operation`: ten required-body operations depend on the missing input
  rule, and v0 command create also collides `--enterprise` between scope and
  body. Details are recorded in `inbox.md` and `spec.md` Issues.

- `connection`: custom connection update requires a JSON body whose scalar
  properties are optional and whose `config` property is object-only. It
  depends on the unresolved required-body rule.

- `custom-action`: create requires object-only properties and update has no
  required scalar input. Both depend on the unresolved required-body/property
  enforcement rule recorded in the steering files.

- `dep-sync`: create requires an empty JSON object, whose implicit versus
  explicit CLI input behavior is not defined by the locked body rules.

- `device-support`: Google EMM policy update requires a JSON body with only an
  optional array property, so it depends on the unresolved explicit-body rule.

- `directory-record`: create and update require JSON bodies without required
  properties, so their input enforcement depends on the unresolved body rule.

- `emm`: EMM detail and account creation require JSON bodies with no required
  properties, so they depend on the unresolved required-body rule.

- `foundry`: build and device-model updates require JSON bodies whose scalar
  properties are all optional, so required input behavior is undefined.

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

## Final status

The final packet/state/blocker table will be written here after all packets are
complete or blocked.
