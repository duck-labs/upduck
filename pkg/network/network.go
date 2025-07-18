package network

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/duck-labs/upduck/pkg/config"
	"github.com/duck-labs/upduck/types"
)

const wgConfigTowerTemplate = `[Interface]
PrivateKey = {{.PrivateKey}}
ListenPort = 51820
Address = {{.Address}}

PostUp = iptables -A FORWARD -i "{{.Name}}" -s {{.Address}} -d {{.Address}} -j ACCEPT
PreDown = iptables -D FORWARD -i "{{.Name}}" -s {{.Address}} -d {{.Address}} -j ACCEPT

PostUp = iptables -A FORWARD -i "{{.Name}}" -s {{.Address}} -j DROP
PreDown = iptables -D FORWARD -i "{{.Name}}" -s {{.Address}} -j DROP

{{range .Peers}}
[Peer]
PublicKey = {{.PublicKey}}
AllowedIPs = {{.Address}}

{{end}}`

const wgConfigServerTemplate = `[Interface]
PrivateKey = {{.PrivateKey}}
Address = {{.Address}}

{{range .Peers}}
[Peer]
PublicKey = {{.PublicKey}}
AllowedIPs = {{.Address}}
Endpoint = {{.Endpoint}}:51820
PersistentKeepalive = 25
{{end}}`

func GenerateWireguardKeys() (*types.WireguardConfig, error) {
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.PublicKey().String()

	return &types.WireguardConfig{
		PrivateKey: privateKey.String(),
		PublicKey:  publicKey,
	}, nil
}

func GetNextAvailableNetworkAddress(connectionsConfig *types.ConnectionsConfig, networkBlock *net.IPNet) (*net.IPNet, error) {
	used := make(map[uint32]bool)

	if len(connectionsConfig.Networks) > 0 {
		for _, peer := range connectionsConfig.Networks[0].Peers {
			ipInt := binary.BigEndian.Uint32([]byte(peer.Address))
			used[ipInt] = true
		}
	}

	netIP := networkBlock.IP.To4()
	ipInt := binary.BigEndian.Uint32(netIP)
	// first address (+0) is the "Network Address"
	// the last address (+255) is the "Broadcast Address"
	// the usable ranges between Net+1 and Net+254 (< net + 255)
	for nextIntIP := ipInt + 1; nextIntIP < ipInt+255; nextIntIP++ {
		if !used[nextIntIP] {
			nextNetIP := make(net.IP, 4)
			binary.BigEndian.PutUint32(nextNetIP, nextIntIP)
			return &net.IPNet{IP: nextNetIP, Mask: net.CIDRMask(32, 32)}, nil
		}
	}

	return nil, fmt.Errorf("no allocatable address")
}

func GetNextAvailableNetworkBlock(connectionsConfig *types.ConnectionsConfig) (*net.IPNet, error) {
	usedSubnets := make(map[int]bool)

	for _, network := range connectionsConfig.Networks {
		_, netBlock, err := net.ParseCIDR(network.Address)
		if err != nil {
			continue
		}

		if len(netBlock.IP) >= 3 && netBlock.IP[0] == 10 && netBlock.IP[1] == 5 {
			usedSubnets[int(netBlock.IP[2])] = true
		}
	}

	for subnet := 0; subnet < 256; subnet++ {
		if !usedSubnets[subnet] {
			return &net.IPNet{
				IP:   net.IP{10, 5, byte(subnet), 0},
				Mask: net.CIDRMask(24, 32),
			}, nil
		}
	}

	return nil, fmt.Errorf("no available network blocks in 10.5.x.0/24 range")
}

func WriteWireguardInterfaces(serverType string) error {
	connectionsConfig, err := config.LoadConnectionsConfig()
	if err != nil {
		return fmt.Errorf("error loading connections: %v", err)
	}

	wgConfig, err := config.LoadWireguardConfig()
	if err != nil {
		return fmt.Errorf("error loading wireguard config: %v", err)
	}

	err = config.EnsureWireguardDir()
	if err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}

	var wgTemplate string

	if serverType == "tower" {
		wgTemplate = wgConfigTowerTemplate
	}

	if serverType == "server" {
		wgTemplate = wgConfigServerTemplate
	}

	for nindex, network := range connectionsConfig.Networks {
		netName := fmt.Sprintf("udck-%c%d", serverType[0], nindex)
		configPath := fmt.Sprintf("%s/%s.conf", config.WireguardConfigDir, netName)

		tmpl, err := template.New(netName).Parse(wgTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse template: %v", err)
		}

		file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("failed to open wg file: %v", err)
		}
		defer file.Close()

		var peers []map[string]interface{}

		for _, np := range network.Peers {
			peer := map[string]interface{}{
				"PublicKey": np.PublicKey,
				"Address":   np.Address,
			}

			if serverType == "server" {
				peer["Endpoint"] = np.Endpoint
			}

			peers = append(peers, peer)
		}

		wgInterfaceConfig := map[string]interface{}{
			"Name":       netName,
			"PrivateKey": wgConfig.PrivateKey,
			"Address":    network.Address,
			"Peers":      peers,
		}

		err = tmpl.Execute(file, wgInterfaceConfig)
		if err != nil {
			return fmt.Errorf("failed to execute template: %v", err)
		}

		err = startWireGuardInterface(netName, configPath)
		if err != nil {
			return fmt.Errorf("failed to start wg interface: %v", err)
		}
	}

	return nil
}

func startWireGuardInterface(netName string, configPath string) error {
	wgClient, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to start wgClient: %v", err)
	}

	_, err = wgClient.Device(netName)
	if err == nil {
		downCmd := exec.Command("wg-quick", "down", configPath)
		output, err := downCmd.CombinedOutput()
		if err != nil {
			if !strings.Contains(string(output), "does not exist") {
				return fmt.Errorf("failed to stop wg interface: %v | %s", err, string(output))
			}
		}

		time.Sleep(1 * time.Second)
	}

	upCmd := exec.Command("wg-quick", "up", configPath)
	output, err := upCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start wg interface: %v | %s", err, string(output))
	}

	_, err = wgClient.Device(netName)
	if err != nil {
		return fmt.Errorf("failed to get wg interface after starting: %v", err)
	}

	return nil
}

func GenerateTimeOrderedID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func ResolveServerToIP(connectionsConfig *types.ConnectionsConfig, server string) (string, error) {
	if net.ParseIP(server) != nil {
		return server, nil
	}

	for _, network := range connectionsConfig.Networks {
		for _, peer := range network.Peers {
			if peer.ID == server {
				ip, _, err := net.ParseCIDR(peer.Address)

				if err != nil {
					return "", fmt.Errorf("failed to parse address")
				}

				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("server '%s' not found in connections", server)
}
