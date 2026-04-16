# Smyth

Smyth is a user-facing manifest authoring CLI for Anvil.

It is intended to generate YAML manifests for supported resource kinds, validate them with shared schema code from `alloy`, and write them into a manifest directory that defaults to the current working directory.

This repository is the public product repository for Smyth. It is intended to show the tool direction, implementation patterns, and documentation without exposing real production data.

## Status

This is an initial working scaffold.

The repository currently includes the first Go CLI entrypoint under `cmd/smyth` and `internal/cli`, plus repository guidance that defines Smyth's role as the authoring counterpart to `anvil`.

## Core Principles

- Manifests stay atomic: each generated manifest describes one resource kind.
- Authoring UX can be more operator-friendly than the reconcile path, but emitted manifests must stay explicit and reviewable.
- Shared manifest structs and validation belong in `alloy`, not in duplicated local schema code.
- Filesystem behavior should stay unsurprising: write to the current directory by default and allow explicit override when needed.
- Distribution stays simple: Smyth is a Go CLI intended to ship as a versioned binary.

## V1 Scope

The v1 direction is a manifest-authoring CLI with shared schema validation and straightforward filesystem output.

Initial expected support:

- `GitHubRepository` manifest authoring
- validation through `alloy`
- writing generated YAML manifests into the current directory by default
- explicit manifest-directory override support

## Non-Goals

Smyth v1 does not include:

- reconciliation or apply behavior
- background services
- generic plugin systems
- cross-resource orchestration engines
- copied schema definitions that should live in `alloy`

## Repository Boundary

This repository remains public by design.

- Public documentation and examples must use sanitized placeholder values.
- Public content must never include real credentials or operational resource identifiers.
- Real implementation-repository manifests belong elsewhere.

## Local Development

Run the CLI help locally with:

```bash
go run ./cmd/smyth --help
```

Run the current test suite with:

```bash
go test ./...
```

Build a local binary with:

```bash
go build -o bin/smyth ./cmd/smyth
./bin/smyth --help
```

Build the Docker image locally with:

```bash
docker build -t smyth:local .
docker run --rm smyth:local
```

## Architecture Decisions

Strategic and architectural decisions for Smyth should be tracked as ADRs under [docs/adr](docs/adr/README.md).

## Relationship to Anvil and Alloy

- `smyth` owns manifest authoring UX, defaults, and filesystem output.
- `anvil` owns manifest loading and reconciliation behavior.
- `alloy` owns shared manifest schema types, version constants, and schema-oriented validation used by both tools.

## Distribution

Smyth publishes:

- GitHub Release archives for direct CLI consumption
- Docker images to GitHub Container Registry

The GitHub Actions workflow publishes the container image to `ghcr.io` and links it to this repository so the package can be managed from the repository's package settings.

Example published image reference:

```text
ghcr.io/example-org/smyth:latest
```

Replace `example-org` with the repository owner when pulling from GitHub Container Registry.

Container images are intended to remain private. Pulling from `ghcr.io` requires authenticating with Docker using a GitHub personal access token (classic) that has at least `read:packages`.

Example authentication flow:

```bash
export CR_PAT=YOUR_GITHUB_PAT
echo "$CR_PAT" | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin
docker pull ghcr.io/example-org/smyth:latest
```

If the package is owned by an organization that uses SSO, authorize the token for that organization before pulling images locally.

## AI-Assisted Development

AI agents may be used in this repository for coding assistance, drafting, and documentation generation.

They are used to accelerate implementation and communication, not as a substitute for engineering judgment. Code and documentation kept in this repository are expected to be reviewed and understood by the repository author.
