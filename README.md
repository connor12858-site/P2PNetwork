# P2PNetwork

A **decentralized, privacy-focused overlay network** written in Go using libp2p. Run a peer-to-peer network with bootstrap nodes, distributed peer discovery, and encrypted direct connections.

## 🚀 Quick Start

**Want to get started in 15 minutes?** Follow the [Getting Started Guide](../documentation/src/content/docs/getting-started/index.mdx) or visit [Getting Started](#getting-started-walkthrough) below.

## 📖 Documentation

Complete documentation is available in the [documentation/](../documentation/) folder and online:

- **[Getting Started](https://p2pnetwork.example.com/getting-started/)** — 15 min walkthrough
- **[Architecture](https://p2pnetwork.example.com/architecture/)** — System design and concepts
- **[How to Use](https://p2pnetwork.example.com/how-to-use/)** — Detailed usage guide
- **[Services](https://p2pnetwork.example.com/services/)** — Bootstrap, discovery, and connection services
- **[Network Protocol](https://p2pnetwork.example.com/network-protocol/)** — Technical protocol details
- **[Deployment](https://p2pnetwork.example.com/deployment/)** — Production setup and scaling
- **[FAQ](https://p2pnetwork.example.com/faq/)** — Frequently asked questions

## 🏗️ Project Structure

```
P2PNetwork/
├── README.md (this file)
├── network.md (protocol specifications - in progress)
├── GoP2P/
│   ├── go.mod (dependencies)
│   ├── Makefile (build targets)
│   ├── build.ps1 (PowerShell build script)
│   ├── bs-nodes (bootstrap multiaddr file, generated at runtime)
│   ├── bin/ (compiled binaries)
│   ├── cmd/
│   │   ├── bootstrap/ (Bootstrap node implementation)
│   │   │   └── main.go
│   │   └── peer/ (Peer node implementation)
│   │       └── main.go
│   ├── pkg/
│   │   └── node/ (Core node logic)
│   │       └── node.go
│   └── replace/ (Module replacements)
│       └── github.com/wlynxg/anet/ (Android network interface compatibility)
└── planning/
    └── chat.md (Design vision and architecture decisions)
```

## 🎯 What This Does

- **Bootstrap Node** — Initializes the network and provides peer discovery
- **Peer Node** — Joins the network, discovers other peers, and maintains connections
- **Peer Discovery** — Automatic peer finding via DHT (Distributed Hash Table)
- **Encrypted Connections** — All peer-to-peer traffic is encrypted by default
- **CLI Interface** — Interactive commands: `peers`, `discovered`, `connect`, `disconnect`, `exit`

## ⚡ Quick Start Walkthrough

### Prerequisites

- Go 1.26.2 or later
- Git
- 5+ GB free disk space

### Step 1: Clone & Build

```bash
git clone https://github.com/connor12858-site/P2PNetwork.git
cd P2PNetwork/GoP2P

# Build bootstrap node
go build -o bootstrap ./cmd/bootstrap

# Build peer node
go build -o peer ./cmd/peer
```

### Step 2: Start Bootstrap Node

**Terminal 1:**
```bash
./bootstrap
```

Copy the multiaddr output (looks like `/ip4/127.0.0.1/tcp/52837/p2p/12D3Koo...`).

### Step 3: Start Peer Nodes

**Terminal 2:**
```bash
./peer /ip4/127.0.0.1/tcp/52837/p2p/12D3Koo...
```

**Terminal 3:**
```bash
./peer -port 52839 /ip4/127.0.0.1/tcp/52837/p2p/12D3Koo...
```

### Step 4: Test the Network

In any terminal:
```bash
peers       # View connected peers
discovered  # View discovered peers
exit        # Shutdown gracefully
```

**Congratulations!** You now have a working 3-node P2P network. See [Getting Started Guide](https://p2pnetwork.example.com/getting-started/) for detailed steps and troubleshooting.

## 🔧 Build Options

### Build Both Binaries

```bash
go build -o bootstrap ./cmd/bootstrap
go build -o peer ./cmd/peer
```

### Optimized Release Build

```bash
go build -ldflags="-s -w" -o bootstrap ./cmd/bootstrap  # Smaller binary
```

### Cross-Compile (Windows → Linux)

```bash
GOOS=linux GOARCH=amd64 go build -o bootstrap-linux ./cmd/bootstrap
```

## 📊 Architecture Overview

```
Peer Node
├─ Peer ID (cryptographic key)
├─ libp2p Host
│  ├─ Encrypted connections (Noise Protocol)
│  ├─ NAT traversal
│  └─ Connection management
├─ Kademlia DHT
│  ├─ Peer discovery
│  ├─ Peer registration
│  └─ Lookup service
└─ CLI Interface
   ├─ Command: peers (list connected)
   ├─ Command: discovered (list DHT peers)
   ├─ Command: connect (manual connection)
   └─ Command: disconnect (disconnect peer)
```

## 🌐 Network Protocol

- **Transport:** TCP (QUIC planned)
- **Encryption:** Noise Protocol XX (Curve25519 + ChaCha20-Poly1305)
- **Discovery:** Kademlia DHT
- **Discovery Topic:** `fgov-network`
- **Default Port:** 52837 (bootstrap), auto-assigned (peers)

## 🔐 Security

- ✓ **Encrypted connections** — All peer-to-peer traffic is encrypted
- ✓ **Peer identity verification** — Cryptographic keys prove identity
- ✓ **Connection authentication** — Handshake prevents spoofing
- ⚠️ **Metadata privacy** — Connection metadata is visible (planned: multi-hop routing)
- ⚠️ **Sybil resistance** — Nothing stops creating many node IDs (planned: reputation)

See [Network Protocol](https://p2pnetwork.example.com/network-protocol/) for security details.

## 📦 Pre-Built Binaries

When releases are published, download platform-specific binaries:

- Windows: `bootstrap.exe`, `peer.exe`
- macOS (Intel): `bootstrap-darwin-amd64`, `peer-darwin-amd64`
- macOS (Apple Silicon): `bootstrap-darwin-arm64`, `peer-darwin-arm64`
- Linux: `bootstrap-linux-amd64`, `peer-linux-amd64`, `peer-linux-arm64`

[Download from Releases](https://github.com/connor12858-site/P2PNetwork/releases)

## 🚢 Deployment

See [Deployment Guide](https://p2pnetwork.example.com/deployment/) for:

- Running bootstrap nodes 24/7
- Monitoring and logging
- Multi-node setups
- Production scaling
- Systemd service setup

## 📈 Performance Characteristics

| Metric | Value | Notes |
|--------|-------|-------|
| Bootstrap startup | < 100ms | DHT join |
| Peer join time | 1-2s | Advertise + propagation |
| Discovery interval | ~30s | Periodic peer finding |
| Connection timeout | 30s | Per attempt |
| Latency (direct) | ~10-50ms | Same as TCP |
| Typical hops | 1-2 | Direct or via bootstrap |

## 🐛 Troubleshooting

### "Port already in use"
```bash
./peer -port 52840 <bootstrap-multiaddr>
```

### "Can't connect to bootstrap"
- Check bootstrap is running
- Verify multiaddr is correct
- Check firewall allows port 52837

### Build fails
```bash
go mod download
go clean -modcache
go build -o bootstrap ./cmd/bootstrap
```

See [How to Use - Troubleshooting](https://p2pnetwork.example.com/how-to-use/#troubleshooting) for more.

## 🗺️ Roadmap

- **v1.0** — Bootstrap and peer nodes, discovery (current)
- **v1.1** — 2-hop routing, connection persistence, metrics
- **v1.2** — `.fgov` domain resolution, identity system UI
- **v2.0** — Multi-hop routing (3+), hidden services, app ecosystem

## 📝 Planning & Vision

See [planning/chat.md](planning/chat.md) for the complete vision:

- Architecture decisions
- Design rationale
- Why certain technologies were chosen
- What we're building (and what we're NOT)
- Realistic expectations

## 🤝 Contributing

1. **Report bugs:** [GitHub Issues](https://github.com/connor12858-site/P2PNetwork/issues)
2. **Ask questions:** [GitHub Discussions](https://github.com/connor12858-site/P2PNetwork/discussions)
3. **Submit PRs:** Follow the standard fork & pull request workflow

## 🔒 Security

For security vulnerabilities, please report privately instead of creating public issues. See SECURITY.md (if present) or contact the maintainer.

## 📄 License

See LICENSE file in repository.

## 📚 Additional Resources

- **[Full Documentation](https://p2pnetwork.example.com/)** — Complete guides and API reference
- **[GitHub Repository](https://github.com/connor12858-site/P2PNetwork)** — Source code and issues
- **[libp2p Documentation](https://docs.libp2p.io/)** — Underlying networking library
- **[Go Documentation](https://golang.org/doc/)** — Go language reference

## 🎓 Learning Path

1. **[Getting Started](https://p2pnetwork.example.com/getting-started/)** — Build and run (15 min)
2. **[Architecture](https://p2pnetwork.example.com/architecture/)** — Understand the design
3. **[How to Use](https://p2pnetwork.example.com/how-to-use/)** — Advanced usage
4. **[Network Protocol](https://p2pnetwork.example.com/network-protocol/)** — Deep technical dive
5. **[Deployment](https://p2pnetwork.example.com/deployment/)** — Run in production

## 💡 Philosophy

P2PNetwork is built on these principles:

- **Technically solid** — Clean, maintainable code
- **Usable** — Performance matters as much as privacy
- **Honest** — Realistic about scope and limitations
- **Layered** — Each component is independent and replaceable
- **Practical** — Solves real problems without unnecessary complexity

This is not a revolution. It's infrastructure for building decentralized applications that actually work.

---

**Ready to get started?** Visit [Getting Started](https://p2pnetwork.example.com/getting-started/) →

