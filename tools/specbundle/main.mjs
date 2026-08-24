#!/usr/bin/env node

import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";

const [input, outputDir] = process.argv.slice(2);
if (!input || !outputDir) {
  console.error("usage: main.mjs <bundled-openapi.json> <output-directory>");
  process.exit(2);
}

const source = JSON.parse(await readFile(input, "utf8"));
const sourcePathCount = Object.keys(source.paths).length;
const excludedPaths = new Set([
  "/sys/health",
  "/device/v0/devices/",
]);

const uuidParameter = (name) => ({
  name,
  in: "path",
  required: true,
  schema: { type: "string", format: "uuid" },
});

const jsonResponse = (description, schema = { type: "object", additionalProperties: true }) => ({
  description,
  content: { "application/json": { schema } },
});

const errorResponses = {
  "400": jsonResponse("Bad request"),
  "401": jsonResponse("Unauthorized"),
  "404": jsonResponse("Not found"),
};

const remoteADBProperties = {
  id: { type: "string", format: "uuid" },
  enterprise: { type: "string", format: "uuid" },
  device: { type: "string", format: "uuid" },
  device_name: { type: "string" },
  ip: { type: ["string", "null"] },
  client_port: { type: ["string", "null"] },
  device_port: { type: ["string", "null"] },
  remoteadb_host: { type: ["string", "null"] },
  client_certificate: { type: ["string", "null"] },
  device_certificate: { type: ["string", "null"] },
  state: { type: "string" },
  errors: { type: ["string", "null"] },
};
const remoteADBResponse = { type: "object", properties: remoteADBProperties };
const remoteADBParameters = [uuidParameter("enterprise_id"), uuidParameter("device_id")];
const remoteADBSource = [
  "src/services/shoonyacloud/shoonyapoc/api/remoteadb/urls.py:13-22",
  "src/services/shoonyacloud/shoonyapoc/api/remoteadb/views/remoteadb_connection.py:9-20",
];

// Platform source: api/remoteadb/urls.py:13-22 registers a ModelViewSet below enterprise/device.
source.paths["/v0/enterprise/{enterprise_id}/device/{device_id}/remoteadb/"] = {
  get: {
    operationId: "listRemoteADBConnections",
    summary: "List remote ADB connections for a device",
    tags: ["Remote ADB"],
    parameters: [
      ...remoteADBParameters,
      { name: "limit", in: "query", schema: { type: "integer", minimum: 1 } },
      { name: "offset", in: "query", schema: { type: "integer", minimum: 0 } },
    ],
    responses: {
      "200": jsonResponse("Remote ADB connections", {
        type: "object",
        properties: {
          count: { type: "integer" },
          next: { type: ["string", "null"], format: "uri" },
          previous: { type: ["string", "null"], format: "uri" },
          results: { type: "array", items: remoteADBResponse },
        },
      }),
      ...errorResponses,
    },
    "x-esper-platform-source": remoteADBSource,
  },
  post: {
    operationId: "createRemoteADBConnection",
    summary: "Start a remote ADB connection",
    tags: ["Remote ADB"],
    parameters: remoteADBParameters,
    requestBody: {
      required: true,
      content: {
        "application/json": {
          schema: {
            type: "object",
            required: ["client_certificate"],
            properties: { client_certificate: { type: "string" } },
          },
        },
      },
    },
    responses: { "201": jsonResponse("Remote ADB connection created", remoteADBResponse), ...errorResponses },
    "x-esper-platform-source": remoteADBSource,
  },
};

source.paths["/v0/enterprise/{enterprise_id}/device/{device_id}/remoteadb/{remoteadb_id}/"] = {
  get: {
    operationId: "getRemoteADBConnection",
    summary: "Get remote ADB connection status",
    tags: ["Remote ADB"],
    parameters: [...remoteADBParameters, uuidParameter("remoteadb_id")],
    responses: { "200": jsonResponse("Remote ADB connection", remoteADBResponse), ...errorResponses },
    "x-esper-platform-source": remoteADBSource,
  },
  delete: {
    operationId: "deleteRemoteADBConnection",
    summary: "Delete a remote ADB connection",
    tags: ["Remote ADB"],
    parameters: [...remoteADBParameters, uuidParameter("remoteadb_id")],
    responses: { "204": { description: "Remote ADB connection deleted" }, ...errorResponses },
    "x-esper-platform-source": remoteADBSource,
  },
};

