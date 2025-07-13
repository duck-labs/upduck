<h1 align="center">
  ⬆️🦆 Upduck
  <br/>
  <span align="center">
      <a href="https://github.com/duck-labs/upduck/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="License"></a>
      <a href="https://upduck.io/docs"><img src="https://img.shields.io/badge/Docs-informational" alt="Docs"></a>
    </span>
</h1>

## Introduction

UpDuck is a Golang CLI that helps developers create on-premise datacenters for self-hosted applications. The key point is to use already existing tools and just orchestrate them. Just configuring tools seems to be the easiest part of the journey but it starts to hurt when you have to manage all these tools at the same time.

### Features

- [x] Private networking with WireGuard
- [ ] Reverse proxy tunneling with Nginx
- [ ] Container orchestration with K3s (for servers)

## Installation

### Build from source

```bash
git clone https://github.com/duck-labs/upduck-v2.git
cd upduck-v2
make build
```

### Install as system service

To install UpDuck as a server node:
```bash
sudo ./upduck install server
```

To install UpDuck as a tower node:
```bash
sudo ./upduck install tower
```

## Usage

### Tower Setup (Control Node)

1. **Install as tower**:
   ```bash
   sudo ./upduck install tower
   ```

2. **View your tower's public key**:
   ```bash
   ./upduck connections
   ```

3. **Allow a server to connect**:
   ```bash
   ./upduck allow <server-public-key-digest>
   ```

4. **Forward a domain to a server**:
   ```bash
   ./upduck dns forward example.com myserver 192.168.1.100:3000
   ```

### Server Setup (Worker Node)

1. **Install as server**:
   ```bash
   sudo ./upduck install server
   ```

2. **Get your server's public key**:
   ```bash
   ./upduck connections
   ```

3. **Connect to a tower** (after tower has allowed your key digest):
   ```bash
   ./upduck connect <tower-domain>
   ```

### Management Commands

- **View connections and keys**: `./upduck connections`

## API Endpoints

### Tower Endpoints

- `POST /api/servers/connect`: Server connection endpoint, exposed only by the tower and used by the server's `connect` command:
  ```json
  {
    "public_key": "server-public-key",
    "wg_public_key": "server-wireguard-public-key"
  }
  ```

- `GET /health`: Health check endpoint

## Configuration

UpDuck stores configuration in `/etc/upduck/`:

- `wireguard-config.json`: WireGuard keys, generated during the setup;
- `connections.json`: Wireguard network and peers list and, for the tower, a list allowed keys digest data.

For development/testing, you can override the config directory and start both the tower and the server on the same machine:
```bash
# starting the tower
UPDUCK_CONFIG_DIR="/etc/upduck-tower" sudo -E ./upduck server --type tower --port 8081

# starting the server
UPDUCK_CONFIG_DIR="/etc/upduck-server" sudo -E ./upduck server --type server --port 8082

# making the server reach the tower
UPDUCK_CONFIG_DIR="/etc/upduck-server" sudo -E ./upduck connect 127.0.0.1:8081
```

### Building

```bash
make build
```

## Architecture

UpDuck uses a tower-server architecture:

- **Tower**: Central control node that manages servers and provides reverse proxy. Must be used within a server that has a public IP available;
- **Server**: Worker nodes that run applications and connect to towers, mainly servers on the homelab.

The system uses WireGuard for secure networking and provides HTTP APIs for management.

A RSA key-pair is generated to ensure end-to-end encryption of the management API's.

## License

MIT License - see LICENSE file for details.