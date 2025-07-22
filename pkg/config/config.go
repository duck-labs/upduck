package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/duck-labs/upduck/pkg/types"
)

var (
	ConfigDir             = getConfigDir()
	WireguardConfigDir    = filepath.Join(ConfigDir, "wg-config")
	WireguardConfigFile   = filepath.Join(ConfigDir, "wireguard-config.json")
	ConnectionsConfigFile = filepath.Join(ConfigDir, "connections.json")
	NodeConfigFile        = filepath.Join(ConfigDir, "config.json")
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

func WriteNodeConfig(nodeType string) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	config := types.NodeConfig{
		Type: nodeType,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(NodeConfigFile, data, 0600)
}

func LoadNodeConfig() (*types.NodeConfig, error) {
	data, err := os.ReadFile(NodeConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not configured")
		}
		return nil, err
	}

	var config types.NodeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
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
				Networks:       []types.Network{},
				AllowedKeys:    []string{},
				EncryptionKeys: []types.EncryptionKey{},
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
