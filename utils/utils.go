package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/duck-labs/upduck-v2/types"
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

var (
	ConfigDir             = getConfigDir()
	WireguardConfigDir    = filepath.Join(ConfigDir, "wg-config")
	WireguardConfigFile   = filepath.Join(ConfigDir, "wireguard-config.json")
	ConnectionsConfigFile = filepath.Join(ConfigDir, "connections.json")
	RSAPublicKey          = filepath.Join(ConfigDir, "public-key.pem")
	RSAPrivateKey         = filepath.Join(ConfigDir, "private-key.pem")
)

func getConfigDir() string {
	dir := os.Getenv("UPDUCK_CONFIG_DIR")

	if dir != "" {
		return dir
	}

	return "/etc/upduck"
}

func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir, 0755)
}

func EnsureWireguardDir() error {
	return os.MkdirAll(WireguardConfigDir, 0755)
}

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

func SaveWireguardConfig(config *types.WireguardConfig) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(WireguardConfigFile, data, 0600)
}

func LoadWireguardConfig() (*types.WireguardConfig, error) {
	data, err := os.ReadFile(WireguardConfigFile)
	if err != nil {
		return nil, err
	}

	var config types.WireguardConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func GenerateRSAKeys() (*types.RSAKeysConfig, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA private key: %v", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKey := &privateKey.PublicKey
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	if err := EnsureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to ensure config directory: %v", err)
	}

	if err := os.WriteFile(RSAPrivateKey, privateKeyPEM, 0600); err != nil {
		return nil, fmt.Errorf("failed to write private key file: %v", err)
	}

	if err := os.WriteFile(RSAPublicKey, publicKeyPEM, 0644); err != nil {
		return nil, fmt.Errorf("failed to write public key file: %v", err)
	}

	return &types.RSAKeysConfig{
		PrivateKey: string(privateKeyPEM),
		PublicKey:  string(publicKeyPEM),
	}, nil
}

func LoadRSAKeys() (*types.RSAKeysConfig, error) {
	privData, err := os.ReadFile(RSAPrivateKey)
	if err != nil {
		return nil, err
	}

	pubData, err := os.ReadFile(RSAPublicKey)
	if err != nil {
		return nil, err
	}

	return &types.RSAKeysConfig{PrivateKey: string(privData), PublicKey: string(pubData)}, nil
}

func LoadConnectionsConfig() (*types.ConnectionsConfig, error) {
	data, err := os.ReadFile(ConnectionsConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.ConnectionsConfig{
				Networks:    []types.Network{},
				AllowedKeys: []string{},
			}, nil
		}
		return nil, err
	}

	var config types.ConnectionsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConnectionsConfig(config *types.ConnectionsConfig) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConnectionsConfigFile, data, 0644)
}

func GetPublicKeyDigest(publicKey string) string {
	hash := sha256.Sum256([]byte(publicKey))
	return hex.EncodeToString(hash[:])[:16]
}

func IsWireguardInstalled() bool {
	_, err := exec.LookPath("wg")
	return err == nil
}

func IsK3sInstalled() bool {
	_, err := exec.LookPath("k3s")
	return err == nil
}

func IsNginxInstalled() bool {
	_, err := exec.LookPath("nginx")
	return err == nil
}

func InstallWireguard() error {
	fmt.Println("Installing WireGuard...")

	managers := [][]string{
		{"apt", "update", "&&", "apt", "install", "-y", "wireguard"},
		{"yum", "install", "-y", "wireguard-tools"},
		{"pacman", "-S", "--noconfirm", "wireguard-tools"},
	}

	for _, manager := range managers {
		if _, err := exec.LookPath(manager[0]); err == nil {
			cmd := exec.Command("sh", "-c", fmt.Sprintf("sudo %s", exec.Command(manager[0], manager[1:]...).String()))
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("failed to install WireGuard - no supported package manager found")
}

func InstallK3s() error {
	fmt.Println("Installing K3s...")
	cmd := exec.Command("sh", "-c", "curl -sfL https://get.k3s.io | sh -")
	return cmd.Run()
}

func InstallNginx() error {
	fmt.Println("Installing Nginx...")

	managers := [][]string{
		{"apt", "update", "&&", "apt", "install", "-y", "nginx"},
		{"yum", "install", "-y", "nginx"},
		{"pacman", "-S", "--noconfirm", "nginx"},
	}

	for _, manager := range managers {
		if _, err := exec.LookPath(manager[0]); err == nil {
			cmd := exec.Command("sh", "-c", fmt.Sprintf("sudo %s", exec.Command(manager[0], manager[1:]...).String()))
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("failed to install Nginx - no supported package manager found")
}

func CreateNginxConfig(domain, serverIP, port string) error {
	configContent := fmt.Sprintf(`server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass http://%s:%s;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}`, domain, serverIP, port)

	configPath := filepath.Join("/etc/nginx/sites-available", domain)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return err
	}

	enabledPath := filepath.Join("/etc/nginx/sites-enabled", domain)
	return os.Symlink(configPath, enabledPath)
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

func WriteWireguardInterfaces(serverType string) error {
	connectionsConfig, err := LoadConnectionsConfig()
	if err != nil {
		return fmt.Errorf("error loading connections: %v", err)
	}

	wgConfig, err := LoadWireguardConfig()
	if err != nil {
		return fmt.Errorf("error loading wireguard config: %v", err)
	}

	err = EnsureWireguardDir()
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
		netName := fmt.Sprintf("upduck-wg-%s-%d", serverType, nindex)
		configPath := fmt.Sprintf("%s/%s.conf", WireguardConfigDir, netName)

		tmpl, err := template.New(netName).Parse(wgTemplate)
		if err != nil {
			return err
		}

		file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
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

		config := map[string]interface{}{
			"Name":       netName,
			"PrivateKey": wgConfig.PrivateKey,
			"Address":    network.Address,
			"Peers":      peers,
		}

		err = tmpl.Execute(file, config)
		if err != nil {
			return err
		}

		err = startWireGuardInterface(netName, configPath)
		if err != nil {
			return err
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
		downCmd := exec.Command("wg-quick", "down", netName)
		if err := downCmd.Run(); err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}

	upCmd := exec.Command("wg-quick", "up", configPath)
	output, err := upCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start WireGuard interface with wg-quick: %w, output: %s", err, string(output))
	}

	_, err = wgClient.Device(netName)
	if err != nil {
		return fmt.Errorf("failed to get WireGuard interface after starting: %w", err)
	}

	return nil
}
