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
| 9 | command-operation | 37 | `command`, `command-history`, `command-inbox`, `command-request`, `command-request-status`, `command-status`, `device-operation`, `operation`, `operation-list`, `stat`, `status` | Pending |
| 10 | connection | 2 | `connection`, `custom-connection` | Pending |
| 11 | content | 7 | `content`, `download`, `remote-file` | Pending |
| 12 | converge | 3 | `converge` | Pending |
| 13 | custom-action | 6 | `custom-action`, `script` | Pending |
| 14 | dep-sync | 3 | `dep-sync-request` | Pending |
| 15 | device-support | 15 | `device-eventfeed`, `device-google-account-emm-managed`, `device-google-account-policy`, `device-heartbeat`, `device-heartbeat-list`, `device-request`, `devicestate`, `foundation-version-list`, `google-account`, `rv-activity-feed` | Pending |
| 16 | directory-record | 5 | `directory-record` | Pending |
| 17 | emm | 9 | `emm`, `emm-account`, `emm-detail`, `emm-enrollment-begin`, `emm-enrollment-complete`, `emm-instance`, `emm-web-token` | Pending |
| 18 | foundry | 6 | `foundry-build`, `foundry-device-model`, `foundry-event` | Pending |
| 19 | geofence | 9 | `create-apply-geo-fence`, `geofence`, `the-geofence` | Pending |
| 20 | policy | 5 | `policy` | Pending |
| 21 | provisioning-profile | 5 | `provisioning-profile`, `provisioning-profile-version` | Pending |
| 22 | report-telemetry | 19 | `device-location`, `device-report`, `device-tile-report`, `event-feed`, `report-info`, `report-status`, `report-type`, `specific-location`, `status-metric`, `subscription`, `subscription-report`, `telemetry-graph-data` | Pending |
| 23 | role-scope | 7 | `role`, `scope` | Pending |
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

## Final status

The final packet/state/blocker table will be written here after all packets are
complete or blocked.
