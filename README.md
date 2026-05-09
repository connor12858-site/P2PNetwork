# P2PNetwork

Minimal libp2p-based node implementation with example CLI and GUI apps.

Build and run (quick):

```bash
# Build bootstrap and peer binaries
go build -o bin/bootstrap ./Network/cmd/bootstrap
go build -o bin/peer ./Network/cmd/peer

# Build GUI (on Windows with helper script)
powershell -NoProfile -ExecutionPolicy Bypass -File Network/build.ps1 -v "1.1"
```

Key folders:
- `Network/pkg/node` — core node implementation (DHT, discovery, connection management)
- `Network/cmd/bootstrap` — bootstrap node
- `Network/cmd/peer` — peer CLI
- `Network/cmd/gui` — small Fyne GUI

Configuration notes:
- `Network/bs-nodes` is produced by the bootstrap node and used by peers to connect.
- Use `go mod tidy` if modules are missing.
# P2PNetwork

P2PNetwork is a Go/libp2p project in the `Network/` folder of this repository. It currently ships a bootstrap node, a peer node, and a Fyne desktop GUI.

## Quick start

```bash
cd Network
go build -o bootstrap ./cmd/bootstrap
go build -o peer ./cmd/peer
go build -o gui ./cmd/gui
```

Run `bootstrap` first, then `peer`, and optionally `gui`.

The bootstrap node writes its address to `bs-nodes` in the working directory. Peer nodes read that file automatically when it is present.

## Documentation

The documentation site lives in `documentation/`.

- [Getting Started](documentation/src/content/docs/getting-started/index.mdx)
- [Architecture](documentation/src/content/docs/architecture/index.mdx)
- [How to Use](documentation/src/content/docs/how-to-use/index.mdx)
- [GUI](documentation/src/content/docs/gui/index.mdx)
- [Services](documentation/src/content/docs/services/index.mdx)
- [Network Protocol](documentation/src/content/docs/network-protocol/index.mdx)
- [Downloads](documentation/src/content/docs/downloads/index.mdx)
- [Release Notes](documentation/src/content/docs/release-notes/index.mdx)
- [Deployment](documentation/src/content/docs/deployment/index.mdx)
- [FAQ](documentation/src/content/docs/faq/index.mdx)

## Current structure

```text
P2PNetwork/
├── Network/
│   ├── cmd/
│   │   ├── bootstrap/
│   │   ├── peer/
│   │   └── gui/
│   ├── pkg/node/
│   ├── Makefile
│   └── build.ps1
├── Releases/
├── Plans/
├── network.md
└── documentation/
```

## Current releases

The `Releases/` folder currently contains these Windows binaries:

- `bootstrap-v1.0.exe`
- `peer-node-v1.0.exe`
- `bootstrap-v1.1.exe`
- `peer-v1.1.exe`
- `gui-v1.1.exe`

## What the project does today

- starts a bootstrap node on port `52837`
- connects peer nodes through libp2p
- discovers peers on the `fgov-network` topic
- exchanges peer names over the `/fgov` stream protocol
- shows the same node state in a terminal or GUI

## Planning notes

The broader long-term vision lives in `Plans/chat.md`. The repository also has an empty `network.md` placeholder for future protocol notes.

## Contributing

Use GitHub issues for normal bugs and feature work. Report security issues privately.

