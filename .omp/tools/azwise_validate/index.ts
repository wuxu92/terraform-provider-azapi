import type { CustomToolFactory } from "@oh-my-pi/pi-coding-agent";
import { readFile, access, readdir } from "node:fs/promises";
import { join, dirname, resolve } from "node:path";

// ---------------------------------------------------------------------------
// Azurerm repo resolution (shared logic with azwise_extract)
// ---------------------------------------------------------------------------

async function resolveAzurermPath(cwd: string): Promise<string | null> {
  const envPath = process.env.AZURERM_PATH;
  if (envPath) {
    const candidate = resolve(envPath);
    try {
      await access(join(candidate, "internal", "services"));
      return candidate;
    } catch {}
  }
  const sibling = resolve(cwd, "..", "terraform-provider-azurerm");
  try {
    await access(join(sibling, "internal", "services"));
    return sibling;
  } catch {}
  let dir = cwd;
  for (let i = 0; i < 5; i++) {
    try {
      const goWork = await readFile(join(dir, "go.work"), "utf-8");
      const match = goWork.match(/use\s+(\S*terraform-provider-azurerm\S*)/);
      if (match) {
        const candidate = resolve(dir, match[1]);
        try {
          await access(join(candidate, "internal", "services"));
          return candidate;
        } catch {}
      }
    } catch {}
    const parent = dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }
  return null;
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface Issue {
  severity: "error" | "warning";
  category: string;
  propertyPath: string;
  ruleKind: string;
  message: string;
  detail?: string;
}

interface RuleEntry {
  kind: "StringRule" | "IntRule" | "FloatRule";
  propertyPath: string;
  allowedValues?: string[];
  raw: string;
}

interface GoField {
  name: string;
  typeName: string;
  jsonTag: string;
  isPointer: boolean;
  isSlice: boolean;
}

// ---------------------------------------------------------------------------
// Go knowledge file parser
// ---------------------------------------------------------------------------

function parseKnowledgeFile(source: string): {
  resourceType: string;
  apiVersions: string[];
  rules: RuleEntry[];
  computedFields: string[];
  forceNewPaths: string[];
  sensitiveFields: string[];
  defaultValuePaths: string[];
} {
  const resourceType =
    source.match(/ResourceType:\s*"([^"]+)"/)?.[1] ?? "";
  const apiVersions: string[] = [];
  const avMatch = source.match(
    /ApiVersions:\s*\[\]string\{([^}]+)\}/
  );
  if (avMatch) {
    for (const m of avMatch[1].matchAll(/"([^"]+)"/g)) {
      apiVersions.push(m[1]);
    }
  }

  const rules: RuleEntry[] = [];

  // Extract StringRules with AllowedValues and PropertyPath
  const stringRulesBlock = extractSliceBlock(source, "StringRules");
  if (stringRulesBlock) {
    for (const entry of extractStructLiterals(stringRulesBlock)) {
      const pp =
        entry.match(/PropertyPath:\s*"([^"]*)"/)?.[1] ?? "";
      if (!pp) continue; // name rule, skip
      const avs: string[] = [];
      const avBlock = entry.match(
        /AllowedValues:\s*\[\]string\{([^}]+)\}/
      );
      if (avBlock) {
        for (const m of avBlock[1].matchAll(/"([^"]+)"/g)) {
          avs.push(m[1]);
        }
      }
      rules.push({
        kind: "StringRule",
        propertyPath: pp,
        allowedValues: avs.length > 0 ? avs : undefined,
        raw: entry.trim(),
      });
    }
  }

  // Extract IntRules
  const intRulesBlock = extractSliceBlock(source, "IntRules");
  if (intRulesBlock) {
    for (const entry of extractStructLiterals(intRulesBlock)) {
      const pp =
        entry.match(/PropertyPath:\s*"([^"]*)"/)?.[1] ?? "";
      if (!pp) continue;
      rules.push({
        kind: "IntRule",
        propertyPath: pp,
        raw: entry.trim(),
      });
    }
  }

  // Extract FloatRules
  const floatRulesBlock = extractSliceBlock(source, "FloatRules");
  if (floatRulesBlock) {
    for (const entry of extractStructLiterals(floatRulesBlock)) {
      const pp =
        entry.match(/PropertyPath:\s*"([^"]*)"/)?.[1] ?? "";
      if (!pp) continue;
      rules.push({
        kind: "FloatRule",
        propertyPath: pp,
        raw: entry.trim(),
      });
    }
  }

  // Extract ComputedFields
  const computedFields: string[] = [];
  const cfMatch = source.match(
    /ComputedFields:\s*\[\]string\{([^}]+)\}/
  );
  if (cfMatch) {
    for (const m of cfMatch[1].matchAll(/"([^"]+)"/g)) {
      computedFields.push(m[1]);
    }
  }

  // Extract ForceNew paths
  const forceNewPaths: string[] = [];
  const fnBlock = extractSliceBlock(source, "ForceNew");
  if (fnBlock) {
    for (const m of fnBlock.matchAll(/PropertyPath:\s*"([^"]+)"/g)) {
      forceNewPaths.push(m[1]);
    }
  }

  // Extract SensitiveFields
  const sensitiveFields: string[] = [];
  const sfMatch = source.match(
    /SensitiveFields:\s*\[\]string\{([^}]+)\}/
  );
  if (sfMatch) {
    for (const m of sfMatch[1].matchAll(/"([^"]+)"/g)) {
      sensitiveFields.push(m[1]);
    }
  }

  // Extract DefaultValues paths
  const defaultValuePaths: string[] = [];
  const dvBlock = extractSliceBlock(source, "DefaultValues");
  if (dvBlock) {
    for (const m of dvBlock.matchAll(/PropertyPath:\s*"([^"]+)"/g)) {
      defaultValuePaths.push(m[1]);
    }
  }

  return {
    resourceType,
    apiVersions,
    rules,
    computedFields,
    forceNewPaths,
    sensitiveFields,
    defaultValuePaths,
  };
}

