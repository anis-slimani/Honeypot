package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logger.New(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Starting Simple SSH Honeypot...")

	// Initialize database
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	logger.Info("Database initialized successfully")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start SSH server (vrai SSH)
	go func() {
		if err := startRealSSH(ctx, cfg, logger, db); err != nil {
			logger.Errorf("SSH server error: %v", err)
		}
	}()

	logger.Infof("Simple SSH honeypot started on %s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down honeypot...")
	cancel()

	// Give services time to shutdown gracefully
	time.Sleep(2 * time.Second)
	logger.Info("Honeypot stopped")
}

// startRealSSH démarre un vrai serveur SSH
func startRealSSH(ctx context.Context, cfg *config.Config, logger logger.Logger, db *sql.DB) error {
	// Configuration SSH simple
	sshConfig := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			username := conn.User()
			passwordStr := string(password)
			remoteAddr := conn.RemoteAddr().String()

			logger.Infof("Login attempt from %s: username=%s, password=%s", 
				remoteAddr, username, passwordStr)

			// Vérifier si c'est un utilisateur factice valide
			success := validateCredentials(username, passwordStr, cfg)

			// Enregistrer la tentative de connexion
			connection := &models.Connection{
				RemoteAddr:  remoteAddr,
				Username:    username,
				Password:    passwordStr,
				Success:     success,
				ConnectedAt: time.Now(),
			}

			if err := database.SaveConnection(connection); err != nil {
				logger.Errorf("Failed to save connection: %v", err)
			}

			if success {
				logger.Infof("Successful login from %s: %s", remoteAddr, username)
				return &ssh.Permissions{}, nil
			}

			logger.Warnf("Failed login attempt from %s: %s", remoteAddr, username)
			return nil, fmt.Errorf("authentication failed")
		},
	}

	// Générer une clé privée
	privateKey, err := generatePrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create signer: %w", err)
	}
	sshConfig.AddHostKey(signer)

	// Démarrer l'écoute
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	defer listener.Close()

	logger.Infof("SSH server listening on %s", addr)

	// Accepter les connexions
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				logger.Errorf("Failed to accept connection: %v", err)
				continue
			}

			// Traiter la connexion dans une goroutine
			go handleRealSSHConnection(conn, sshConfig, logger, db)
		}
	}
}

// handleRealSSHConnection gère une vraie connexion SSH
func handleRealSSHConnection(conn net.Conn, config *ssh.ServerConfig, logger logger.Logger, db *sql.DB) {
	defer conn.Close()

	// Handshake SSH
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		logger.Errorf("Failed to handshake: %v", err)
		return
	}
	defer sshConn.Close()

	logger.Infof("New SSH connection from %s for user %s", 
		sshConn.RemoteAddr().String(), sshConn.User())

	// Traiter les requêtes globales
	go ssh.DiscardRequests(reqs)

	// Traiter les canaux
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			logger.Errorf("Could not accept channel: %v", err)
			continue
		}

		// Démarrer une vraie session shell
		go handleRealSession(channel, requests, logger, db, sshConn.User(), sshConn.RemoteAddr().String())
	}
}

// handleRealSession gère une vraie session shell
func handleRealSession(channel ssh.Channel, requests <-chan *ssh.Request, logger logger.Logger, db *sql.DB, username, remoteAddr string) {
	defer channel.Close()

	// Traiter les requêtes de la session
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				if req.WantReply {
					req.Reply(true, nil)
				}
			case "pty-req":
				if req.WantReply {
					req.Reply(true, nil)
				}
			case "window-change":
				if req.WantReply {
					req.Reply(true, nil)
				}
			default:
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}
	}()

	// Démarrer un vrai shell
	cmd := exec.Command("/bin/bash")
	cmd.Stdin = channel
	cmd.Stdout = channel
	cmd.Stderr = channel

	// Enregistrer la connexion réussie
	connection := &models.Connection{
		RemoteAddr:  remoteAddr,
		Username:    username,
		Success:     true,
		ConnectedAt: time.Now(),
	}

	if err := database.SaveConnection(connection); err != nil {
		logger.Errorf("Failed to save successful connection: %v", err)
	}

	// Exécuter le shell
	if err := cmd.Run(); err != nil {
		logger.Errorf("Shell execution error: %v", err)
	}
}

// validateCredentials valide les identifiants
func validateCredentials(username, password string, cfg *config.Config) bool {
	for _, user := range cfg.Auth.FakeUsers {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

// generatePrivateKey génère une clé privée RSA
func generatePrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
