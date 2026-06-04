# Azwise Agent — Design Analysis

## 1. Task Definition

Build an agent that automatically extracts operational knowledge from `terraform-provider-azurerm` source code and produces structured output consumable by `terraform-provider-azapi`.

### What the agent must extract

| Category | Pattern to detect | Scale |
|---|---|---|
| ForceNew rules | `ForceNew: true` in schema + conditional logic | 4,954 annotations / 1,319 files |
| Soft-delete / purge | `PurgeSoftDelete`, `RecoverSoftDeleted` patterns | 36 files |
| Custom timeouts | `pluginsdk.DefaultTimeout` overrides | 3,022 annotations |
| Naming validation | `validate/` packages with regex/length rules | 101 directories |
| Sensitive properties | `Sensitive: true` in schema | 1,004 annotations / 376 files |
| CustomizeDiff | `CustomizeDiff` cross-field constraints | 175 files |
| DiffSuppressFunc | `DiffSuppressFunc` per-property | 221 files |
| Custom pollers | `custompollers/` directories | 18 directories |

### Required agent capabilities

1. **Go source code reading** — navigate 131 service packages, read resource files
2. **Pattern matching** — AST-aware or regex-based detection of Go struct literals, function calls, assignments
3. **Structured extraction** — from matched patterns, extract property paths, values, conditions into typed records
4. **Cross-file correlation** — e.g., linking a `ForceNew: true` to its containing resource type and Terraform field path, then mapping to an ARM JSON path
5. **Parallel processing** — 1,134 resource files is too much for a single serial pass
6. **Output generation** — produce Go source files or JSON knowledge files for the `internal/azure/azwise/` package
7. **Validation** — verify extracted knowledge compiles and makes sense (e.g., a ForceNew property path actually exists in the Azure API schema)

## 2. Platform Comparison

### pi (upstream)

| Attribute | Value |
|---|---|
| Package scope | `@earendil-works/pi-*` |
| Packages | 4: `pi-ai`, `pi-agent-core`, `pi-coding-agent`, `pi-tui` |
| Runtime | Bun (TypeScript) |
| Built-in tools | ~15 (read, write, edit, bash, search, find, etc.) |
| Subagent / task system | **None** — no `task/` directory, no parallel agent spawning |
| AST tools | **None** — no `ast_grep` or `ast_edit` |
| LSP integration | **None** |
| Eval (in-process Python/JS) | **None** |
| Custom agent definitions | Not supported (no agent discovery, no frontmatter format) |
| Extension system | ✅ `ExtensionAPI` with 30+ events, tool registration, commands |
| Custom tools | ✅ `CustomToolFactory` pattern |
| Skills | ✅ `SKILL.md` format |
| SDK | ✅ `createAgentSession()` with in-memory/file sessions |
| Native acceleration | **None** — pure TypeScript |
| Extension examples | 80+ examples |

### oh-my-pi (fork, what we're running on now)

| Attribute | Value |
|---|---|
| Package scope | `@oh-my-pi/pi-*` |
| Packages | 5: `pi-ai`, `pi-agent-core`, `pi-coding-agent`, `pi-tui`, `pi-natives` |
| Runtime | Bun (TypeScript) + Rust native bindings |
| Built-in tools | 32+ (everything pi has plus ast_grep, ast_edit, eval, debug, browser, lsp, task, irc) |
| Subagent / task system | ✅ Full: `task/executor.ts`, parallel spawning, agent discovery, IRC peer messaging, schema-validated outputs |
| AST tools | ✅ `ast_grep` (structural search), `ast_edit` (structural rewrite) — both via tree-sitter |
| LSP integration | ✅ Full LSP client (diagnostics, definition, references, rename, code actions) |
| Eval (in-process Python/JS) | ✅ IPython kernel + persistent JS VM |
| Custom agent definitions | ✅ Markdown + YAML frontmatter, discovered from `~/.omp/agent/agents/` and `.omp/agents/` |
| Extension system | ✅ Same API as pi, plus hooks subsystem |
| Custom tools | ✅ Same factory pattern, plus `pushPendingAction` for approval flows |
| Skills | ✅ Same format, plus multi-provider discovery (native, claude, codex, agents) |
| SDK | ✅ Same API, plus MCP integration, LSP warmup, subagent options |
| Native acceleration | ✅ Rust-backed: file search, pattern matching, bash/PTY, text operations |
| Extension examples | ~15 examples (fewer than pi, but richer built-in surface) |

