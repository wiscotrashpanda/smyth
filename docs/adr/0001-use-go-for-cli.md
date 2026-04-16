# 0001: Use Go for the Smyth CLI

- Status: Accepted
- Date: 2026-04-16

## Context

Smyth is intended to be a user-facing CLI for creating and updating manifest files that Anvil later reconciles.

The most plausible implementation language options for Smyth were Go, Python, and TypeScript.

Python would support very fast iteration and would be a comfortable choice for a prompt-heavy or filesystem-oriented CLI. TypeScript would also be a reasonable option, especially for interactive command-line UX and rich package ecosystems around prompts and developer tooling.

At the same time, Smyth is meant to live alongside Anvil as part of the same tool family. It is expected to consume shared manifest types and schema validation from `alloy`, ship as a versioned CLI binary, and remain easy to install and run in developer workstations or automation contexts without requiring a separately managed runtime environment.

That makes the language decision more than a matter of developer preference. It affects packaging, distribution, dependency sharing with `alloy`, and the consistency of the broader Anvil/Smyth toolchain.

## Decision

Smyth will be implemented in Go.

## Rationale

- Go produces a single compiled binary, which is a good fit for a CLI that should be easy to install and run.
- Using Go keeps Smyth aligned with Anvil and Alloy, reducing friction around shared types, module boundaries, and contributor context switching.
- A compiled distribution model simplifies release, version pinning, and cross-platform packaging compared with a runtime-dependent Python or Node.js toolchain.
- Strong typing is valuable for manifest authoring, where the CLI is constructing structured documents that should match shared schema definitions precisely.
- Go is well suited to straightforward filesystem, YAML, and command-dispatch code without requiring a large framework.
- Choosing the same implementation language across the related repositories keeps the tool family conceptually tighter and avoids introducing unnecessary polyglot maintenance overhead early.

## Consequences

### Positive

- Smyth can share idioms, module workflows, and contributor expectations with Anvil and Alloy.
- The project can distribute a simple binary artifact for local use and automation.
- Shared manifest construction and validation code can integrate naturally with the Go types provided by Alloy.
- The CLI should remain predictable as it grows beyond basic scaffold commands into manifest generation and writing behavior.

### Negative

- Python or TypeScript might have allowed faster early experimentation for highly interactive CLI flows.
- Some CLI UX libraries and ecosystem conveniences are richer or more familiar in Python and TypeScript than in Go.
- The project takes on the cost of building operator-facing command UX in a language that is not always the fastest for scripting-style iteration.

## Alternatives Considered

### Python

Python would have been a strong option for rapid prototyping and interactive command workflows. It was not chosen because the overhead of runtime management and the weaker fit with the shared Go-based schema/tooling ecosystem mattered more than short-term implementation speed.

### TypeScript

TypeScript would have been a defensible choice for a polished interactive CLI, especially with the Node.js package ecosystem. It was not chosen because introducing a separate runtime and language family for Smyth would make the Anvil, Smyth, and Alloy toolchain less consistent and increase long-term maintenance overhead.
