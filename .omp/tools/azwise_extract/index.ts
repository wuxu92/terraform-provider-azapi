import type { CustomToolFactory } from "@oh-my-pi/pi-coding-agent";
import { readFile, access } from "node:fs/promises";
import { join, dirname, resolve, relative } from "node:path";

/**
 * Auto-resolve the azurerm repo path. Checks in order:
 * 1. AZURERM_PATH environment variable
 * 2. Sibling directory named terraform-provider-azurerm next to cwd
 * 3. go.work file in parent directories referencing the azurerm module
 * Returns the resolved path or null.
 */
async function resolveAzurermPath(cwd: string): Promise<string | null> {
  // 1. Environment variable
  const envPath = process.env.AZURERM_PATH;
  if (envPath) {
    const candidate = resolve(envPath);
    try {
      await access(join(candidate, "internal", "services"));
      return candidate;
    } catch {}
  }

  // 2. Sibling directory (../terraform-provider-azurerm relative to cwd)
  const sibling = resolve(cwd, "..", "terraform-provider-azurerm");
  try {
    await access(join(sibling, "internal", "services"));
    return sibling;
  } catch {}

  // 3. Walk up from cwd looking for go.work that references azurerm
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

const factory: CustomToolFactory = (pi) => ({
  name: "azwise_extract",
  label: "Azwise Extract",
  description:
    "Scans terraform-provider-azurerm source to extract knowledge patterns. Categories: forcenew, sensitive, timeouts, softdelete, relational (grep Go source — relational scans ConflictsWith/RequiredWith/ExactlyOneOf/AtLeastOneOf cross-property constraints), schema (reads provider-schema.json for field flags), mapping (d.Set/d.Get → ARM paths), validation (ValidateFunc patterns), automap (combines schema+mapping to produce ARM paths mechanically — handles snake_case→camelCase, block nesting, envelope exclusion). Use block= to narrow schema/mapping/validation/automap to a nested block subtree (e.g. block=blob_properties). The azurerm repo path is auto-detected.",
  parameters: pi.zod.object({
    azurerm_path: pi.zod.string().optional().describe("Path to azurerm repo root (auto-detected if omitted)"),
    category: pi.zod
      .enum(["forcenew", "sensitive", "timeouts", "softdelete", "relational", "schema", "mapping", "validation", "automap", "all"])
      .default("all")
      .describe("Which pattern category to extract"),
    service: pi.zod
      .string()
      .optional()
      .describe("Filter to one service package name (for grep-based categories)"),
    resource_name: pi.zod
      .string()
      .optional()
      .describe("AzureRM resource name for schema/mapping/validation/automap, e.g. 'azurerm_key_vault'. Enables per-resource mode with cross-referencing."),
    block: pi.zod
      .string()
      .optional()
      .describe("Narrow to a nested block subtree, e.g. 'blob_properties' or 'network_rules.private_link_access'. Only affects schema/mapping/validation/automap with resource_name."),
  }),

  async execute(toolCallId, params, onUpdate, ctx, signal) {
    let azurerm_path = params.azurerm_path;
    if (!azurerm_path) {
      azurerm_path = await resolveAzurermPath(pi.cwd);
      if (!azurerm_path) {
        return {
          content: [{ type: "text", text:
            "Error: Could not find terraform-provider-azurerm.\n" +
            "Set the AZURERM_PATH environment variable, place the repo as a sibling directory, " +
            "or pass azurerm_path explicitly." }],
        };
      }
    }
    const { category, service, resource_name, block } = params;
    const servicesDir = join(azurerm_path, "internal", "services");
    try {
      await access(servicesDir);
    } catch {
      return { content: [{ type: "text", text: `Error: ${servicesDir} not found` }] };
    }

    const searchPath = service ? join(servicesDir, service) : servicesDir;
    const categories =
      category === "all"
        ? (["forcenew", "sensitive", "timeouts", "softdelete", "relational", "schema", "mapping", "validation", "automap"] as const)
        : ([category] as const);

    // Run grep, returning matched lines. Exit code 1 = no matches (not an error).
    async function grep(pattern: string, path: string): Promise<string[]> {
      try {
        const result = await pi.exec(
          "grep",
          ["-rn", "--include=*.go", pattern, path],
          { signal },
        );
        return result.stdout
          .split("\n")
          .filter(
            (l) =>
              l &&
              !l.includes("_test.go") &&
              !l.includes("/vendor/") &&
              !l.includes("/testdata/") &&
              !l.includes("/migration/"),
          );
      } catch {
        return []; // exit code 1 = no matches
      }
    }

    // Parse "filepath:linenum:content" from grep output.
    function parseGrepLine(line: string) {
      const idx1 = line.indexOf(":");
      const idx2 = line.indexOf(":", idx1 + 1);
      if (idx1 < 0 || idx2 < 0) return null;
      return {
        file: line.slice(0, idx1),
        lineNum: parseInt(line.slice(idx1 + 1, idx2), 10),
        content: line.slice(idx2 + 1),
      };
    }

    // Derive service name from file path.
    function serviceFromPath(filePath: string): string {
      const rel = relative(servicesDir, filePath);
      return rel.split("/")[0] || "unknown";
    }

    // Given file content lines and a match line number, scan backward tracking
    // braces to find the enclosing schema field name: `"field_name": {`.
    function extractFieldName(lines: string[], matchLine: number): string {
      let depth = 0;
      for (let i = matchLine - 1; i >= 0; i--) {
        const line = lines[i];
        for (let c = 0; c < line.length; c++) {
          if (line[c] === "}") depth++;
          else if (line[c] === "{") depth--;
        }
        if (depth < 0) {
          // Check this line for "field_name": {
          const m = line.match(/"(\w+)"\s*:\s*\{/);
          if (m) return m[1];
          // Check preceding lines for multi-line field defs: "field_name":\n{
          for (let k = i - 1; k >= Math.max(0, i - 3); k--) {
            const pm = lines[k].match(/"(\w+)"\s*:\s*$/);
            if (pm) return pm[1];
          }
          return "unknown";
        }
      }
      return "unknown";
    }

    // Cache for file reads to avoid reading the same file multiple times.
    const fileCache = new Map<string, string[]>();
    async function getFileLines(filePath: string): Promise<string[]> {
      let lines = fileCache.get(filePath);
      if (!lines) {
        const text = await readFile(filePath, "utf-8");
        lines = text.split("\n");
        fileCache.set(filePath, lines);
      }
      return lines;
    }

    // Find the Go source file for a given azurerm resource name.
    async function findResourceFile(name: string): Promise<string | null> {
      const fileName = name.replace(/^azurerm_/, "") + "_resource.go";
      try {
        const result = await pi.exec(
          "find", [servicesDir, "-name", fileName, "-not", "-path", "*/migration/*"],
          { signal },
        );
        const found = result.stdout.trim().split("\n").filter(Boolean);
        return found.length > 0 ? found[0] : null;
      } catch { return null; }
    }

    // --- ForceNew / Sensitive extraction ---
    async function extractFieldAnnotations(pattern: string, label: string) {
      onUpdate?.({ content: [{ type: "text", text: `Scanning for ${label} patterns...` }] });
      const grepLines = await grep(pattern, searchPath);

      // Group grep matches by file.
      const byFile = new Map<string, number[]>();
      for (const raw of grepLines) {
        const parsed = parseGrepLine(raw);
        if (!parsed) continue;
        let arr = byFile.get(parsed.file);
        if (!arr) {
          arr = [];
          byFile.set(parsed.file, arr);
        }
        arr.push(parsed.lineNum);
      }

      // Extract field names by reading files with matches.
      const byService = new Map<
        string,
        { relativePath: string; fields: string[] }[]
      >();
      let totalAnnotations = 0;

      for (const [filePath, lineNums] of byFile) {
        const lines = await getFileLines(filePath);
        const fields: string[] = [];
        for (const ln of lineNums) {
          fields.push(extractFieldName(lines, ln - 1)); // grep is 1-indexed
        }
        totalAnnotations += fields.length;

        const svc = serviceFromPath(filePath);
        let arr = byService.get(svc);
        if (!arr) {
          arr = [];
          byService.set(svc, arr);
        }
        arr.push({ relativePath: relative(azurerm_path, filePath), fields });
      }

      const services = [...byService.entries()]
        .sort((a, b) => a[0].localeCompare(b[0]))
        .map(([svc, files]) => ({
          service: svc,
          files,
          totalFields: files.reduce((s, f) => s + f.fields.length, 0),
        }));

      return { totalFiles: byFile.size, totalAnnotations, services };
    }

    // --- Timeout extraction ---
    async function extractTimeouts() {
      onUpdate?.({ content: [{ type: "text", text: "Scanning for timeout patterns..." }] });
      const grepLines = await grep("DefaultTimeout", searchPath);

      const re =
        /(\w+):\s*(?:pluginsdk|schema)\.DefaultTimeout\((\d+)\s*\*\s*time\.(Minute|Hour|Second)\)/;

      const byFile = new Map<
        string,
        { create?: string; read?: string; update?: string; delete?: string }
      >();

      for (const raw of grepLines) {
        const parsed = parseGrepLine(raw);
        if (!parsed) continue;
        const m = parsed.content.match(re);
        if (!m) continue;

        const op = m[1].toLowerCase() as "create" | "read" | "update" | "delete";
        const val = parseInt(m[2], 10);
        const unit = m[3];
        const formatted =
          unit === "Hour"
            ? `${val}h`
            : unit === "Minute"
              ? `${val}m`
              : `${val}s`;

        let entry = byFile.get(parsed.file);
        if (!entry) {
          entry = {};
          byFile.set(parsed.file, entry);
        }
        entry[op] = formatted;
      }

      const byService = new Map<
        string,
        {
          relativePath: string;
          create?: string;
          read?: string;
          update?: string;
          delete?: string;
        }[]
      >();

      for (const [filePath, timeouts] of byFile) {
        const svc = serviceFromPath(filePath);
        let arr = byService.get(svc);
        if (!arr) {
          arr = [];
          byService.set(svc, arr);
        }
        arr.push({ relativePath: relative(azurerm_path, filePath), ...timeouts });
      }

      const services = [...byService.entries()]
        .sort((a, b) => a[0].localeCompare(b[0]))
        .map(([svc, files]) => ({ service: svc, files }));

      return { totalFiles: byFile.size, services };
    }

    // --- Soft-delete extraction ---
    async function extractSoftDelete() {
      onUpdate?.({ content: [{ type: "text", text: "Scanning for soft-delete patterns..." }] });
      const grepLines = await grep(
        "PurgeSoftDelete\\|purge.*Deleted\\|RecoverSoftDeleted",
        searchPath,
      );

      const byFile = new Map<string, string[]>();
      for (const raw of grepLines) {
        const parsed = parseGrepLine(raw);
        if (!parsed) continue;
        // Extract the matching function/method name from the line.
        const trimmed = parsed.content.trim();
        let arr = byFile.get(parsed.file);
        if (!arr) {
          arr = [];
          byFile.set(parsed.file, arr);
        }
        arr.push(trimmed);
      }

      const byService = new Map<
        string,
        { relativePath: string; patterns: string[] }[]
      >();

      for (const [filePath, patterns] of byFile) {
        const svc = serviceFromPath(filePath);
        let arr = byService.get(svc);
        if (!arr) {
          arr = [];
          byService.set(svc, arr);
        }
        arr.push({ relativePath: relative(azurerm_path, filePath), patterns });
      }

      const services = [...byService.entries()]
        .sort((a, b) => a[0].localeCompare(b[0]))
        .map(([svc, files]) => ({ service: svc, files }));

      return { totalFiles: byFile.size, services };
    }

    // --- Schema extraction from provider-schema.json ---
    async function extractSchema() {
      onUpdate?.({ content: [{ type: "text", text: "Reading provider-schema.json..." }] });
      const schemaPath = join(azurerm_path, ".release", "provider-schema.json");
      let raw: string;
      try {
        raw = await readFile(schemaPath, "utf-8");
      } catch {
        return { error: `${schemaPath} not found. Run 'make generate' in the azurerm repo first.` };
      }
      const schema = JSON.parse(raw);
      const resources: Record<string, any> = schema.providerSchema?.resources ?? {};

      // If a specific resource is requested, return its full schema.
      if (resource_name) {
        const r = resources[resource_name];
        if (!r) {
          // Fuzzy match: try prefix
          const matches = Object.keys(resources).filter(k => k.includes(resource_name));
          return {
            error: `Resource '${resource_name}' not found.`,
            suggestions: matches.slice(0, 10),
          };
        }
        // Classify fields
        const computed: string[] = [];
        const defaults: string[] = [];
        const forceNew: string[] = [];
        const sensitive: string[] = [];
        const required: string[] = [];
        const optional: string[] = [];

        function classifyFields(fields: Record<string, any>, prefix = "") {
          for (const [name, f] of Object.entries(fields)) {
            const path = prefix ? `${prefix}.${name}` : name;
            if (f.computed && !f.optional && !f.required) computed.push(path);
            if (f.computed && f.optional) defaults.push(path);
            if (f.forceNew) forceNew.push(path);
            if (f.sensitive) sensitive.push(path);
            if (f.required) required.push(path);
            if (f.optional && !f.computed) optional.push(path);
            // Recurse into nested schema (TypeList/TypeSet with object elem)
            if (f.elem?.schema) {
              classifyFields(f.elem.schema, path);
            }
          }
        }

        // When block= is set, navigate to the sub-schema for that block.
        let classifyRoot = r.schema || {};
        let classifyPrefix = "";
        if (block) {
          const parts = block.split(".");
          let cursor: any = r.schema || {};
          for (const part of parts) {
            const field = cursor[part];
            if (!field || !field.elem?.schema) {
              return { error: `Block '${block}' not found in ${resource_name}. Available top-level blocks: ${Object.keys(r.schema || {}).filter(k => (r.schema || {})[k]?.elem?.schema).join(", ")}` };
            }
            cursor = field.elem.schema;
          }
          classifyRoot = cursor;
          classifyPrefix = block;
        }
        classifyFields(classifyRoot, classifyPrefix);

        // Extract API version from the resource's Go source file imports.
        let apiVersions: string[] = [];
        const resourceFilePath = await findResourceFile(resource_name);
        if (resourceFilePath) {
          try {
            const srcContent = await readFile(resourceFilePath, "utf-8");
            // Match ARM API versions from import paths like resource-manager/keyvault/2023-02-01/vaults.
            // Data-plane versions (e.g. 7-4) are intentionally excluded — they have no ARM equivalent.
            const versionRe = /(?:resource-manager|data-plane)\/[\w.-]+\/(20\d{2}-\d{2}-\d{2}(?:-preview)?)\//g;
            let vm: RegExpExecArray | null;
            const vset = new Set<string>();
            while ((vm = versionRe.exec(srcContent)) !== null) vset.add(vm[1]);
            apiVersions = [...vset].sort();
          } catch {}
        }

        // Collect block names for discoverability
        const blockNames: string[] = [];
        function collectBlocks(fields: Record<string, any>, prefix = "") {
          for (const [name, f] of Object.entries(fields)) {
            const path = prefix ? `${prefix}.${name}` : name;
            if (f.elem?.schema) {
              blockNames.push(path);
              collectBlocks(f.elem.schema, path);
            }
          }
        }
        if (!block) collectBlocks(r.schema || {});

        return {
          resource: resource_name,
          ...(block ? { block } : {}),
          resourceFilePath: resourceFilePath ? relative(azurerm_path, resourceFilePath) : null,
          apiVersions,
          totalFields: Object.keys(classifyRoot).length,
          computed,
          defaults,
          forceNew,
          sensitive,
          required,
          optional,
          ...(blockNames.length > 0 ? { blocks: blockNames } : {}),
          rawSchema: block ? classifyRoot : r.schema,
        };
      }

      // No specific resource: if service filter is set, list resources matching that service.
      // azurerm resource names follow the pattern azurerm_<service>_...
      const filtered = service
        ? Object.keys(resources).filter(k => k.startsWith(`azurerm_${service}`))
        : Object.keys(resources);

      return {
        totalResources: filtered.length,
        resources: filtered.sort(),
      };
    }

    // --- Mapping extraction: d.Set/d.Get -> ARM property paths ---
    async function extractMapping() {
      onUpdate?.({ content: [{ type: "text", text: "Extracting Terraform → ARM field mappings..." }] });

      interface FieldMapping {
        terraformField: string;
        sdkPath: string;     // Go SDK access path as written
        armPath: string;     // Inferred ARM JSON path
        pattern: string;     // Which extraction pattern matched
        line: number;
      }

      function goFieldToArm(goField: string): string {
        return goField[0].toLowerCase() + goField.slice(1);
      }

      function addMapping(map: Map<string, FieldMapping[]>, file: string, mapping: FieldMapping) {
        let arr = map.get(file);
        if (!arr) { arr = []; map.set(file, arr); }
        if (!arr.some(x => x.terraformField === mapping.terraformField && x.armPath === mapping.armPath)) {
          arr.push(mapping);
        }
      }

      // Detect property aliases in file content: `props := model.Properties` etc.
      // Returns a set of variable names that alias model.Properties.
      function detectPropertyAliases(content: string): Set<string> {
        const aliases = new Set<string>();
        // Patterns: `props := model.Properties`, `if props := model.Properties;`,
        //           `props := account.Properties`, `props := resp.Model.Properties`
        const re = /(\w+)\s*:=\s*(?:\w+\.)?(?:Model\.)?Properties\b/g;
        let m: RegExpExecArray | null;
        while ((m = re.exec(content)) !== null) {
          const name = m[1];
          if (name !== "if" && name !== "for" && name !== "var") aliases.add(name);
        }
        return aliases;
      }

      // Extract expand/flatten function references from file content.
      interface ExpandFlattenRef {
        name: string;
        terraformField: string | null;  // associated d.Set/d.Get field, if detectable
        direction: "expand" | "flatten";
        line: number;
      }

      function findExpandFlatten(lines: string[]): ExpandFlattenRef[] {
        const refs: ExpandFlattenRef[] = [];
        const seen = new Set<string>();
        for (let i = 0; i < lines.length; i++) {
          const line = lines[i];
          // flatten calls in d.Set context
          let m = line.match(/d\.Set\("(\w+)",\s*(flatten\w+)\(/);
          if (m && !seen.has(m[2])) {
            seen.add(m[2]);
            refs.push({ name: m[2], terraformField: m[1], direction: "flatten", line: i + 1 });
            continue;
          }
          // flatten calls assigned to variable before d.Set
          m = line.match(/(?:=|:=)\s*(flatten\w+)\(/);
          if (m && !seen.has(m[1])) {
            seen.add(m[1]);
            // Check next few lines for d.Set
            let field: string | null = null;
            for (let j = i + 1; j < Math.min(i + 5, lines.length); j++) {
              const sm = lines[j].match(/d\.Set\("(\w+)"/);
              if (sm) { field = sm[1]; break; }
            }
            refs.push({ name: m[1], terraformField: field, direction: "flatten", line: i + 1 });
            continue;
          }
          // expand calls
          m = line.match(/(expand\w+)\(/);
          if (m && !seen.has(m[1])) {
            seen.add(m[1]);
            // Check same line or nearby for d.Get
            let field: string | null = null;
            const gm = line.match(/d\.Get\("(\w+)"\)/);
            if (gm) field = gm[1];
            refs.push({ name: m[1], terraformField: field, direction: "expand", line: i + 1 });
          }
        }
        return refs;
      }

      // Apply all d.Set/d.Get patterns to a single line, returning the mapping if matched.
      function matchLine(
        line: string, lineNum: number,
        propAliases: Set<string>,
      ): FieldMapping | null {
        let m: RegExpMatchArray | null;

        // 1. d.Set("field", model.Properties.GoField) — direct property access
        m = line.match(/d\.Set\("(\w+)",\s*(?:string\()?(?:pointer\.From\()?model\.Properties\.(\w+)/);
        if (m) {
          const arm = "properties." + goFieldToArm(m[2]);
          return { terraformField: m[1], sdkPath: `model.Properties.${m[2]}`, armPath: arm, pattern: "model.Properties.X", line: lineNum };
        }

        // 2. d.Set("field", props.GoField) — aliased property access
        for (const alias of propAliases) {
          const re = new RegExp(`d\\.Set\\("(\\w+)",\\s*(?:string\\()?(?:pointer\\.From\\()?${alias}\\.(\\w+)`);
          m = line.match(re);
          if (m) {
            const arm = "properties." + goFieldToArm(m[2]);
            return { terraformField: m[1], sdkPath: `${alias}.${m[2]}`, armPath: arm, pattern: "alias.X", line: lineNum };
          }
        }

        // 3. d.Set("field", flatten...(model.Properties.GoField))
        m = line.match(/d\.Set\("(\w+)",\s*flatten\w+\((?:&)?model\.Properties\.(\w+)/);
        if (m) {
          const arm = "properties." + goFieldToArm(m[2]);
          return { terraformField: m[1], sdkPath: `flatten(model.Properties.${m[2]})`, armPath: arm, pattern: "flatten(Properties.X)", line: lineNum };
        }

        // 4. d.Set("field", flatten...(props.GoField)) — flatten with alias
        for (const alias of propAliases) {
          const re = new RegExp(`d\\.Set\\("(\\w+)",\\s*flatten\\w+\\((?:&)?${alias}\\.(\\w+)`);
          m = line.match(re);
          if (m) {
            const arm = "properties." + goFieldToArm(m[2]);
            return { terraformField: m[1], sdkPath: `flatten(${alias}.${m[2]})`, armPath: arm, pattern: "flatten(alias.X)", line: lineNum };
          }
        }

        // 5. d.Set("field", model.Properties.Parent.Child) — nested property access
        m = line.match(/d\.Set\("(\w+)",\s*(?:string\()?(?:pointer\.From\()?model\.Properties\.(\w+)\.(\w+)/);
        if (m) {
          const arm = "properties." + goFieldToArm(m[2]) + "." + goFieldToArm(m[3]);
          return { terraformField: m[1], sdkPath: `model.Properties.${m[2]}.${m[3]}`, armPath: arm, pattern: "model.Properties.X.Y", line: lineNum };
        }

        // 6. d.Set("field", key.GoField) — data-plane key patterns
        m = line.match(/d\.Set\("(\w+)",\s*(?:string\()?(?:pointer\.From\()?key\.(\w+)/);
        if (m) {
          const arm = "properties.key." + goFieldToArm(m[2]);
          return { terraformField: m[1], sdkPath: `key.${m[2]}`, armPath: arm, pattern: "key.X", line: lineNum };
        }

        // 7. d.Set("field", attributes.GoField)
        m = line.match(/d\.Set\("(\w+)",\s*(?:string\()?(?:pointer\.From\()?attributes\.(\w+)/);
        if (m) {
          const arm = "properties.attributes." + goFieldToArm(m[2]);
          return { terraformField: m[1], sdkPath: `attributes.${m[2]}`, armPath: arm, pattern: "attributes.X", line: lineNum };
        }

        // 8. d.Set("field", model.GoField) — top-level model fields (Location, Tags, etc.)
        m = line.match(/d\.Set\("(\w+)",\s*(?:string\()?(?:pointer\.From\()?(?:location\.Normalize(?:Nilable)?\()?model\.(\w+)/);
        if (m && !m[2].startsWith("Properties") && !m[2].startsWith("Key")) {
          const arm = goFieldToArm(m[2]);
          return { terraformField: m[1], sdkPath: `model.${m[2]}`, armPath: arm, pattern: "model.X", line: lineNum };
        }

        // 9. d.Set("field", account.GoField) — alternate model variable names
        m = line.match(/d\.Set\("(\w+)",\s*(?:string\()?(?:pointer\.From\()?(?:location\.Normalize(?:Nilable)?\()?account\.(\w+)/);
        if (m && !m[2].startsWith("Properties") && !m[2].startsWith("Model")) {
          const arm = goFieldToArm(m[2]);
          return { terraformField: m[1], sdkPath: `account.${m[2]}`, armPath: arm, pattern: "model.X", line: lineNum };
        }

        return null;
      }

      // --- Per-resource mode ---
      if (resource_name) {
        const filePath = await findResourceFile(resource_name);
        if (!filePath) {
          return { error: `Resource file not found for ${resource_name}` };
        }

        const content = await readFile(filePath, "utf-8");
        const lines = content.split("\n");
        const propAliases = detectPropertyAliases(content);
        const mappings: FieldMapping[] = [];
        const seen = new Set<string>();

        // Process all d.Set lines
        for (let i = 0; i < lines.length; i++) {
          const line = lines[i];
          if (!line.includes('d.Set("')) continue;
          const mapping = matchLine(line, i + 1, propAliases);
          if (mapping) {
            const key = `${mapping.terraformField}:${mapping.armPath}`;
            if (!seen.has(key)) {
              seen.add(key);
              mappings.push(mapping);
            }
          }
        }

        // Process d.Get lines for Create/Update direction
        for (let i = 0; i < lines.length; i++) {
          const line = lines[i];
          if (!line.includes('d.Get("')) continue;
          // .Properties.GoField = ... d.Get("field")
          const m = line.match(/\.Properties\.(\w+)\s*=.*d\.Get\("(\w+)"\)/);
          if (m) {
            const arm = "properties." + goFieldToArm(m[1]);
            const key = `${m[2]}:${arm}`;
            if (!seen.has(key)) {
              seen.add(key);
              mappings.push({ terraformField: m[2], sdkPath: `Properties.${m[1]}`, armPath: arm, pattern: "Properties.X = d.Get(Y)", line: i + 1 });
            }
          }
        }

        // Discover expand/flatten functions
        const expandFlatten = findExpandFlatten(lines);

        // Cross-reference with schema to find unmapped fields
        let unmappedFields: string[] = [];
        let schemaFieldCount = 0;
        try {
          const schemaPath = join(azurerm_path, ".release", "provider-schema.json");
          const schemaRaw = await readFile(schemaPath, "utf-8");
          const schema = JSON.parse(schemaRaw);
          const resource = schema.providerSchema?.resources?.[resource_name];
          if (resource?.schema) {
            // When block= is set, navigate to the sub-schema
            let fieldSource = resource.schema;
            let fieldPrefix = "";
            if (block) {
              const parts = block.split(".");
              for (const part of parts) {
                const field = fieldSource[part];
                if (field?.elem?.schema) {
                  fieldSource = field.elem.schema;
                } else {
                  fieldSource = {};
                  break;
                }
              }
              fieldPrefix = block + ".";
            }
            const allFields = new Set(Object.keys(fieldSource).map(f => fieldPrefix + f));
            schemaFieldCount = allFields.size;
            const mappedFields = new Set(mappings.map(m => m.terraformField));
            // Also count fields associated with expand/flatten functions
            for (const ef of expandFlatten) {
              if (ef.terraformField) mappedFields.add(ef.terraformField);
            }
            // Exclude identity/meta fields that don't map to ARM body
            const metaFields = new Set(["id", "name", "resource_group_name", "location", "tags", "timeouts"]);
            unmappedFields = [...allFields].filter(f => !mappedFields.has(f) && !metaFields.has(f)).sort();
          }
        } catch {}

        // When block= is set, filter mappings and expand/flatten to just that block
        let filteredMappings = mappings;
        let filteredExpandFlatten = expandFlatten;
        if (block) {
          filteredMappings = mappings.filter(m => m.terraformField.startsWith(block + ".") || m.terraformField === block);
          filteredExpandFlatten = expandFlatten.filter(ef => ef.terraformField && (ef.terraformField.startsWith(block + ".") || ef.terraformField === block || ef.name.toLowerCase().includes(block.replace(/_/g, "").toLowerCase())));
        }

        return {
          resource: resource_name,
          ...(block ? { block } : {}),
          resourceFile: relative(azurerm_path, filePath),
          propertyAliases: [...propAliases],
          mappings: filteredMappings,
          expandFlattenFunctions: filteredExpandFlatten,
          unmappedFields,
          totalMapped: filteredMappings.length,
          totalUnmapped: unmappedFields.length,
          schemaFieldCount,
        };
      }

      // --- Service-wide mode (original behavior) ---
      const setLines = await grep('d\\.Set("', searchPath);
      const getLines = await grep('d\\.Get("', searchPath);
      const byFile = new Map<string, FieldMapping[]>();

      for (const raw of setLines) {
        const parsed = parseGrepLine(raw);
        if (!parsed) continue;
        const mapping = matchLine(parsed.content, parsed.lineNum, new Set());
        if (mapping) addMapping(byFile, parsed.file, mapping);
      }

      for (const raw of getLines) {
        const parsed = parseGrepLine(raw);
        if (!parsed) continue;
        const { file, lineNum, content: line } = parsed;
        const m = line.match(/\.Properties\.(\w+)\s*=.*d\.Get\("(\w+)"\)/);
        if (m) {
          const arm = "properties." + goFieldToArm(m[1]);
          addMapping(byFile, file, { terraformField: m[2], sdkPath: `Properties.${m[1]}`, armPath: arm, pattern: "Properties.X = d.Get(Y)", line: lineNum });
        }
      }

      // Group by service
      const byService = new Map<string, { relativePath: string; mappings: FieldMapping[] }[]>();
      for (const [filePath, mappings] of byFile) {
        const svc = serviceFromPath(filePath);
        let arr = byService.get(svc);
        if (!arr) { arr = []; byService.set(svc, arr); }
        arr.push({ relativePath: relative(azurerm_path, filePath), mappings });
      }

      const services = [...byService.entries()]
        .sort((a, b) => a[0].localeCompare(b[0]))
        .map(([svc, files]) => ({ service: svc, files }));

      const totalMappings = [...byFile.values()].reduce((sum, arr) => sum + arr.length, 0);
      return { totalFiles: byFile.size, totalMappings, services };
    }

    // --- Validation extraction: ValidateFunc / ValidateDiagFunc patterns ---
    async function extractValidation() {
      onUpdate?.({ content: [{ type: "text", text: "Extracting validation patterns..." }] });

      interface ValidationRule {
        field: string;          // schema field name (e.g. "sku_name")
        type: string;           // "StringInSlice" | "IntBetween" | "StringLenBetween" | "StringMatch" | "IsUUID" | "custom"
        values?: string[];      // for StringInSlice
        min?: number;
        max?: number;
        regex?: string;
        caseSensitive?: boolean;
        reference?: string;     // for custom validators (function name)
        line: number;
      }

      // Per-resource mode: read the file, parse schema blocks with their field names
      if (resource_name) {
        const filePath = await findResourceFile(resource_name);
        if (!filePath) {
          return { error: `Resource file not found for ${resource_name}` };
        }

        const content = await readFile(filePath, "utf-8");
        const lines = content.split("\n");
        const rules: ValidationRule[] = [];

        for (let i = 0; i < lines.length; i++) {
          const line = lines[i];
          if (!line.includes("ValidateFunc:") && !line.includes("ValidateDiagFunc:")) continue;

          const field = extractFieldName(lines, i);

          // validation.StringInSlice([]string{...}, bool) — single or multi-line
          let m = line.match(/validation\.StringInSlice\(\[\]string\{([^}]*)\}/);
          if (!m && line.match(/validation\.StringInSlice\(\[\]string\{/)) {
            // Multi-line: collect lines until closing brace
            let valStr = line.replace(/.*validation\.StringInSlice\(\[\]string\{/, "");
            for (let j = i + 1; j < Math.min(i + 30, lines.length); j++) {
              const jline = lines[j].trim();
              valStr += " " + jline;
              if (jline.includes("}")) break;
            }
            const vals = [...valStr.matchAll(/"([^"]+)"/g)].map(x => x[1]);
            const constVals = [...valStr.matchAll(/string\((\w+\.\w+)\)/g)].map(x => x[1]);
            const caseFull = valStr + lines.slice(i, Math.min(i + 5, lines.length)).join(" ");
            const caseSens = !caseFull.includes(", true)") && !caseFull.includes(",true)");
            rules.push({
              field, type: "StringInSlice",
              values: vals.length > 0 ? vals : undefined,
              reference: constVals.length > 0 ? constVals.join(", ") : undefined,
              caseSensitive: caseSens,
              line: i + 1,
            });
            continue;
          }
          if (m) {
            const vals = [...m[1].matchAll(/"([^"]+)"/g)].map(x => x[1]);
            const constVals = [...m[1].matchAll(/string\((\w+\.\w+)\)/g)].map(x => x[1]);
            const caseSens = !line.includes(", true)") && !line.includes(",true)");
            rules.push({
              field, type: "StringInSlice",
              values: vals.length > 0 ? vals : undefined,
              reference: constVals.length > 0 ? constVals.join(", ") : undefined,
              caseSensitive: caseSens,
              line: i + 1,
            });
            continue;
          }

          // PossibleValuesFor...() — SDK-generated enum lists
          m = line.match(/validation\.StringInSlice\((\w+\.PossibleValuesFor\w+)\(\)/);
          if (m) {
            rules.push({ field, type: "StringInSlice", reference: m[1], line: i + 1 });
            continue;
          }

          // validation.IntBetween(min, max)
          m = line.match(/validation\.IntBetween\((\d+),\s*(\d+)\)/);
          if (m) {
            rules.push({ field, type: "IntBetween", min: parseInt(m[1]), max: parseInt(m[2]), line: i + 1 });
            continue;
          }

          // validation.FloatBetween(min, max)
          m = line.match(/validation\.FloatBetween\(([\d.]+),\s*([\d.]+)\)/);
          if (m) {
            rules.push({ field, type: "FloatBetween", min: parseFloat(m[1]), max: parseFloat(m[2]), line: i + 1 });
            continue;
          }

          // validation.StringLenBetween(min, max)
          m = line.match(/validation\.StringLenBetween\((\d+),\s*(\d+)\)/);
          if (m) {
            rules.push({ field, type: "StringLenBetween", min: parseInt(m[1]), max: parseInt(m[2]), line: i + 1 });
            continue;
          }

          // validation.StringMatch(regexp.MustCompile("..."), "...")
          m = line.match(/validation\.StringMatch\(regexp\.MustCompile\(["`]([^"`]+)["`]\)/);
          if (m) {
            rules.push({ field, type: "StringMatch", regex: m[1], line: i + 1 });
            continue;
          }

          // validation.IsUUID
          if (line.includes("validation.IsUUID")) {
            rules.push({ field, type: "IsUUID", line: i + 1 });
            continue;
          }

          // validation.StringIsNotEmpty
          if (line.includes("validation.StringIsNotEmpty")) {
            rules.push({ field, type: "StringIsNotEmpty", line: i + 1 });
            continue;
          }

          // Custom/external validator — capture the function reference
          m = line.match(/Validate(?:Func|DiagFunc):\s*(\S+)/);
          if (m) {
            const ref = m[1].replace(/,\s*$/, "");
            rules.push({ field, type: "custom", reference: ref, line: i + 1 });
          }
        }

        // When block= is set, filter rules to fields inside that block
        const filteredRules = block
          ? rules.filter(r => r.field.startsWith(block + ".") || r.field === block)
          : rules;

        return {
          resource: resource_name,
          ...(block ? { block } : {}),
          resourceFile: relative(azurerm_path, filePath),
          rules: filteredRules,
          totalRules: filteredRules.length,
        };
      }

      // Service-wide mode: grep-based summary (counts only)
      const validateLines = await grep("Validate\\(Func\\|DiagFunc\\):", searchPath);
      const byService = new Map<string, number>();
      for (const raw of validateLines) {
        const parsed = parseGrepLine(raw);
        if (!parsed) continue;
        const svc = serviceFromPath(parsed.file);
        byService.set(svc, (byService.get(svc) || 0) + 1);
      }
      const services = [...byService.entries()]
        .sort((a, b) => a[0].localeCompare(b[0]))
        .map(([svc, count]) => ({ service: svc, validationCount: count }));
      return { totalFiles: byService.size, services };
    }
    // --- Automap: mechanically produce ARM paths for schema fields ---
    async function extractAutomap() {
      if (!resource_name) {
        return { error: "automap requires resource_name (e.g. resource_name='azurerm_storage_account')" };
      }
      onUpdate?.({ content: [{ type: "text", text: "Auto-mapping Terraform fields to ARM paths..." }] });

      // 1. Get schema classification
      const schemaResult = await extractSchema();
      if ("error" in schemaResult) return schemaResult;

      // 2. Get mapping data (d.Set/d.Get extractions)
      const mappingResult = await extractMapping();
      const existingMappings = new Map<string, string>();
      if (!("error" in mappingResult) && "mappings" in mappingResult) {
        for (const m of mappingResult.mappings as any[]) {
          existingMappings.set(m.terraformField, m.armPath);
        }
      }

      // 3. Mechanical snake_case → camelCase
      function snakeToCamel(s: string): string {
        return s.replace(/_([a-z])/g, (_, c) => c.toUpperCase());
      }

      // Envelope / meta fields that don't map to ARM body properties
      const envelopeFields = new Set([
        "id", "name", "resource_group_name", "location", "tags", "timeouts",
        "identity", // handled separately by AzAPI
      ]);

      // Provider-internal computed fields that have no ARM equivalent
      const providerInternalPatterns = [
        /^primary_/, /^secondary_/, /_connection_string$/,
        /^resource_id$/, /^versionless_id$/, /^public_key_/,
      ];

      function isProviderInternal(field: string): boolean {
        return providerInternalPatterns.some(p => p.test(field));
      }

      // 4. Read the full schema to understand block nesting
      const schemaPath = join(azurerm_path, ".release", "provider-schema.json");
      const raw = await readFile(schemaPath, "utf-8");
      const fullSchema = JSON.parse(raw);
      const resourceSchema = fullSchema.providerSchema?.resources?.[resource_name]?.schema || {};

      // Known Terraform block → ARM path overrides (common patterns)
      // The mapping extraction gives us d.Set("blob_properties", flatten...(model.Properties.BlobServiceProperties))
      // We use expand/flatten function associations to infer block ARM paths.
      const blockArmPaths = new Map<string, string>();

      // Build block ARM paths from expand/flatten data
      if (!("error" in mappingResult) && "expandFlattenFunctions" in mappingResult) {
        for (const ef of (mappingResult as any).expandFlattenFunctions) {
          if (!ef.terraformField) continue;
          // Check if there's an existing mapping for this block
          const mapped = existingMappings.get(ef.terraformField);
          if (mapped) {
            blockArmPaths.set(ef.terraformField, mapped);
          }
        }
      }

      interface AutomapEntry {
        terraformPath: string;       // full dotted Terraform path
        armPath: string | null;      // inferred ARM path, null if unresolvable
        source: "mapping" | "mechanical" | "envelope" | "provider_internal";
        flags: string[];             // schema flags: computed, required, optional, sensitive, forceNew, default
        isBlock: boolean;            // true if this field is a nested block
      }

      const entries: AutomapEntry[] = [];

      function getFieldFlags(field: string, schemaData: any): string[] {
        const flags: string[] = [];
        const s = schemaData;
        if (!s) return flags;
        if (s.computed && !s.optional && !s.required) flags.push("computed");
        if (s.computed && s.optional) flags.push("default");
        if (s.required) flags.push("required");
        if (s.optional && !s.computed) flags.push("optional");
        if (s.sensitive) flags.push("sensitive");
        if (s.forceNew) flags.push("forceNew");
        return flags;
      }

      // 5. Walk the schema tree, producing ARM paths
      function walkSchema(
        fields: Record<string, any>,
        tfPrefix: string,
        armPrefix: string,
      ) {
        for (const [name, f] of Object.entries(fields)) {
          const tfPath = tfPrefix ? `${tfPrefix}.${name}` : name;
          const isBlock = !!f.elem?.schema;
          const flags = getFieldFlags(name, f);

          // Skip envelope fields at top level
          if (!tfPrefix && envelopeFields.has(name)) {
            entries.push({ terraformPath: tfPath, armPath: null, source: "envelope", flags, isBlock });
            continue;
          }

          // Skip provider-internal fields
          if (isProviderInternal(name)) {
            entries.push({ terraformPath: tfPath, armPath: null, source: "provider_internal", flags, isBlock });
            continue;
          }

          // Check existing mapping from d.Set/d.Get extraction
          const mappedArm = existingMappings.get(tfPath) || existingMappings.get(name);
          if (mappedArm) {
            entries.push({ terraformPath: tfPath, armPath: mappedArm, source: "mapping", flags, isBlock });
            // If it's a block, recurse into its children using the mapped ARM path
            if (isBlock && f.elem.schema) {
              walkSchema(f.elem.schema, tfPath, mappedArm);
            }
            continue;
          }

          // Check if we have a block ARM path override
          if (isBlock && blockArmPaths.has(name)) {
            const armPath = blockArmPaths.get(name)!;
            entries.push({ terraformPath: tfPath, armPath, source: "mapping", flags, isBlock });
            if (f.elem.schema) {
              walkSchema(f.elem.schema, tfPath, armPath);
            }
            continue;
          }

          // Mechanical conversion: snake_case → camelCase under the current ARM prefix
          const camelName = snakeToCamel(name);
          const armPath = armPrefix ? `${armPrefix}.${camelName}` : `properties.${camelName}`;
          entries.push({ terraformPath: tfPath, armPath, source: "mechanical", flags, isBlock });

          // Recurse into blocks
          if (isBlock && f.elem.schema) {
            walkSchema(f.elem.schema, tfPath, armPath);
          }
        }
      }

      // When block= is set, start from that block's sub-schema
      let walkRoot = resourceSchema;
      let walkTfPrefix = "";
      let walkArmPrefix = "";
      if (block) {
        const parts = block.split(".");
        let cursor = resourceSchema;
        let armParts: string[] = [];
        for (const part of parts) {
          const field = cursor[part];
          if (!field?.elem?.schema) {
            return { error: `Block '${block}' not found in ${resource_name}.` };
          }
          // Check if we have a mapped ARM path for this block level
          const mapped = existingMappings.get(part) || blockArmPaths.get(part);
          armParts.push(mapped || `properties.${snakeToCamel(part)}`);
          cursor = field.elem.schema;
        }
        walkRoot = cursor;
        walkTfPrefix = block;
        // Use the last mapped ARM path if available, otherwise construct from parts
        walkArmPrefix = armParts[armParts.length - 1];
      }

      walkSchema(walkRoot, walkTfPrefix, walkArmPrefix);

      // Partition results
      const mapped = entries.filter(e => e.armPath && e.source === "mapping");
      const automapped = entries.filter(e => e.armPath && e.source === "mechanical");
      const skipped = entries.filter(e => e.source === "envelope" || e.source === "provider_internal");
      const totalWithArm = mapped.length + automapped.length;

      return {
        resource: resource_name,
        ...(block ? { block } : {}),
        totalFields: entries.length,
        totalWithArmPath: totalWithArm,
        totalSkipped: skipped.length,
        mapped: mapped.map(e => ({ tf: e.terraformPath, arm: e.armPath, flags: e.flags, isBlock: e.isBlock })),
        automapped: automapped.map(e => ({ tf: e.terraformPath, arm: e.armPath, flags: e.flags, isBlock: e.isBlock })),
        skipped: skipped.map(e => ({ tf: e.terraformPath, reason: e.source, flags: e.flags })),
        // Provide a size hint for the agent to decide on decomposition strategy
        ...((!block && Object.keys(resourceSchema).length > 30) ? { recommendation: "Large resource — consider using block= to process nested blocks individually" } : {}),
      };
    }

    // --- Relational / cross-property constraints extraction ---
    // Scans schema fields for ConflictsWith / RequiredWith / ExactlyOneOf /
    // AtLeastOneOf attributes. Each occurrence records the subject Terraform
    // field (the field the attribute is declared on), the referenced Terraform
    // field paths, the kind, and the source file:line.
    async function extractRelational() {
      onUpdate?.({ content: [{ type: "text", text: "Scanning for relational / cross-property constraints..." }] });

      const KINDS = ["ConflictsWith", "RequiredWith", "ExactlyOneOf", "AtLeastOneOf"] as const;
      type Kind = (typeof KINDS)[number];

      interface RelationalOccurrence {
        kind: Kind;
        subject: string;      // enclosing schema field the attribute is declared on
        references: string[]; // referenced Terraform field paths
        relativePath: string;
        line: number;
      }

      // Collect the string literals of a `[]string{...}` attribute value that
      // begins on lines[startIdx]. Handles single-line and multi-line forms.
      function collectStringSlice(lines: string[], startIdx: number): string[] {
        let buf = lines[startIdx];
        const braceIdx = buf.search(/\[\]string\s*\{/);
        if (braceIdx >= 0) buf = buf.slice(braceIdx);
        if (!buf.includes("}")) {
          for (let j = startIdx + 1; j < Math.min(startIdx + 40, lines.length); j++) {
            buf += " " + lines[j];
            if (lines[j].includes("}")) break;
          }
        }
        return [...buf.matchAll(/"([^"]+)"/g)].map((x) => x[1]);
      }

      // Extract all relational occurrences from a file's lines.
      function occurrencesFromLines(lines: string[], relPath: string): RelationalOccurrence[] {
        const found: RelationalOccurrence[] = [];
        for (let i = 0; i < lines.length; i++) {
          for (const kind of KINDS) {
            if (new RegExp(`\\b${kind}:\\s*\\[\\]string\\s*\\{`).test(lines[i])) {
              found.push({
                kind,
                subject: extractFieldName(lines, i),
                references: collectStringSlice(lines, i),
                relativePath: relPath,
                line: i + 1,
              });
              break; // at most one kind per line
            }
          }
        }
        return found;
      }

      function countByKind(occ: RelationalOccurrence[]) {
        const byKind = { ConflictsWith: 0, RequiredWith: 0, ExactlyOneOf: 0, AtLeastOneOf: 0 };
        for (const o of occ) byKind[o.kind]++;
        return byKind;
      }

      // --- Per-resource mode ---
      if (resource_name) {
        const filePath = await findResourceFile(resource_name);
        if (!filePath) {
          return { error: `Resource file not found for ${resource_name}` };
        }
        const lines = await getFileLines(filePath);
        let occurrences = occurrencesFromLines(lines, relative(azurerm_path, filePath));

        // When block= is set, keep occurrences whose subject or references fall
        // inside that block sub-tree.
        if (block) {
          const blockLeaf = block.split(".").pop() ?? block;
          occurrences = occurrences.filter(
            (o) =>
              o.subject === blockLeaf ||
              o.references.some((r) => r.startsWith(block) || r.split(".").includes(blockLeaf)),
          );
        }

        return {
          resource: resource_name,
          ...(block ? { block } : {}),
          resourceFile: relative(azurerm_path, filePath),
          occurrences,
          totalOccurrences: occurrences.length,
          byKind: countByKind(occurrences),
        };
      }

      // --- Service-wide mode ---
      const grepLines = await grep(
        "ConflictsWith:\\|RequiredWith:\\|ExactlyOneOf:\\|AtLeastOneOf:",
        searchPath,
      );
      const files = new Set<string>();
      for (const raw of grepLines) {
        const parsed = parseGrepLine(raw);
        if (parsed) files.add(parsed.file);
      }

      const byService = new Map<string, { relativePath: string; occurrences: RelationalOccurrence[] }[]>();
      let totalOccurrences = 0;
      for (const filePath of files) {
        const lines = await getFileLines(filePath);
        const occ = occurrencesFromLines(lines, relative(azurerm_path, filePath));
        if (occ.length === 0) continue;
        totalOccurrences += occ.length;
        const svc = serviceFromPath(filePath);
        let arr = byService.get(svc);
        if (!arr) { arr = []; byService.set(svc, arr); }
        arr.push({ relativePath: relative(azurerm_path, filePath), occurrences: occ });
      }

      const services = [...byService.entries()]
        .sort((a, b) => a[0].localeCompare(b[0]))
        .map(([svc, filesArr]) => ({
          service: svc,
          files: filesArr,
          totalOccurrences: filesArr.reduce((s, f) => s + f.occurrences.length, 0),
        }));

      return { totalFiles: byService.size, totalOccurrences, services };
    }

    // --- Execute requested categories ---
    const details: Record<string, unknown> = {};
    const summaryParts: string[] = [];

    for (const cat of categories) {
      switch (cat) {
        case "forcenew": {
          const r = await extractFieldAnnotations(
            "ForceNew:.*true",
            "ForceNew",
          );
          details.forcenew = r;
          summaryParts.push(
            `forcenew: ${r.totalFiles} files, ${r.totalAnnotations} annotations`,
          );
          break;
        }
        case "sensitive": {
          const r = await extractFieldAnnotations(
            "Sensitive:.*true",
            "Sensitive",
          );
          details.sensitive = r;
          summaryParts.push(
            `sensitive: ${r.totalFiles} files, ${r.totalAnnotations} annotations`,
          );
          break;
        }
        case "timeouts": {
          const r = await extractTimeouts();
          details.timeouts = r;
          summaryParts.push(`timeouts: ${r.totalFiles} files`);
          break;
        }
        case "softdelete": {
          const r = await extractSoftDelete();
          details.softdelete = r;
          summaryParts.push(`softdelete: ${r.totalFiles} files`);
          break;
        }
        case "schema": {
          const r = await extractSchema();
          details.schema = r;
          if ("error" in r) {
            summaryParts.push(`schema: error - ${r.error}`);
          } else if ("resource" in r) {
            const verStr = r.apiVersions.length > 0 ? `, apiVersions: ${r.apiVersions.join(", ")}` : "";
            const blockStr = r.blocks?.length ? `, ${r.blocks.length} blocks` : "";
            summaryParts.push(
              `schema: ${r.resource}${r.block ? ` [${r.block}]` : ""} — ${r.totalFields} fields, ${r.computed.length} computed, ${r.defaults.length} defaults, ${r.forceNew.length} forceNew, ${r.sensitive.length} sensitive${blockStr}${verStr}`,
            );
          } else {
            summaryParts.push(`schema: ${r.totalResources} resources`);
          }
          break;
        }
        case "mapping": {
          const r = await extractMapping();
          details.mapping = r;
          if ("error" in r) {
            summaryParts.push(`mapping: error - ${r.error}`);
          } else if ("resource" in r) {
            summaryParts.push(
              `mapping: ${r.resource}${r.block ? ` [${r.block}]` : ""} — ${r.totalMapped} mapped, ${r.totalUnmapped} unmapped of ${r.schemaFieldCount} schema fields, ${r.expandFlattenFunctions.length} expand/flatten functions`,
            );
          } else {
            summaryParts.push(`mapping: ${r.totalFiles} files, ${r.totalMappings} field mappings`);
          }
          break;
        }
        case "validation": {
          const r = await extractValidation();
          details.validation = r;
          if ("error" in r) {
            summaryParts.push(`validation: error - ${r.error}`);
          } else if ("resource" in r) {
            summaryParts.push(`validation: ${r.resource}${r.block ? ` [${r.block}]` : ""} — ${r.totalRules} rules`);
          } else {
            summaryParts.push(`validation: ${r.totalFiles} services`);
          }
          break;
        }
        case "automap": {
          const r = await extractAutomap();
          details.automap = r;
          if ("error" in r) {
            summaryParts.push(`automap: error - ${r.error}`);
          } else {
            summaryParts.push(
              `automap: ${r.resource}${r.block ? ` [${r.block}]` : ""} — ${r.totalWithArmPath} ARM paths (${r.mapped.length} from d.Set/d.Get, ${r.automapped.length} mechanical), ${r.totalSkipped} skipped${r.recommendation ? ` ⚠ ${r.recommendation}` : ""}`,
            );
          }
          break;
        }
        case "relational": {
          const r = await extractRelational();
          details.relational = r;
          if ("error" in r) {
            summaryParts.push(`relational: error - ${r.error}`);
          } else if ("resource" in r) {
            const b = r.byKind;
            summaryParts.push(
              `relational: ${r.resource}${r.block ? ` [${r.block}]` : ""} — ${r.totalOccurrences} constraints (${b.ConflictsWith} conflictsWith, ${b.RequiredWith} requiredWith, ${b.ExactlyOneOf} exactlyOneOf, ${b.AtLeastOneOf} atLeastOneOf)`,
            );
          } else {
            summaryParts.push(`relational: ${r.totalFiles} files, ${r.totalOccurrences} constraints`);
          }
          break;
        }
      }
    }

    return { content: [{ type: "text", text: summaryParts.join("\n") }], details };
  },
});

export default factory;
