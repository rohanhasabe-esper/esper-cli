# Canonical OpenAPI specifications

These resolved OpenAPI 3.1 files are generated from the private
`github.com/esper-code/esper-api-docs` repository at revision
`d5563a0e35361f9c0dd0aefec140f6071619e8a1`.

The upstream assembly joins 12 platform service specifications, rewrites
cross-service references, prefixes colliding tags and components, and splits
the result into 205 path files. `tools/specbundle/bundle.sh` bundles that split
source and partitions it into API generations without external references.

Two source paths are intentionally excluded:

- `/device/v0/devices/` is deprecated in the device API public specification.
- `/sys/health` is an operational health endpoint, not a customer API.

Run the following from an authenticated local checkout:

```sh
tools/specbundle/bundle.sh /path/to/esper-api-docs
npx --yes @redocly/cli@1.34.5 lint --config spec/redocly.yaml
```

`manifest.json` records the source, excluded, emitted, and per-generation path
counts. Generated specifications are JSON-formatted YAML so downstream Go tools
can parse them with the standard library.

## Platform reconciliation

The bundler adds customer routes present in `esper-platform` but absent from the
docs source. Each generated operation carries an `x-esper-platform-source`
citation with the exact router and view files:

- Remote ADB nested routes from `api/remoteadb/urls.py`.
- Telemetry graph data from `shoonyapoc/urls.py` and
  `api/device/views/telemetry_graph.py`.
- Legacy device-group commands from `api/enterprise/urls.py` and
  `api/device/views/group_command.py`.

Unversioned core routes are emitted as the `legacy` generation. Versioned core
and service routes are grouped by `v0`, `v1`, or `v2`; independent API families
such as `pipelines-v0`, `authn2`, `authz2`, and `foundry` retain distinct
generation names. This partition feeds the locked generation-collision rule.