/** Extract the content of a named slice field, e.g. `StringRules: []StringRule{...}` */
function extractSliceBlock(
  source: string,
  fieldName: string
): string | null {
  // Match `FieldName: []Type{` and then balanced braces
  const re = new RegExp(
    `${fieldName}:\\s*\\[\\][A-Za-z]+\\{`
  );
  const m = re.exec(source);
  if (!m) return null;
  const start = m.index + m[0].length;
  let depth = 1;
  let i = start;
  while (i < source.length && depth > 0) {
    if (source[i] === "{") depth++;
    else if (source[i] === "}") depth--;
    i++;
  }
  return source.slice(start, i - 1);
}

/** Extract individual `{ ... }` struct literals from a slice block */
function extractStructLiterals(block: string): string[] {
  const results: string[] = [];
  let i = 0;
  while (i < block.length) {
    if (block[i] === "{") {
      let depth = 1;
      let j = i + 1;
      while (j < block.length && depth > 0) {
        if (block[j] === "{") depth++;
        else if (block[j] === "}") depth--;
        j++;
      }
      results.push(block.slice(i, j));
      i = j;
    } else {
      i++;
    }
  }
  return results;
}

// ---------------------------------------------------------------------------
// SDK model parser
// ---------------------------------------------------------------------------

function parseGoStruct(source: string): Map<string, GoField[]> {
  const structs = new Map<string, GoField[]>();
  const re = /type\s+(\w+)\s+struct\s*\{/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(source)) !== null) {
    const name = m[1];
    const start = m.index + m[0].length;
    let depth = 1;
    let i = start;
    while (i < source.length && depth > 0) {
      if (source[i] === "{") depth++;
      else if (source[i] === "}") depth--;
      i++;
    }
    const body = source.slice(start, i - 1);
    const fields: GoField[] = [];
    // Match lines like:
    //   FieldName *TypeName `json:"fieldName,omitempty"`
    //   FieldName []TypeName `json:"fieldName,omitempty"`
    //   FieldName TypeName `json:"fieldName,omitempty"`
    const fieldRe =
      /^\s*(\w+)\s+(\*?)(\[\])?\s*(\w+)\s+`json:"([^"]+)"`/gm;
    let fm: RegExpExecArray | null;
    while ((fm = fieldRe.exec(body)) !== null) {
      const jsonTag = fm[5].split(",")[0]; // strip ,omitempty
      fields.push({
        name: fm[1],
        isPointer: fm[2] === "*",
        isSlice: fm[3] === "[]",
        typeName: fm[4],
        jsonTag,
      });
    }
    structs.set(name, fields);
  }
  return structs;
}

