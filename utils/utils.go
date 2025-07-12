package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/duck-labs/upduck-v2/types"
)

var (
	ConfigDir            = getConfigDir()
	WireguardConfigFile  = filepath.Join(ConfigDir, "wireguard-config.json")
	ConnectionsConfigFile = filepath.Join(ConfigDir, "connections.json")
)

func getConfigDir() string {
	if dir := os.Getenv("UPDUCK_CONFIG_DIR"); dir != "" {
		return dir
	}

	return "/etc/upduck"
}

func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir, 0755)
}

// TODO: call wireguard command to generate both keys
func GenerateWireguardKeys() (*types.WireguardConfig, error) {
	return &types.WireguardConfig{
		PrivateKey: "",
		PublicKey:  "",
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

func LoadConnectionsConfig() (*types.ConnectionsConfig, error) {
	data, err := os.ReadFile(ConnectionsConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.ConnectionsConfig{
				Connections: []types.Connection{},
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