// Platform source: shoonyapoc/urls.py:158-160 and telemetry_graph.py:44-74.
source.paths["/graph/{category}/{metric}/"] = {
  get: {
    operationId: "getTelemetryGraphData",
    summary: "Get telemetry graph data",
    tags: ["Device Telemetry"],
    parameters: [
      { name: "category", in: "path", required: true, schema: { type: "string" } },
      { name: "metric", in: "path", required: true, schema: { type: "string" } },
      { name: "from_time", in: "query", required: true, schema: { type: "string", format: "date-time" } },
      { name: "to_time", in: "query", required: true, schema: { type: "string", format: "date-time" } },
      { name: "period", in: "query", required: true, schema: { type: "string" } },
      { name: "statistic", in: "query", required: true, schema: { type: "string" } },
      uuidParameter("device_id"),
      { name: "enterprise_id", in: "query", schema: { type: "string", format: "uuid" } },
    ].map((parameter) => parameter.name === "device_id" ? { ...parameter, in: "query", required: true } : parameter),
    responses: { "200": jsonResponse("Telemetry data"), ...errorResponses },
    "x-esper-platform-source": [
      "src/services/shoonyacloud/shoonyapoc/shoonyapoc/urls.py:158-160",
      "src/services/shoonyacloud/shoonyapoc/api/device/views/telemetry_graph.py:44-74",
    ],
  },
};

const groupCommandParameters = [uuidParameter("enterprise_id"), uuidParameter("group_id")];
const groupCommandSource = [
  "src/services/shoonyacloud/shoonyapoc/api/enterprise/urls.py:250-256",
  "src/services/shoonyacloud/shoonyapoc/api/device/views/group_command.py:43-93",
];
const groupCommandResponse = {
  type: "object",
  properties: {
    id: { type: "string", format: "uuid" },
    command: { type: "string" },
    command_args: { type: "object", additionalProperties: true },
    state: { type: "string" },
    details: { type: "object", additionalProperties: true },
  },
};

// Platform source: api/enterprise/urls.py:250-256 nests GroupCommandViewSet under devicegroup.
source.paths["/enterprise/{enterprise_id}/devicegroup/{group_id}/command/"] = {
  get: {
    operationId: "listLegacyDeviceGroupCommands",
    summary: "List commands for a device group",
    tags: ["Device Group Command"],
    parameters: [
      ...groupCommandParameters,
      { name: "limit", in: "query", schema: { type: "integer", minimum: 1 } },
      { name: "offset", in: "query", schema: { type: "integer", minimum: 0 } },
    ],
    responses: {
      "200": jsonResponse("Device group commands", {
        type: "object",
        properties: {
          count: { type: "integer" },
          next: { type: ["string", "null"], format: "uri" },
          previous: { type: ["string", "null"], format: "uri" },
          results: { type: "array", items: groupCommandResponse },
        },
      }),
      ...errorResponses,
    },
    "x-esper-platform-source": groupCommandSource,
  },
  post: {
    operationId: "createLegacyDeviceGroupCommand",
    summary: "Send a command to a device group",
    tags: ["Device Group Command"],
    parameters: groupCommandParameters,
    requestBody: {
      required: true,
      content: {
        "application/json": {
          schema: {
            type: "object",
            required: ["command"],
            properties: {
              command: { type: "string" },
              command_args: { type: "object", additionalProperties: true },
              schedule: { type: "object", additionalProperties: true },
            },
          },
        },
      },
    },
    responses: { "201": jsonResponse("Device group command created", groupCommandResponse), ...errorResponses },
    "x-esper-platform-source": groupCommandSource,
  },
};

source.paths["/enterprise/{enterprise_id}/devicegroup/{group_id}/command/{command_id}/"] = {
  get: {
    operationId: "getLegacyDeviceGroupCommand",
    summary: "Get a device group command",
    tags: ["Device Group Command"],
    parameters: [...groupCommandParameters, uuidParameter("command_id")],
    responses: { "200": jsonResponse("Device group command", groupCommandResponse), ...errorResponses },
    "x-esper-platform-source": groupCommandSource,
  },
};

function generationFor(apiPath) {
  if (apiPath.startsWith("/authn2/")) return "authn2";
  if (apiPath.startsWith("/authz2/")) return "authz2";
  if (apiPath.startsWith("/apps/v0/")) return "apps-v0";
  if (apiPath.startsWith("/commands/v0/")) return "commands-v0";
  if (apiPath.startsWith("/device/v0/")) return "device-v0";
  if (apiPath.startsWith("/v1/foundry/")) return "foundry-v1";
  if (apiPath.startsWith("/onboarding/v0/")) return "onboarding-v0";
  if (apiPath.startsWith("/v0/operations")) return "operations-v0";
  if (apiPath.startsWith("/pipelines/v0/")) return "pipelines-v0";
  if (apiPath.startsWith("/report/v0/")) return "reports-v0";
  if (apiPath.startsWith("/tenant/v0/")) return "tenant-v0";
  if (apiPath.startsWith("/v1/")) return "v1";
  if (apiPath.startsWith("/v2/") || apiPath.startsWith("/api/v2/")) return "v2";
  if (apiPath.startsWith("/v0/") || apiPath.startsWith("/enterprise/") ||
      apiPath.startsWith("/user") || apiPath.startsWith("/graph/")) return "v0";
  throw new Error(`no generation mapping for ${apiPath}`);
}

function decodePointerPart(part) {
  return part.replaceAll("~1", "/").replaceAll("~0", "~");
}