function parseEnumValues(
  source: string
): Map<string, string[]> {
  const enums = new Map<string, string[]>();
  const re =
    /func\s+PossibleValuesFor(\w+)\(\)\s*\[\]string\s*\{/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(source)) !== null) {
    const typeName = m[1];
    const start = m.index + m[0].length;
    let depth = 1;
    let i = start;
    while (i < source.length && depth > 0) {
      if (source[i] === "{") depth++;
      else if (source[i] === "}") depth--;
      i++;
    }
    const body = source.slice(start, i - 1);
    const values: string[] = [];
    // Look for the return statement's []string{...}
    const returnMatch = body.match(
      /return\s*\[\]string\{([^}]+)\}/
    );
    if (returnMatch) {
      // Values are either `string(ConstName)` or `"literal"`
      for (const vm of returnMatch[1].matchAll(
        /string\(\w+\)|"([^"]+)"/g
      )) {
        if (vm[1]) {
          values.push(vm[1]);
        }
      }
    }
    // If we only got `string(ConstName)` patterns, resolve from const blocks
    if (values.length === 0) {
      // Fall back to extracting from const declarations
      const constRe = new RegExp(
        `(\\w+)\\s+${typeName}\\s*=\\s*"([^"]+)"`,
        "g"
      );
      let cm: RegExpExecArray | null;
      while ((cm = constRe.exec(source)) !== null) {
        values.push(cm[2]);
      }
    }
    if (values.length > 0) {
      enums.set(typeName, values);
    }
  }
  return enums;
}

// ---------------------------------------------------------------------------
// Path resolution engine
// ---------------------------------------------------------------------------

interface ResolvedPath {
  goType: string;
  isEnum: boolean;
  isNumeric: boolean;
  isString: boolean;
  isBool: boolean;
  isStruct: boolean;
  isSlice: boolean;
  enumValues?: string[];
  traceLog: string[];
}

const GO_PRIMITIVE_TYPES: Record<
  string,
  { isNumeric: boolean; isString: boolean; isBool: boolean }
> = {
  string: { isNumeric: false, isString: true, isBool: false },
  bool: { isNumeric: false, isString: false, isBool: true },
  int64: { isNumeric: true, isString: false, isBool: false },
  int32: { isNumeric: true, isString: false, isBool: false },
  int: { isNumeric: true, isString: false, isBool: false },
  float64: { isNumeric: true, isString: false, isBool: false },
  float32: { isNumeric: true, isString: false, isBool: false },
};

function resolvePropertyPath(
  path: string,
  rootStructName: string,
  structs: Map<string, GoField[]>,
  enums: Map<string, string[]>,
  allConstants: Map<string, string[]>
): ResolvedPath | null {
  const segments = path.split(".");
  const traceLog: string[] = [];
  let currentType = rootStructName;

  for (let si = 0; si < segments.length; si++) {
    let seg = segments[si];
    // Strip array index markers like [*]
    seg = seg.replace(/\[\*\]$/, "");

    const fields = structs.get(currentType);
    if (!fields) {
      traceLog.push(
        `✗ struct ${currentType} not found in SDK models`
      );
      return null;
    }

    const field = fields.find((f) => f.jsonTag === seg);
    if (!field) {
      traceLog.push(
        `✗ no field with json:"${seg}" in ${currentType} (available: ${fields.map((f) => f.jsonTag).join(", ")})`
      );
      return null;
    }

    traceLog.push(
      `${seg} → ${field.name}: ${field.isPointer ? "*" : ""}${field.isSlice ? "[]" : ""}${field.typeName} in ${currentType}`
    );

    if (si === segments.length - 1) {
      // Leaf — determine type category
      const typeName = field.typeName;
      const prim = GO_PRIMITIVE_TYPES[typeName];
      if (prim) {
        return {
          goType: typeName,
          isEnum: false,
          isNumeric: prim.isNumeric,
          isString: prim.isString,
          isBool: prim.isBool,
          isStruct: false,
          isSlice: field.isSlice,
          traceLog,
        };
      }
      // Check if it's an enum
      const enumVals =
        enums.get(typeName) || allConstants.get(typeName);
      if (enumVals) {
        return {
          goType: typeName,
          isEnum: true,
          isNumeric: false,
          isString: true, // Go enums based on string
          isBool: false,
          isStruct: false,
          isSlice: field.isSlice,
          enumValues: enumVals,
          traceLog,
        };
      }
      // Check if it's a struct
      if (structs.has(typeName)) {
        return {
          goType: typeName,
          isEnum: false,
          isNumeric: false,
          isString: false,
          isBool: false,
          isStruct: true,
          isSlice: field.isSlice,
          traceLog,
        };
      }
      // Unknown type — could be interface{}, map, etc.
      return {
        goType: typeName,
        isEnum: false,
        isNumeric: false,
        isString: false,
        isBool: false,
        isStruct: false,
        isSlice: field.isSlice,
        traceLog,
      };
    }

    // Navigate deeper
    currentType = field.typeName;
    if (!structs.has(currentType)) {
      traceLog.push(
        `✗ intermediate type ${currentType} not found in SDK models (segment: ${seg})`
      );
      return null;
    }
  }
  return null;
}

