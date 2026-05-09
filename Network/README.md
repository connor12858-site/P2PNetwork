**GoP2P — Minimal P2P Networking Core**

This repository contains the minimal networking core used for the P2P project. It intentionally excludes application-layer code (UI, mobile app logic, example apps). For full documentation, guides, and CLI reference, see the main docs site: https://docs.connor12858.ca — this README is a concise, developer-focused companion.

**Project Goals**:
- Provide a small, well-scoped libp2p-based node implementation.
- Ship two CLI binaries: a `bootstrap` node (generates bootstrap addresses) and a `peer` node (interactive peer client).
- Keep the codebase lightweight and focused on networking primitives.

**Repository layout**
- `cmd/bootstrap/` — bootstrap node entrypoint that outputs persistent bootstrap addresses (writes `bs-nodes`).
- `cmd/peer/` — interactive peer CLI for connecting to peers, listing peers and discovery.
- `pkg/node/` — core networking primitives: host creation, DHT, discovery, connect/disconnect helpers.
- `replace/` — local module stubs used to satisfy transitive dependencies for local development.
- `build.ps1`, `Makefile` — convenience build scripts (Windows and Make targets).

**Requirements**
- Go 1.26.x (or compatible Go toolchain) installed and on PATH.
- Network access for libp2p bootstrap and discovery unless running strictly local tests.

**Quick start (build)**

Using PowerShell (Windows):

```powershell
# From the repository folder: GoP2P
powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1 -OutDir .\bin
```

Using Make (any platform with make + Go):

```bash
make all
# or individually
make peer
make bootstrap
```

Binaries will be written to `./bin/` by the PowerShell script and to the local folder by the Makefile targets. The repository `.gitignore` excludes build artifacts and `bs-nodes`.

**Running the bootstrap node**

The bootstrap node prints a bootstrap multiaddr and writes `bs-nodes` to disk. Example:

```bash
# Run the bootstrap node
./bin/bootstrap-node.exe

# or (if built to current dir)
./bootstrap-node
```

`bs-nodes` contains addresses for peers to connect to as a starting point. Keep this file secure if used for private networks.

**Running the peer CLI**

The `peer` binary is a simple interactive client with commands such as `connect`, `disconnect`, `peers`, `discovered`, and `exit`.

```bash
./bin/peer-node.exe
# then use commands inside the CLI prompt, for example:
# connect <multiaddr>
# peers
# discovered
# disconnect <peerID>
# exit
```

See `cmd/peer` for implementation details and available commands.

**Configuration & Notes**
- This project uses a local `replace` module under `replace/anet` to satisfy a transitive dependency required by the transport stack. This is intentional for local development and avoids depending on platform-internal packages.
- The code is designed for small-scale or private P2P experiments. If you plan to run a public network, review NAT traversal, security, and encryption settings carefully.

**Development workflow**
- Edit code in `pkg/node` for changes to the networking primitives.
- Run `go test ./...` from the `GoP2P` folder to run package tests (if any).
- Use the provided `build.ps1` or `Makefile` to produce local binaries for testing.

**Debugging & Troubleshooting**
- If `go build` complains about missing modules, run:

```bash
go mod tidy
```

- If you see errors referencing `github.com/wlynxg/anet`, ensure the `replace` path in `go.mod` points to `./replace/anet` and that the files under `replace/anet` exist.

**Where to get more information**
This README is intentionally focused on local development and the minimal networking core. For complete documentation, design rationale, tutorials, and extended examples, please visit the main docs site:

https://docs.connor12858.ca

Prefer the docs site for end-user guides, API docs, and the canonical source of truth. This repo is the lightweight code distribution and developer starter.

**Contribution & Contact**
- Issues and pull requests are welcome. For sensitive network or security discussions, open an issue or contact the maintainer via the project site.

**License**
- See the top-level repository `LICENSE` if present. If no license is included, treat the code as proprietary until a license is added.
