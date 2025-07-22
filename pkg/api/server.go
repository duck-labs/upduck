package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/duck-labs/upduck/pkg/config"
	"github.com/duck-labs/upduck/pkg/crypto"
	"github.com/duck-labs/upduck/pkg/network"
	"github.com/duck-labs/upduck/pkg/types"
)

type Server struct {
	nodeType            string
	port                string
	fileWatcherCtx      context.Context
	fileWatcherCancel   context.CancelFunc
	lastConnectionsHash string
	httpServer          *http.Server
}

func NewServer(nodeType, port string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		nodeType:          nodeType,
		port:              port,
		fileWatcherCtx:    ctx,
		fileWatcherCancel: cancel,
		httpServer: &http.Server{
			Addr: ":" + port,
		},
	}
}

func (s *Server) Start() error {
	go s.watchConnectionsFile()

	http.HandleFunc("/api/servers/network/", s.handleServerConnect)
	http.HandleFunc("/health", s.handleHealth)

	log.Printf("Starting UpDuck %s server on port %s", s.nodeType, s.port)
	log.Printf("File watcher started for connections config")
	return s.httpServer.ListenAndServe()
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

	path := strings.TrimPrefix(r.URL.Path, "/api/servers/network/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "connect" {
		http.Error(w, "Invalid URL format. Expected: /api/servers/network/{networkID}/connect", http.StatusBadRequest)
		return
	}
	networkID := parts[0]

	var request types.ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	pubKeyDigest := crypto.GetPublicKeyDigest(request.PublicKey)

	connectionsConfig, err := config.LoadConnectionsConfig()
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

	var targetNetwork *types.Network
	var networkIndex int
	for i, netw := range connectionsConfig.Networks {
		if netw.ID == networkID {
			targetNetwork = &connectionsConfig.Networks[i]
			networkIndex = i
			break
		}
	}

	if targetNetwork == nil {
		http.Error(w, "Network not found", http.StatusNotFound)
		return
	}

	for _, peer := range targetNetwork.Peers {
		if request.WGPublicKey == peer.PublicKey {
			http.Error(w, "Server already connected to this network", http.StatusConflict)
			return
		}
	}

	wgConfig, err := config.LoadWireguardConfig()
	if err != nil {
		log.Printf("Error loading WireGuard config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, wgNetworkBlock, err := net.ParseCIDR(targetNetwork.Address)
	if err != nil {
		log.Printf("Error parsing network address: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	wgAddress, err := network.GetNextAvailableNetworkAddress(connectionsConfig, wgNetworkBlock)
	if err != nil {
		log.Printf("Error generating new address: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	newPeer := types.Peer{
		ID:        network.GenerateTimeOrderedID(),
		PublicKey: request.WGPublicKey,
		Address:   wgAddress.String(),
	}

	connectionsConfig.Networks[networkIndex].Peers = append(connectionsConfig.Networks[networkIndex].Peers, newPeer)

	newEncryptionKey := types.EncryptionKey{
		ID:        newPeer.ID,
		Type:      "network_peer",
		PublicKey: request.PublicKey,
	}

	connectionsConfig.EncryptionKeys = append(connectionsConfig.EncryptionKeys, newEncryptionKey)

	if err := config.SaveConnectionsConfig(connectionsConfig); err != nil {
		log.Printf("Error saving connections config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := types.ConnectResponse{
		WGPublicKey:    wgConfig.PublicKey,
		WGNetworkBlock: wgNetworkBlock.String(),
		WGAddress:      wgAddress.String(),
		PublicKey:      wgConfig.PublicKey,
		PeerID:         newPeer.ID,
		NetworkID:      targetNetwork.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Server connected to network %s: %s", networkID, crypto.GetPublicKeyDigest(request.PublicKey))
}

func (s *Server) watchConnectionsFile() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.fileWatcherCtx.Done():
			return
		case <-ticker.C:
			currentHash := s.getConnectionsFileHash()
			if currentHash != s.lastConnectionsHash && currentHash != "" {
				log.Printf("Connections file changed, reloading WireGuard interfaces...")
				if err := network.WriteWireguardInterfaces(s.nodeType); err != nil {
					log.Printf("Error writing WireGuard interfaces: %v", err)
				} else {
					log.Printf("WireGuard interfaces updated successfully")
				}
				s.lastConnectionsHash = currentHash
			}
		}
	}
}

func (s *Server) getConnectionsFileHash() string {
	data, err := os.ReadFile(config.ConnectionsConfigFile)
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
	s.fileWatcherCancel()

	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
		s.httpServer.Close()
	}

	log.Printf("Server stopped successfully")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"node_type": s.nodeType,
	})
}