// ---------------------------------------------------------------------------
// SDK path resolution: resource type → vendor directory
// ---------------------------------------------------------------------------

async function findSdkPackage(
  azurermPath: string,
  resourceType: string,
  apiVersion: string
): Promise<string | null> {
  // Microsoft.Storage/storageAccounts → storage, storageaccounts
  const parts = resourceType.split("/");
  if (parts.length < 2) return null;
  const namespace = parts[0]; // Microsoft.Storage
  const resourceName = parts[1]; // storageAccounts

  // Extract service from namespace: Microsoft.Storage → storage
  const service = namespace
    .replace(/^Microsoft\./i, "")
    .toLowerCase();
  const resourceDir = resourceName.toLowerCase();

  const sdkBase = join(
    azurermPath,
    "vendor",
    "github.com",
    "hashicorp",
    "go-azure-sdk",
    "resource-manager"
  );

  // Try exact service name first, then common aliases
  const serviceCandidates = [service];

  // Try exact apiVersion first, then scan for available versions
  const apiVersionCandidates: string[] = [];
  if (apiVersion) apiVersionCandidates.push(apiVersion);

  for (const svc of serviceCandidates) {
    const svcDir = join(sdkBase, svc);
    try {
      await access(svcDir);
    } catch {
      continue;
    }

    // Find API versions
    let versions: string[];
    if (apiVersionCandidates.length > 0) {
      versions = apiVersionCandidates;
    } else {
      try {
        versions = await readdir(svcDir);
      } catch {
        continue;
      }
    }

    for (const ver of versions) {
      const pkg = join(svcDir, ver, resourceDir);
      try {
        await access(pkg);
        return pkg;
      } catch {}
    }
  }
  return null;
}

async function loadSdkPackage(pkgDir: string): Promise<{
  structs: Map<string, GoField[]>;
  enums: Map<string, string[]>;
  constants: Map<string, string[]>;
}> {
  const allStructs = new Map<string, GoField[]>();
  const allEnums = new Map<string, string[]>();
  const allConstants = new Map<string, string[]>();

  const files = await readdir(pkgDir);
  for (const f of files) {
    if (!f.endsWith(".go")) continue;
    const content = await readFile(join(pkgDir, f), "utf-8");
    if (f.startsWith("model_") || f === "model.go") {
      const parsed = parseGoStruct(content);
      for (const [k, v] of parsed) allStructs.set(k, v);
    }
    if (f.startsWith("constant") || f === "constants.go") {
      const parsed = parseEnumValues(content);
      for (const [k, v] of parsed) allEnums.set(k, v);
      // Also extract const declarations for types we might not get from PossibleValuesFor
      const constTypeRe =
        /type\s+(\w+)\s+string/g;
      let cm: RegExpExecArray | null;
      while ((cm = constTypeRe.exec(content)) !== null) {
        const typeName = cm[1];
        if (!allConstants.has(typeName)) {
          const vals: string[] = [];
          const valRe = new RegExp(
            `\\w+\\s+${typeName}\\s*=\\s*"([^"]+)"`,
            "g"
          );
          let vm: RegExpExecArray | null;
          while ((vm = valRe.exec(content)) !== null) {
            vals.push(vm[1]);
          }
          if (vals.length > 0) allConstants.set(typeName, vals);
        }
      }
    }
  }
  return { structs: allStructs, enums: allEnums, constants: allConstants };
}

