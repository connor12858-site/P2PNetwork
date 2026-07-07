# P2PNetwork

A lightweight peer-to-peer networking project built with **Go + libp2p**, with:
- a **bootstrap service** for peer registration,
- a **CLI peer node** for terminal-based networking,
- and a **GUI peer app** using Fyne.

The repository also includes a small **FastAPI bootstrap registry** used by peers to discover initial bootstrap addresses.

## What this project does

This project demonstrates how to:
- create libp2p nodes,
- connect peers using multiaddresses,
- use a Kademlia DHT for routing/discovery,
- advertise and discover peers,
- and exchange friendly peer display names over a custom protocol.

## Repository layout

```text
P2PNetwork/
├── cmd/
│   ├── bootstrap/      # Bootstrap node executable entrypoint
│   ├── peer/           # CLI peer executable entrypoint
│   └── gui/            # Fyne GUI executable entrypoint
├── pkg/
│   ├── node/           # Core libp2p node, discovery, connection logic
│   ├── gui/            # Thin GUI helper package
│   └── util/           # Config loader and utility helpers
├── server/
│   ├── main.py         # FastAPI bootstrap registry API
│   ├── Model.py        # Pydantic model(s)
│   └── data/           # Bootstrap registry data + server config
├── bin/
│   ├── config.yaml     # Example peer config
│   └── boot-config.yaml# Example bootstrap-node config
└── build.ps1           # PowerShell build script for cmd binaries
```

## Core components

### 1) `cmd/bootstrap`
Runs a libp2p node intended to act as a stable bootstrap peer.

It:
- loads `boot-config.yaml`,
- starts a node,
- registers itself in the FastAPI server via `POST /register`,
- and unregisters on shutdown via `DELETE /register/{peer_id}`.

### 2) `cmd/peer`
Runs a CLI node that:
- loads `config.yaml`,
- fetches bootstrap peers from the server root endpoint (`GET /`),
- starts discovery routines,
- and supports terminal commands:
  - `connect <multiaddr>`
  - `peers`
  - `discovered`
  - `exit`

### 3) `cmd/gui`
Runs a desktop app (Fyne) to start/stop a node and view connected peers.

### 4) `server/` (FastAPI bootstrap registry)
Provides a minimal registry to share bootstrap nodes:
- `GET /` – list registered nodes
- `POST /register` – register/update bootstrap node (requires query auth)
- `DELETE /register/{peer_id}` – unregister bootstrap node (requires query auth)
- `GET /health` – health/status
- `GET /help` – endpoint list

## Prerequisites

### Go side
- Go toolchain compatible with `go.mod` (currently `go 1.26.2`)
- Network access between peers

For GUI builds on Linux, system OpenGL/X11 development libraries are required (Fyne/GLFW dependency).

### Python side (bootstrap server)
- Python 3.10+
- `fastapi`, `pydantic`, `pyyaml`, `uvicorn`

## Configuration

### Peer config (`config.yaml`)
Expected fields:

```yaml
port: 52837
server: http://127.0.0.1:8000
password: abc123
topic: fgov
name: Desk
logging: false
```

### Bootstrap config (`boot-config.yaml`)
Expected fields:

```yaml
port: 0
server: http://127.0.0.1:8000
password: abc123
topic: fgov
name: Test-bootstrap
logging: false
```

### Server config (`server/data/config.yaml`)

```yaml
port: 8000
host: 0.0.0.0
password: abc123
bs-path: ./data/bootstraps
```

## Quick start

### 1) Start the bootstrap registry API

From `/home/runner/work/P2PNetwork/P2PNetwork/server`:

```bash
python -m venv .venv
source .venv/bin/activate
pip install fastapi pydantic pyyaml uvicorn
uvicorn main:app --host 0.0.0.0 --port 8000
```

### 2) Start a bootstrap node

From `/home/runner/work/P2PNetwork/P2PNetwork`:

```bash
cp ./bin/boot-config.yaml ./boot-config.yaml
go run ./cmd/bootstrap
```

### 3) Start one or more peer nodes

Open a new terminal in `/home/runner/work/P2PNetwork/P2PNetwork`:

```bash
cp ./bin/config.yaml ./config.yaml
go run ./cmd/peer
```

Repeat with different `name`/`port` values to simulate multiple peers.

## Build

PowerShell script (Windows-focused):

```powershell
./build.ps1 -OutDir ./bin -v 1.0
```

This compiles binaries from all `cmd/*` entrypoints.

## Networking notes

- Node listening address: `/ip4/0.0.0.0/tcp/<port>`
- Full peer address format: `<multiaddr>/p2p/<peer-id>`
- Bootstrap server stores node records in newline-delimited JSON (`server/data/bootstraps`)
- Peer names are exchanged over custom libp2p protocol ID `/fgov`

## Known development caveats

- GUI builds may fail on minimal Linux environments missing OpenGL/X11 headers.
- `cmd/peer` currently advertises using configured `topic` but discovery lookup is fixed to `fgov-network` in `pkg/node/discovery.go`.

## Troubleshooting

### Peer cannot fetch bootstrap list
- Ensure API is running at `config.yaml -> server`.
- Check `GET /health`.
- Ensure password in peer/bootstrap config matches `server/data/config.yaml`.

### Peer connection fails
- Verify reachable IP/port in reported multiaddress.
- Confirm firewall/NAT allows traffic.
- Confirm target process is still running.

### GUI fails to compile on Linux
- Install OpenGL + X11 development packages required by GLFW/Fyne.

## Future improvements

- Add authentication beyond shared query-string password.
- Add structured logging and metrics.
- Add automated tests for node/discovery behavior.
- Unify discovery topic usage between advertise and lookup.