package types

type NodeConfig struct {
	Type string `json:"node_type"`
}

type WireguardConfig struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

type RSAKeysConfig struct {
	PrivateKey string `json:"-"`
	PublicKey  string `json:"-"`
}

type Peer struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
	Address   string `json:"address,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
}

type Network struct {
	ID      string `json:"id"`
	Address string `json:"address,omitempty"`
	Peers   []Peer `json:"peers"`
}

type EncryptionKey struct {
	ID        string `json:"id"`
	Type 	string `json:"type"`
	PublicKey string `json:"public_key"`
}

type ConnectionsConfig struct {
	Networks []Network `json:"networks"`
	AllowedKeys []string     `json:"allowed_keys,omitempty"`
	EncryptionKeys []EncryptionKey `json:"encryption_keys,omitempty"`
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
	NetworkID string `json:"network_id"`
	PeerID string `json:"peer_id"`
}
