# Smyth

## Product Overview

Smyth is the operator-facing manifest authoring tool for Anvil.

It is a Go CLI that creates and updates YAML manifests for resources that Anvil later reconciles. Smyth is responsible for authoring experience, sensible defaults, and writing manifest files into a target manifest directory that defaults to the current working directory.

Smyth is not a reconciliation engine, background service, or provider runtime.

## Core Principles

- Generated manifests remain atomic: each emitted manifest describes one resource kind.
- Authoring UX can be interactive and opinionated, but the emitted manifest shape must stay explicit and reviewable.
- The default write location is the current working directory unless the operator provides a manifest directory explicitly.
- Output should be deterministic and unsurprising: gather input, build the typed manifest, validate it, and then write it.
- Schema ownership stays shared: common manifest structs, kind constants, and schema-oriented validation belong in `alloy`.
- Execution stays simple: prefer direct command handling and focused packages over framework-heavy abstractions.
- The product optimizes for clarity, safety, and debuggability over abstraction, extensibility, or future-proofing.

## V1 Scope

The v1 product direction is a manifest-authoring CLI that helps operators create validated manifests for Anvil-managed resources and save them into a manifest directory.

Initial expected support:

- authoring `GitHubRepository` manifests
- validating manifest shape through `alloy` before write
- writing generated YAML files into the current directory by default
- allowing an explicit manifest directory override when needed

## Explicit Non-Goals

Smyth v1 does not include:

- reconciliation logic
- provider-specific apply operations
- long-running controllers or background services
- state persistence beyond the manifest files it writes
- generic plugin systems
- cross-resource orchestration engines
- duplicated schema structs or validation rules copied from `alloy`

## Public Repository Boundary

This repository is intended to remain public.

- Public examples and documentation must use sanitized placeholder values such as `example-org`, `example-repo`, and `123456789012`.
- The repository must never include real organization names, repository names, account IDs, credentials, or operational values.
- Real environment-specific manifests belong in separate implementation repositories.

## Shared Code Boundary

Smyth should keep authoring-specific CLI behavior in this repository and depend on the separate `alloy` project for shared manifest schema code.

- Common manifest structs, schema versions, kind constants, and schema-oriented validation should be added to `alloy`, not redefined locally in `smyth`.
- When Smyth needs new shared types or schema changes, update `alloy` first, then consume the new version here through the Go module dependency.
- Keep `smyth` focused on command UX, defaults, manifest construction, file naming, and filesystem write behavior.
- Keep reconciliation planning, provider apply behavior, and manifest consumption logic out of this repository.

## Coding Patterns

- Follow the lightweight Go layout already used in `anvil`: `cmd/<app>` for the entrypoint and `internal/...` for application code.
- Prefer the standard library before adding CLI or framework dependencies.
- Keep command handling explicit and easy to read.
- Write focused table-free tests when a few direct cases are clearer than abstraction.
- Add ADRs under `docs/adr/` when a decision has meaningful alternatives or trade-offs.

## Working Style

- Keep durable project guidance in this file.
- Keep `README.md` public-facing and concise.
- Add focused documentation when it materially helps contributors understand the product or implementation.
- Prefer direct implementation work over process-heavy planning artifacts.