### Key differences that matter for azwise

| Capability | pi | oh-my-pi | Impact |
|---|---|---|---|
| **Subagent parallelism** | ❌ | ✅ task tool with parallel executor | Critical — 131 service packages need parallel processing |
| **AST-aware Go analysis** | ❌ | ✅ ast_grep with tree-sitter | Critical — structural matching of Go schema definitions beats regex |
| **Custom agent types** | ❌ | ✅ `.omp/agents/*.md` discovery | High — define an `azwise` agent type with specialized prompt/tools/output schema |
| **Structured output schema** | ❌ | ✅ Frontmatter `output:` with JSON schema validation | High — agent output is validated against schema, not free text |
| **IRC inter-agent comms** | ❌ | ✅ peer-to-peer messaging between live agents | Medium — coordination between parallel extractors |
| **In-process eval** | ❌ | ✅ Python + JS kernels | Medium — post-processing, data transformation, validation scripts |
| **Native search speed** | ❌ | ✅ Rust-backed grep/find | Medium — scanning 1,319 files for patterns |

## 3. Recommendation: oh-my-pi

oh-my-pi is the clear choice. The three decisive factors:

1. **Subagent system** — pi has no way to spawn parallel agents. Processing 131 service packages serially is impractical. OMP's task executor runs agents in parallel with IRC coordination and schema-validated structured output.

2. **AST tools** — `ast_grep` with tree-sitter can structurally match Go patterns like `ForceNew: true` inside schema definitions, extracting the containing field name and resource type. Regex-only extraction would be fragile and miss structural context.

3. **Custom agent definitions** — OMP discovers agent `.md` files from `.omp/agents/`, letting us define an `azwise` agent with a specialized system prompt, restricted tool set, output schema, and thinking level. pi has no equivalent.

## 4. Architecture Options

### Option A: Custom Agent Definition (`.omp/agents/azwise.md`)

Define azwise as a discoverable agent type that the `task` tool can spawn.

```
.omp/agents/azwise.md          # Agent definition with frontmatter
tools/azwise/SKILL.md           # Skill with extraction instructions
tools/azwise/extract.ts         # Custom tool for structured extraction
```

**How it works:**
- User runs OMP, types "extract ForceNew knowledge from azurerm"
- OMP loads the `azwise` skill, which provides context about what to extract
- Agent uses `task` tool to spawn parallel `azwise` subagents, one per service package
- Each subagent uses `ast_grep` / `search` / `read` to find patterns, then returns structured JSON via `yield`
- Parent agent aggregates results and writes output files

**Pros:** Fully integrated into OMP workflow, interactive, can iterate on results.
**Cons:** Depends on LLM quality for extraction logic; each run costs tokens.

### Option B: SDK Script (`tools/azwise/extract.ts`)

A standalone Bun script using OMP's SDK to orchestrate extraction programmatically.

```ts
import { createAgentSession, SessionManager } from "@oh-my-pi/pi-coding-agent";

const { session } = await createAgentSession({
  sessionManager: SessionManager.inMemory(),
  toolNames: ["read", "search", "find", "ast_grep", "write"],
});

for (const service of services) {
  await session.prompt(`Extract ForceNew rules from ${service}...`);
}
```

**Pros:** Repeatable, scriptable, can run in CI.
**Cons:** Still LLM-dependent per extraction; harder to iterate interactively.

### Option C: Custom Tool + Agent Hybrid

A custom tool (`azwise_extract`) that does the mechanical pattern matching in TypeScript (no LLM needed), and an agent definition for the parts that need reasoning (conditional ForceNew logic, cross-field constraints).