function referencedComponents(value, found = new Set()) {
  if (Array.isArray(value)) {
    for (const item of value) referencedComponents(item, found);
    return found;
  }
  if (!value || typeof value !== "object") return found;

  if (typeof value.$ref === "string" && value.$ref.startsWith("#/components/")) {
    found.add(value.$ref);
  }
  for (const child of Object.values(value)) referencedComponents(child, found);
  return found;
}

function componentAt(ref) {
  const parts = ref.slice(2).split("/").map(decodePointerPart);
  let value = source;
  for (const part of parts) value = value?.[part];
  if (value === undefined) throw new Error(`unresolved component reference ${ref}`);
  return { parts, value };
}

function componentsFor(paths) {
  const pending = [...referencedComponents(paths)];
  const visited = new Set();
  const components = {};

  while (pending.length > 0) {
    const ref = pending.pop();
    if (visited.has(ref)) continue;
    visited.add(ref);

    const { parts, value } = componentAt(ref);
    const [, type, name] = parts;
    components[type] ??= {};
    components[type][name] = value;
    for (const nested of referencedComponents(value)) pending.push(nested);
  }

  const securityNames = new Set();
  function collectSecurity(value) {
    if (Array.isArray(value)) {
      for (const item of value) collectSecurity(item);
      return;
    }
    if (!value || typeof value !== "object") return;
    if (Array.isArray(value.security)) {
      for (const requirement of value.security) {
        for (const name of Object.keys(requirement)) securityNames.add(name);
      }
    }
    for (const child of Object.values(value)) collectSecurity(child);
  }
  collectSecurity({ security: source.security, paths });
  for (const name of securityNames) {
    const scheme = source.components?.securitySchemes?.[name];
    if (!scheme) throw new Error(`undefined security scheme ${name}`);
    components.securitySchemes ??= {};
    components.securitySchemes[name] = scheme;
  }

  return components;
}

function convertSchemasTo31(value) {
  if (Array.isArray(value)) {
    for (const item of value) convertSchemasTo31(item);
    return;
  }
  if (!value || typeof value !== "object") return;

  if (value.nullable === true) {
    if (typeof value.type === "string") value.type = [value.type, "null"];
    else if (Array.isArray(value.type) && !value.type.includes("null")) value.type.push("null");
    else if (value.$ref) {
      value.anyOf = [{ $ref: value.$ref }, { type: "null" }];
      delete value.$ref;
    }
  }
  delete value.nullable;

  if (value.exclusiveMinimum === true && typeof value.minimum === "number") {
    value.exclusiveMinimum = value.minimum;
    delete value.minimum;
  } else if (value.exclusiveMinimum === false) {
    delete value.exclusiveMinimum;
  }
  if (value.exclusiveMaximum === true && typeof value.maximum === "number") {
    value.exclusiveMaximum = value.maximum;
    delete value.maximum;
  } else if (value.exclusiveMaximum === false) {
    delete value.exclusiveMaximum;
  }

  for (const child of Object.values(value)) convertSchemasTo31(child);
}

const byGeneration = new Map();
for (const [apiPath, pathItem] of Object.entries(source.paths)) {
  if (excludedPaths.has(apiPath)) continue;
  const generation = generationFor(apiPath);
  if (!byGeneration.has(generation)) byGeneration.set(generation, {});
  byGeneration.get(generation)[apiPath] = pathItem;
}

await mkdir(outputDir, { recursive: true });
const manifest = {
  sourcePathCount,
  addedPathCount: Object.keys(source.paths).length - sourcePathCount,
  excludedPaths: [...excludedPaths].sort(),
  emittedPathCount: 0,
  generations: {},
};

for (const generation of [...byGeneration.keys()].sort()) {
  const paths = byGeneration.get(generation);
  const usedTags = new Set();
  for (const pathItem of Object.values(paths)) {
    for (const operation of Object.values(pathItem)) {
      if (operation && typeof operation === "object" && Array.isArray(operation.tags)) {
        for (const tag of operation.tags) usedTags.add(tag);
      }
    }
  }

  const spec = {
    openapi: "3.1.0",
    info: {
      title: `${source.info.title} - ${generation}`,
      version: source.info.version,
      "x-esper-generation": generation,
      "x-esper-auth": "bearer",
    },
    servers: source.servers,
    tags: (source.tags ?? []).filter((tag) => usedTags.has(tag.name)),
    paths,
    components: componentsFor(paths),
  };
  if (source.security) spec.security = source.security;
  convertSchemasTo31(spec);

  const filename = `${generation}.yaml`;
  await writeFile(path.join(outputDir, filename), `${JSON.stringify(spec, null, 2)}\n`);
  manifest.generations[generation] = Object.keys(paths).length;
  manifest.emittedPathCount += Object.keys(paths).length;
}

await writeFile(
  path.join(outputDir, "manifest.json"),
  `${JSON.stringify(manifest, null, 2)}\n`,
);

console.log(`emitted ${manifest.emittedPathCount} paths across ${byGeneration.size} generations`);
