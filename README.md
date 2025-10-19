# BitTorrent Client

A BitTorrent client implementation in Go following the BitTorrent Protocol Specification.

## Features

- Bencoding support for torrent file parsing
- HTTP tracker communication
- Peer-to-peer protocol implementation
- Concurrent piece downloading
- Resume capability
- CLI interface

## Usage

```bash
# Quick start with Makefile
make build                           # Build the client
make run ARGS="file.torrent"         # Build and run with a torrent file
make dev                             # Development cycle (fmt + test + build)

# Manual build
go build -o torrent-client ./cmd

# Download a torrent file
./torrent-client <torrent-file> [output-path]
./torrent-client --version          # Show version information

# Examples:
./torrent-client ubuntu.torrent
./torrent-client ubuntu.torrent /downloads/ubuntu.iso
./torrent-client http://example.com/file.torrent

# Or run directly with Go
go run cmd/main.go <torrent-file> [output-path]
```

## Makefile Targets

```bash
# Build and Development
make build         # Build the binary
make build-all     # Build for multiple platforms
make dev          # Quick development cycle (fmt + test + build)
make install      # Install to $GOPATH/bin

# Testing and Quality
make test         # Run all tests
make test-coverage # Generate coverage report
make fmt          # Format code
make lint         # Lint code
make check        # Run all quality checks

# Utility
make clean        # Clean build artifacts
make help         # Show all available targets
```

## Project Structure

```
torrent-client/
├── bencode/       # Bencoding implementation
├── torrent/       # Torrent file parsing and metadata
├── peer/          # Peer protocol and connection management
├── client/        # Main client logic and download coordination
├── cmd/           # CLI application
└── README.md
```

## Testing

```bash
go test ./...
```

## Implementation Steps

This client is built following a systematic approach:

1. **Bencoding** - Encode/decode BitTorrent's serialization format
2. **Torrent Parsing** - Extract metadata from .torrent files
3. **Tracker Communication** - Get peer lists from HTTP trackers
4. **Peer Protocol** - Handle BitTorrent peer messages
5. **Connection Management** - Manage multiple peer connections
6. **Piece Tracking** - Coordinate piece downloads across peers
7. **Data Transfer** - Download and verify pieces
8. **File Assembly** - Write completed pieces to disk