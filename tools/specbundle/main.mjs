#!/usr/bin/env node

import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";

const [input, outputDir] = process.argv.slice(2);
if (!input || !outputDir) {
  console.error("usage: main.mjs <bundled-openapi.json> <output-directory>");
  process.exit(2);
}

const source = JSON.parse(await readFile(input, "utf8"));
const excludedPaths = new Set([
  "/sys/health",
  "/device/v0/devices/",
]);

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
      apiPath.startsWith("/user")) return "v0";
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
  sourcePathCount: Object.keys(source.paths).length,
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
