package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

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

func RunCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	return cmd.Run()
}