// ---------------------------------------------------------------------------
// Determine root struct for a resource type
// ---------------------------------------------------------------------------

function findRootStruct(
  structs: Map<string, GoField[]>,
  resourceType: string
): string | null {
  const raw = (resourceType.split("/").pop() ?? "").replace(/s$/, "");
  const singular = raw.charAt(0).toUpperCase() + raw.slice(1);

  // Exact match: PUT body properties struct
  for (const c of [
    `${singular}PropertiesCreateParameters`,
    `${singular}PropertiesUpdateParameters`,
    `${singular}Properties`,
  ]) {
    if (structs.has(c)) return c;
  }

  // Fallback: case-insensitive search
  for (const [name] of structs) {
    if (
      name.toLowerCase().includes(singular.toLowerCase()) &&
      name.endsWith("PropertiesCreateParameters")
    ) {
      return name;
    }
  }
  for (const [name] of structs) {
    if (
      name.toLowerCase().includes(singular.toLowerCase()) &&
      name.endsWith("Properties") &&
      !name.includes("Patch") &&
      !name.includes("Update") &&
      !name.includes("AccessPolicy") &&
      !name.includes("Migration") &&
      !name.includes("Connection") &&
      !name.includes("Endpoint")
    ) {
      return name;
    }
  }
  return null;
}

// Find the GET/response properties struct (superset of create — includes computed fields)
function findResponseStruct(
  structs: Map<string, GoField[]>,
  resourceType: string
): string | null {
  const raw = (resourceType.split("/").pop() ?? "").replace(/s$/, "");
  const singular = raw.charAt(0).toUpperCase() + raw.slice(1);

  // Direct name: e.g. StorageAccountProperties, VaultProperties
  const direct = `${singular}Properties`;
  if (structs.has(direct)) return direct;

  // Fallback: case-insensitive, ending with Properties but not Create/Update/Patch
  for (const [name] of structs) {
    if (
      name.toLowerCase().includes(singular.toLowerCase()) &&
      name.endsWith("Properties") &&
      !name.includes("Create") &&
      !name.includes("Update") &&
      !name.includes("Patch") &&
      !name.includes("AccessPolicy") &&
      !name.includes("Migration") &&
      !name.includes("Connection") &&
      !name.includes("Endpoint")
    ) {
      return name;
    }
  }
  return null;
}

// Find the top-level resource struct that has a "properties" json tag
function findResourceStruct(
  structs: Map<string, GoField[]>,
  resourceType: string
): string | null {
  const resourceName = resourceType.split("/").pop() ?? "";
  const singular = resourceName.replace(/s$/, "");

  // Look for the struct that has json:"properties" — that's the top-level resource
  const candidates: string[] = [];
  for (const [name, fields] of structs) {
    if (
      name.toLowerCase().includes(singular.toLowerCase()) &&
      fields.some((f) => f.jsonTag === "properties")
    ) {
      candidates.push(name);
    }
  }

  // Prefer CreateParameters over the main model
  for (const c of candidates) {
    if (c.includes("Create")) return c;
  }
  return candidates[0] ?? null;
}

// ---------------------------------------------------------------------------
// Validation engine
// ---------------------------------------------------------------------------