```
.omp/agents/azwise.md           # Agent for reasoning-heavy extraction
tools/azwise/SKILL.md            # Extraction knowledge/instructions
.omp/tools/azwise_extract/       # Custom tool for mechanical extraction
  index.ts                       # Registers azwise_extract tool
```

**How it works:**
- `azwise_extract` tool scans Go files with regex/AST patterns, returns structured matches
- For simple cases (ForceNew: true, Sensitive: true, timeouts), the tool extracts directly — no LLM needed
- For complex cases (conditional ForceNew, CustomizeDiff logic), the agent reads the matched code and interprets it
- Agent writes final output files

**Pros:** Cost-efficient (LLM only for hard cases), fast (mechanical extraction is instant), deterministic for simple patterns.
**Cons:** More upfront engineering; two codepaths to maintain.

### Recommendation: Option C (Hybrid)

Most of the extraction is mechanical — 4,954 `ForceNew: true` annotations, 1,004 `Sensitive: true` annotations, and 3,022 timeout values are all simple pattern matches that don't need an LLM. A custom tool handles those instantly and deterministically. The LLM agent handles the ~175 CustomizeDiff files and conditional ForceNew logic that require semantic understanding.

## 5. Distribution & File Layout

### Distribution: Project-level `.omp/` in the repo

Everything is committed to `terraform-provider-azapi/`. OMP auto-discovers agents, tools, and skills from `.omp/` when anyone `cd`s into the repo. Zero setup beyond having OMP installed.

### File layout

```
terraform-provider-azapi/
├── .omp/
│   ├── agents/
│   │   └── azwise.md                    # Agent definition (frontmatter + system prompt)
│   ├── tools/
│   │   └── azwise_extract/
│   │       └── index.ts                 # Custom tool: mechanical pattern extraction
│   └── skills/
│       └── azwise/
│           └── SKILL.md                 # Extraction context, categories, output format
├── tools/
│   └── azwise/
│       ├── azurerm-knowledge-transfer-analysis.md   # Analysis report (exists)
│       └── agent-design-analysis.md                 # This document
└── internal/
    └── azure/
        └── azwise/                      # Go runtime package (generated output)
            ├── registry.go              # Types + registry
            ├── keyvault.go              # Per-service knowledge (generated)
            ├── storage.go
            └── ...
```

### What teammates get automatically

| Component | Discovery path | What it does |
|---|---|---|
| `azwise` agent | `.omp/agents/azwise.md` | Available via `task` tool — specialized for semantic extraction |
| `azwise_extract` tool | `.omp/tools/azwise_extract/` | LLM-callable tool for mechanical Go pattern scanning |
| `azwise` skill | `.omp/skills/azwise/SKILL.md` | Context loaded into system prompt, readable via `skill://azwise` |

## 6. Design Decisions

### Output format: Go source files (registry pattern)

One `.go` file per service package with `init()` registrations. This matches azapi's existing patterns (`customization/`, `skipApiVersions`), supports conditional ForceNew via Go functions, has zero runtime deserialization cost, and keeps knowledge type-checked at compile time.

### Extraction mode: Full every run

No incremental delta tracking. Regenerate all output from scratch each run. The mechanical extraction is fast (~seconds for pattern matching across all files), and full extraction eliminates drift risk. No state files to manage.

### ARM path mapping: Auto-map using convention + AzureRM SDK calls

The agent will trace AzureRM's Create/Update functions to map Terraform `snake_case` field names to ARM `camelCase` JSON paths. AzureRM's typed SDK clients reveal the ARM structure — the Create function builds the SDK request struct, mapping each Terraform field to a typed Go struct field that corresponds to the ARM API property path. This is high effort but gives AzAPI the exact ARM paths it needs, which is the whole point of the knowledge transfer.

### Bicep validation: Deferred to a separate step

Extraction and validation are decoupled. The agent extracts knowledge and writes Go files. Validation against AzAPI's 31,175-type Bicep schema index happens later when integrating into azapi, not during extraction. This keeps the agent focused and avoids a dependency on path mapping being perfect before extraction can complete.