package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"

	"github.com/duck-labs/upduck-v2/types"
	"github.com/duck-labs/upduck-v2/utils"
)

type Server struct {
	nodeType string
	port     string
}

func NewServer(nodeType, port string) *Server {
	return &Server{
		nodeType: nodeType,
		port:     port,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/api/servers/connect", s.handleServerConnect)
	http.HandleFunc("/health", s.handleHealth)

	log.Printf("Starting UpDuck %s server on port %s", s.nodeType, s.port)
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

	// TODO: implement address and block generation
	wgAddress := "10.0.0.2/24"
	wgNetworkBlock := "10.0.0.0/24"

	response := types.ConnectResponse{
		WGPublicKey:    wgConfig.PublicKey,
		WGNetworkBlock: wgNetworkBlock,
		WGAddress:      wgAddress,
		PublicKey:      wgConfig.PublicKey,
	}

	newConnection := types.Connection{
		Type:            "server",
		PublicKeyDigest: pubKeyDigest,
		PublicKey:       request.PublicKey,
		WGPublicKey:     request.WGPublicKey,
		WGAddress:       wgAddress,
		WGNetworkBlock:  wgNetworkBlock,
	}

	connectionsConfig.Connections = append(connectionsConfig.Connections, newConnection)
	if err := utils.SaveConnectionsConfig(connectionsConfig); err != nil {
		log.Printf("Error saving connections config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Server connected: %s", utils.GetPublicKeyDigest(request.PublicKey))
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