function validateRules(
  rules: RuleEntry[],
  rootStruct: string,
  responseStruct: string | null,
  topStruct: string | null,
  structs: Map<string, GoField[]>,
  enums: Map<string, string[]>,
  constants: Map<string, string[]>
): Issue[] {
  const issues: Issue[] = [];

  for (const rule of rules) {
    let path = rule.propertyPath;

    // Skip array wildcard paths — can't trace through SDK structs without element type info
    if (path.includes("[*]")) continue;

    let startStruct: string;
    let propertiesPath = path;

    // Determine which struct to start from
    if (path.startsWith("properties.")) {
      startStruct = rootStruct;
      propertiesPath = path.slice("properties.".length);
    } else if (path.startsWith("sku.") || path.startsWith("extendedLocation") || path === "kind") {
      if (topStruct) {
        startStruct = topStruct;
        propertiesPath = path;
      } else {
        issues.push({
          severity: "warning",
          category: "path_resolution",
          propertyPath: rule.propertyPath,
          ruleKind: rule.kind,
          message: `Cannot resolve top-level path: no resource-level struct found`,
        });
        continue;
      }
    } else {
      startStruct = rootStruct;
    }

    let resolved = resolvePropertyPath(
      propertiesPath,
      startStruct,
      structs,
      enums,
      constants
    );

    // Fallback: if the create struct doesn't have the path, try the response struct
    // (some fields exist only on the response, or on sub-service structs)
    if (!resolved && responseStruct && startStruct !== responseStruct) {
      resolved = resolvePropertyPath(
        propertiesPath,
        responseStruct,
        structs,
        enums,
        constants
      );
    }

    if (!resolved) {
      issues.push({
        severity: "warning",
        category: "path_unresolved",
        propertyPath: rule.propertyPath,
        ruleKind: rule.kind,
        message: `Property path does not resolve in SDK models (may be a cross-service path)`,
        detail: `Tried ${startStruct}${responseStruct && startStruct !== responseStruct ? ` and ${responseStruct}` : ""}. Path may reference a sub-service API (e.g. BlobService, FileService).`,
      });
      continue;
    }
    if (rule.kind === "StringRule") {
      if (resolved.isNumeric) {
        issues.push({
          severity: "error",
          category: "type_mismatch",
          propertyPath: rule.propertyPath,
          ruleKind: rule.kind,
          message: `StringRule targets a numeric field (${resolved.goType})`,
          detail: `This should be an IntRule or FloatRule. Trace: ${resolved.traceLog.join(" → ")}`,
        });
      } else if (resolved.isBool) {
        issues.push({
          severity: "error",
          category: "type_mismatch",
          propertyPath: rule.propertyPath,
          ruleKind: rule.kind,
          message: `StringRule targets a boolean field (${resolved.goType})`,
          detail: `Boolean fields should not have StringRules. Trace: ${resolved.traceLog.join(" → ")}`,
        });
      } else if (resolved.isStruct && !resolved.isSlice) {
        issues.push({
          severity: "warning",
          category: "type_mismatch",
          propertyPath: rule.propertyPath,
          ruleKind: rule.kind,
          message: `StringRule targets a struct field (${resolved.goType}) — path may be incomplete`,
          detail: `Trace: ${resolved.traceLog.join(" → ")}`,
        });
      }

      // Enum completeness check
      if (rule.allowedValues && resolved.isEnum && resolved.enumValues) {
        const sdkSet = new Set(
          resolved.enumValues.map((v) => v.toLowerCase())
        );
        const ruleSet = new Set(
          rule.allowedValues.map((v) => v.toLowerCase())
        );

        const missingInRule: string[] = [];
        for (const v of resolved.enumValues) {
          if (!ruleSet.has(v.toLowerCase())) {
            missingInRule.push(v);
          }
        }
        const extraInRule: string[] = [];
        for (const v of rule.allowedValues) {
          if (!sdkSet.has(v.toLowerCase())) {
            extraInRule.push(v);
          }
        }

        if (missingInRule.length > 0) {
          issues.push({
            severity: "warning",
            category: "enum_incomplete",
            propertyPath: rule.propertyPath,
            ruleKind: rule.kind,
            message: `AllowedValues missing SDK values: ${missingInRule.join(", ")}`,
            detail: `SDK type ${resolved.goType} has: [${resolved.enumValues.join(", ")}]. Rule has: [${rule.allowedValues.join(", ")}]`,
          });
        }
        if (extraInRule.length > 0) {
          issues.push({
            severity: "warning",
            category: "enum_extra",
            propertyPath: rule.propertyPath,
            ruleKind: rule.kind,
            message: `AllowedValues has values not in SDK: ${extraInRule.join(", ")}`,
            detail: `SDK type ${resolved.goType} has: [${resolved.enumValues.join(", ")}]. Rule has: [${rule.allowedValues.join(", ")}]`,
          });
        }
      }
    } else if (rule.kind === "IntRule") {
      if (!resolved.isNumeric) {
        issues.push({
          severity: "error",
          category: "type_mismatch",
          propertyPath: rule.propertyPath,
          ruleKind: rule.kind,
          message: `IntRule targets a non-numeric field (${resolved.goType}, isEnum=${resolved.isEnum}, isString=${resolved.isString})`,
          detail: `This should be a StringRule or the PropertyPath is wrong. Trace: ${resolved.traceLog.join(" → ")}`,
        });
      }
    } else if (rule.kind === "FloatRule") {
      if (!resolved.isNumeric) {
        issues.push({
          severity: "error",
          category: "type_mismatch",
          propertyPath: rule.propertyPath,
          ruleKind: rule.kind,
          message: `FloatRule targets a non-numeric field (${resolved.goType})`,
          detail: `Trace: ${resolved.traceLog.join(" → ")}`,
        });
      }
    }
  }

  return issues;
}

