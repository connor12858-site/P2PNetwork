# P2PNetwork — Minimal libp2p-based Node

This repository contains a minimal libp2p-based node implementation and small example apps (CLI + GUI) used for local testing and experiments.

Quick highlights:
- `pkg/node/` — core libp2p node implementation (DHT, discovery, connection management).
- `cmd/bootstrap/` — bootstrap node that can output a bootstrap multiaddr and write `bs-nodes`.
- `cmd/peer/` — CLI peer client for connecting, listing, and interacting with peers.
- `cmd/gui/` — small Fyne-based desktop GUI to start/stop a node and view peers.

Prerequisites
- Go 1.20+ installed and on `PATH`.

Build
```bash
# Build everything using Make
make all

# Or build the GUI on Windows via the provided script
powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1 -v "1.1"
```

Run
- GUI (Windows): run the produced `./bin/gui-1.1.exe`.
- Bootstrap: run `./bin/bootstrap-node` to generate `bs-nodes`.
- Peer CLI: run `./bin/peer-node` and use commands like `connect`, `peers`, `discovered`, `disconnect`, `exit`.

Notes
- The `replace/` directory contains local module stubs used for development to avoid pulling platform-specific dependencies.
- This project is intended for experiments and small private networks — review networking/NAT/security settings before production use.

See the top-level README for more details.
