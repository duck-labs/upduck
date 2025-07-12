package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/duck-labs/upduck-v2/types"
	"github.com/duck-labs/upduck-v2/utils"
)

type Server struct {
	nodeType            string
	port                string
	fileWatcherCtx      context.Context
	fileWatcherCancel   context.CancelFunc
	lastConnectionsHash string
}

func NewServer(nodeType, port string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		nodeType:          nodeType,
		port:              port,
		fileWatcherCtx:    ctx,
		fileWatcherCancel: cancel,
	}
}

func (s *Server) Start() error {
	go s.watchConnectionsFile()

	http.HandleFunc("/api/servers/connect", s.handleServerConnect)
	http.HandleFunc("/health", s.handleHealth)

	log.Printf("Starting UpDuck %s server on port %s", s.nodeType, s.port)
	log.Printf("File watcher started for connections config")
	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) handleServerConnect(w http.ResponseWriter, r *http.Request) {
	if s.nodeType != "tower" {
		http.Error(w, "This endpoint is only available on tower nodes", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request types.ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	pubKeyDigest := utils.GetPublicKeyDigest(request.PublicKey)

	connectionsConfig, err := utils.LoadConnectionsConfig()
	if err != nil {
		log.Printf("Error loading connections config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	allowed := false
	for _, allowedKey := range connectionsConfig.AllowedKeys {
		if allowedKey == pubKeyDigest {
			allowed = true
			break
		}
	}

	if !allowed {
		log.Printf("Unauthorized connection attempt from public key: %s", pubKeyDigest)
		http.Error(w, "Server public key not allowed", http.StatusUnauthorized)
		return
	}

	wgConfig, err := utils.LoadWireguardConfig()
	if err != nil {
		log.Printf("Error loading WireGuard config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	wgNetworkBlock := &net.IPNet{
		IP:   net.IP{10, 5, 0, 0},
		Mask: net.CIDRMask(24, 32),
	}

	wgAddress, err := utils.GetNextAvailableNetworkAddress(connectionsConfig, wgNetworkBlock)
	if err != nil {
		log.Printf("Error generating new address: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var newNetwork types.Network

	if len(connectionsConfig.Networks) == 0 {
		newNetwork = types.Network{}
	} else {
		newNetwork = connectionsConfig.Networks[0]
	}

	for _, net := range newNetwork.Peers {
		if request.WGPublicKey == net.PublicKey {
			http.Error(w, "Already exists", http.StatusConflict)
			return
		}
	}

	newNetwork.Address = wgNetworkBlock.String()

	newPeer := types.Peer{
		PublicKey: request.WGPublicKey,
		Address:   wgAddress.String(),
	}

	newNetwork.Peers = append(newNetwork.Peers, newPeer)
	connectionsConfig.Networks = []types.Network{newNetwork}

	if err := utils.SaveConnectionsConfig(connectionsConfig); err != nil {
		log.Printf("Error saving connections config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := types.ConnectResponse{
		WGPublicKey:    wgConfig.PublicKey,
		WGNetworkBlock: wgNetworkBlock.String(),
		WGAddress:      wgAddress.String(),
		PublicKey:      wgConfig.PublicKey,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Server connected: %s", utils.GetPublicKeyDigest(request.PublicKey))
}

func (s *Server) watchConnectionsFile() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Printf("Started watching connections file: %s", utils.ConnectionsConfigFile)

	for {
		select {
		case <-s.fileWatcherCtx.Done():
			log.Printf("File watcher stopped")
			return
		case <-ticker.C:
			currentHash := s.getConnectionsFileHash()
			if currentHash != s.lastConnectionsHash && currentHash != "" {
				log.Printf("Connections file changed, triggering callback")

				err := utils.WriteWireguardInterfaces(s.nodeType)

				if err != nil {
					log.Fatalf("failed to write interface: %v", err)
					return
				}

				s.lastConnectionsHash = currentHash
			}
		}
	}
}

func (s *Server) getConnectionsFileHash() string {
	data, err := os.ReadFile(utils.ConnectionsConfigFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Failed to read connections file for hash: %v", err)
		}
		return ""
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *Server) Stop() {
	log.Printf("Stopping server...")
	if s.fileWatcherCancel != nil {
		s.fileWatcherCancel()
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"node_type": s.nodeType,
	})
}

func RunCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	return cmd.Run()
}
