package types

type WireguardConfig struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

type RSAKeysConfig struct {
	PrivateKey string `json:"-"`
	PublicKey  string `json:"-"`
}

type Peer struct {
	PublicKey     string `json:"public_key"`
	Address     string `json:"address,omitempty"`
	Endpoint     string `json:"endpoint"`
}

type Network struct {
	NetworkBlock string `json:"network_block,omitempty"`
	Address string `json:"address,omitempty"`
	Peers []Peer `json:"peers"`
}

type ConnectionsConfig struct {
	Networks []Network `json:"networks"`
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
