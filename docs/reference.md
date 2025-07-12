## Commands Reference

- `upduck install [server|tower]`:
 - ensures that wireguard is installed;
  - generates the private and public key and stores it under `/etc/upduck/wireguard-config.json`
 - for server:
    - ensures that k3s is installed;
  - for tower:
    - ensures that nginx is installed;
  - starts a systemctl service with a golang http server that will be use to interact between a server and a tower;
- `upduck connections`:
  - shows relevant information about remote servers/towers and also prints the public key's digest (used while connecting a server to the tower);

#### Server commands

- `upduck connect [tower-dns]`:
  - after a tower allows the current server's public key, this command is used to make a post request to the tower (`/api/servers/connect`), passing its Wireguard private key and receiving back the tower's public key and also its Wireguard public key. With the result, appends the data to (`/etc/upduck/connections.json`)

#### Tower commands

- `upduck allow [server-pub-key]`:
  - appends the public key into a list of known servers (`/etc/upduck/connections.json`). It is used to filter which servers can connect to this tower;
- `upduck dns forward [domain] [server] [server-local-address]:[PORT]`:
  - creates an nginx configuration to match the specific domain and redirect it to the server's private IP at a specific port (or 80).

## Endpoints

- `/api/servers/connect`: is exposed only in tower nodes. It receives the JSON payload:
 ```json
 {
   "public_key": "",
   "wg_public_key": "",
 }
 ```

 and responds:
 ```json
{
    "wg_public_key": "",
    "wg_network_block": "",
    "wg_address": "",
    "public_key": "",
}
 ```
