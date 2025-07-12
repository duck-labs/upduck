package types

type WireguardConfig struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

type RSAKeysConfig struct {
	PrivateKey string `json:"-"`
	PublicKey  string `json:"-"`
}

type Connection struct {
	Type          string `json:"type"`
	DNS           string `json:"dns,omitempty"`
	PublicKeyDigest     string `json:"public_key_digest"`
	PublicKey     string `json:"public_key"`
	WGPublicKey   string `json:"wg_public_key"`
	WGAddress     string `json:"wg_address,omitempty"`
	WGNetworkBlock string `json:"wg_network_block,omitempty"`
}

type ConnectionsConfig struct {
	Connections []Connection `json:"connections"`
	AllowedKeys []string     `json:"allowed_keys,omitempty"`
}

type ConnectRequest struct {
	PublicKey   string `json:"public_key"`
	WGPublicKey string `json:"wg_public_key"`
}

type ConnectResponse struct {
	WGPublicKey    string `json:"wg_public_key"`
	WGNetworkBlock string `json:"wg_network_block"`
	WGAddress      string `json:"wg_address"`
	PublicKey      string `json:"public_key"`
}