function validatePaths(
  paths: string[],
  label: string,
  rootStruct: string,
  responseStruct: string | null,
  topStruct: string | null,
  structs: Map<string, GoField[]>,
  enums: Map<string, string[]>,
  constants: Map<string, string[]>
): Issue[] {
  const issues: Issue[] = [];

  for (const pp of paths) {
    let path = pp;

    // Skip array wildcard paths
    if (path.includes("[*]")) continue;

    let startStruct: string;

    if (path.startsWith("properties.")) {
      // ComputedFields should resolve against response struct (they're read-only, not in create)
      startStruct = label === "ComputedField" && responseStruct ? responseStruct : rootStruct;
      path = path.slice("properties.".length);
    } else if (topStruct) {
      startStruct = topStruct;
    } else {
      startStruct = rootStruct;
    }

    let resolved = resolvePropertyPath(
      path,
      startStruct,
      structs,
      enums,
      constants
    );

    // Fallback: try the other properties struct
    if (!resolved) {
      const fallback = startStruct === rootStruct ? responseStruct : rootStruct;
      if (fallback && fallback !== startStruct) {
        resolved = resolvePropertyPath(
          path.startsWith("properties.") ? path.slice("properties.".length) : path,
          fallback,
          structs,
          enums,
          constants
        );
      }
    }

    if (!resolved) {
      issues.push({
        severity: "warning",
        category: "path_unresolved",
        propertyPath: pp,
        ruleKind: label,
        message: `${label} path does not resolve in SDK models`,
        detail: `Tried ${startStruct}${responseStruct && startStruct !== responseStruct ? ` and ${responseStruct}` : ""}. May reference an external type or sub-service API.`,
      });
    }
  }

  return issues;
}

// ---------------------------------------------------------------------------
// Tool factory
// ---------------------------------------------------------------------------

const factory: CustomToolFactory = (pi) => ({
  name: "azwise_validate",
  label: "Azwise Validate",
  description:
    "Validates azwise knowledge files against ARM SDK model structs and enum constants. " +
    "Detects type mismatches (StringRule on int field), incomplete enum values, " +
    "invalid property paths, and more. Reads generated Go files from internal/azure/azwise/ " +
    "and cross-references with the SDK vendor directory in the azurerm repo.",
  parameters: pi.zod.object({
    azurerm_path: pi.zod
      .string()
      .optional()
      .describe("Path to azurerm repo root (auto-detected if omitted)"),
    resource: pi.zod
      .string()
      .optional()
      .describe(
        "Resource file to validate (e.g. 'storage' for storage.go). Omit to validate all."
      ),
  }),
  execute: async (args, { onUpdate, cwd }) => {
    const azurermPath =
      args.azurerm_path || (await resolveAzurermPath(cwd));
    if (!azurermPath) {
      return {
        error:
          "Cannot find azurerm repo. Set AZURERM_PATH or place the repo as a sibling directory.",
      };
    }

    const azwiseDir = join(cwd, "internal", "azure", "azwise");
    // Find knowledge files — all .go files that aren't azwise.go, helpers.go, register.go, *_test.go
    const allFiles = await readdir(azwiseDir);
    const knowledgeFiles = allFiles.filter(
      (f) =>
        f.endsWith(".go") &&
        !f.endsWith("_test.go") &&
        f !== "azwise.go" &&
        f !== "helpers.go" &&
        f !== "register.go"
    );

    const targetFiles = args.resource
      ? knowledgeFiles.filter(
          (f) => f === `${args.resource}.go` || f.startsWith(`${args.resource}_`)
        )
      : knowledgeFiles;

    if (targetFiles.length === 0) {
      return {
        error: `No knowledge files found${args.resource ? ` matching '${args.resource}'` : ""}`,
        available: knowledgeFiles,
      };
    }

    const results: Record<
      string,
      {
        resourceType: string;
        sdkPackage: string | null;
        issues: Issue[];
        rulesChecked: number;
        pathsChecked: number;
      }
    > = {};

    for (const file of targetFiles) {
      const source = await readFile(join(azwiseDir, file), "utf-8");
      const parsed = parseKnowledgeFile(source);

      if (!parsed.resourceType) {
        results[file] = {
          resourceType: "",
          sdkPackage: null,
          issues: [
            {
              severity: "error",
              category: "parse",
              propertyPath: "",
              ruleKind: "",
              message: "Could not extract ResourceType from file",
            },
          ],
          rulesChecked: 0,
          pathsChecked: 0,
        };
        continue;
      }

      onUpdate(
        `Validating ${file} (${parsed.resourceType}) — ${parsed.rules.length} rules, ${parsed.computedFields.length + parsed.forceNewPaths.length + parsed.defaultValuePaths.length} paths...`
      );

      const apiVersion = parsed.apiVersions[0] ?? "";
      const sdkPkg = await findSdkPackage(
        azurermPath,
        parsed.resourceType,
        apiVersion
      );

      if (!sdkPkg) {
        results[file] = {
          resourceType: parsed.resourceType,
          sdkPackage: null,
          issues: [
            {
              severity: "error",
              category: "sdk_not_found",
              propertyPath: "",
              ruleKind: "",
              message: `SDK package not found for ${parsed.resourceType} (api-version: ${apiVersion})`,
            },
          ],
          rulesChecked: 0,
          pathsChecked: 0,
        };
        continue;
      }

      const sdk = await loadSdkPackage(sdkPkg);
      const rootStruct = findRootStruct(sdk.structs, parsed.resourceType);
      const responseStruct = findResponseStruct(sdk.structs, parsed.resourceType);
      const topStruct = findResourceStruct(sdk.structs, parsed.resourceType);

      if (!rootStruct) {
        results[file] = {
          resourceType: parsed.resourceType,
          sdkPackage: sdkPkg,
          issues: [
            {
              severity: "error",
              category: "struct_not_found",
              propertyPath: "",
              ruleKind: "",
              message: `Cannot find root properties struct for ${parsed.resourceType} in SDK models`,
              detail: `Available structs: ${[...sdk.structs.keys()].join(", ")}`,
            },
          ],
          rulesChecked: 0,
          pathsChecked: 0,
        };
        continue;
      }

      const issues: Issue[] = [];

      // Validate rules (StringRules, IntRules, FloatRules)
      issues.push(
        ...validateRules(
          parsed.rules,
          rootStruct,
          responseStruct,
          topStruct,
          sdk.structs,
          sdk.enums,
          sdk.constants
        )
      );

      // Validate ComputedFields paths — use response struct (computed fields are read-only)
      issues.push(
        ...validatePaths(
          parsed.computedFields,
          "ComputedField",
          rootStruct,
          responseStruct,
          topStruct,
          sdk.structs,
          sdk.enums,
          sdk.constants
        )
      );

      // Validate ForceNew paths
      issues.push(
        ...validatePaths(
          parsed.forceNewPaths,
          "ForceNew",
          rootStruct,
          responseStruct,
          topStruct,
          sdk.structs,
          sdk.enums,
          sdk.constants
        )
      );

      // Validate DefaultValue paths
      issues.push(
        ...validatePaths(
          parsed.defaultValuePaths,
          "DefaultValue",
          rootStruct,
          responseStruct,
          topStruct,
          sdk.structs,
          sdk.enums,
          sdk.constants
        )
      );

      const pathsChecked =
        parsed.computedFields.length +
        parsed.forceNewPaths.length +
        parsed.defaultValuePaths.length;

      results[file] = {
        resourceType: parsed.resourceType,
        sdkPackage: sdkPkg,
        issues,
        rulesChecked: parsed.rules.length,
        pathsChecked,
      };
    }

    // Summary
    const totalIssues = Object.values(results).reduce(
      (sum, r) => sum + r.issues.length,
      0
    );
    const errors = Object.values(results).reduce(
      (sum, r) =>
        sum + r.issues.filter((i) => i.severity === "error").length,
      0
    );
    const warnings = Object.values(results).reduce(
      (sum, r) =>
        sum + r.issues.filter((i) => i.severity === "warning").length,
      0
    );

    return {
      summary: {
        filesValidated: Object.keys(results).length,
        totalIssues,
        errors,
        warnings,
      },
      results,
    };
  },
});

export default factory;
