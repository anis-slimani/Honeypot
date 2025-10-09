package honeypot

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

// SSHServer représente le serveur SSH du honeypot
type SSHServer struct {
	config     *config.Config
	logger     logger.Logger
	db         *sql.DB
	listener   net.Listener
	connections map[string]*SSHConnection
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// SSHConnection représente une connexion SSH active
type SSHConnection struct {
	ID         int
	RemoteAddr string
	Username   string
	Password   string
	Success    bool
	ConnectedAt time.Time
	Session    *ssh.Session
	Channel    ssh.Channel
	Requests   <-chan *ssh.Request
}

// NewSSHServer crée une nouvelle instance du serveur SSH
func NewSSHServer(cfg *config.Config, log logger.Logger, db *sql.DB) *SSHServer {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SSHServer{
		config:      cfg,
		logger:      log,
		db:          db,
		connections: make(map[string]*SSHConnection),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start démarre le serveur SSH
func (s *SSHServer) Start(ctx context.Context) error {
	// Générer une clé privée pour le serveur SSH
	privateKey, err := s.generatePrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Configuration du serveur SSH
	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: s.publicKeyCallback,
		PasswordCallback:  s.passwordCallback,
		ServerVersion:     "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.2",
	}

	// Ajouter la clé privée au serveur
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create signer: %w", err)
	}
	serverConfig.AddHostKey(signer)

	// Démarrer l'écoute
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.logger.Infof("SSH honeypot listening on %s", addr)

	// Accepter les connexions
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				s.logger.Errorf("Failed to accept connection: %v", err)
				continue
			}

			// Traiter la connexion dans une goroutine
			go s.handleConnection(conn, serverConfig)
		}
	}
}

// generatePrivateKey génère une clé privée RSA
func (s *SSHServer) generatePrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// publicKeyCallback callback pour l'authentification par clé publique
func (s *SSHServer) publicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	// Toujours refuser l'authentification par clé publique
	s.logger.Infof("Public key authentication attempt from %s for user %s", 
		conn.RemoteAddr().String(), conn.User())
	return nil, fmt.Errorf("public key authentication not allowed")
}

// passwordCallback callback pour l'authentification par mot de passe
func (s *SSHServer) passwordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()
	passwordStr := string(password)
	remoteAddr := conn.RemoteAddr().String()

	s.logger.Infof("Login attempt from %s: username=%s, password=%s", 
		remoteAddr, username, passwordStr)

	// Vérifier si c'est un utilisateur factice valide
	success := s.validateCredentials(username, passwordStr)

	// Enregistrer la tentative de connexion
	connection := &models.Connection{
		RemoteAddr:  remoteAddr,
		Username:    username,
		Password:    passwordStr,
		Success:     success,
		ConnectedAt: time.Now(),
	}

	if err := database.SaveConnection(connection); err != nil {
		s.logger.Errorf("Failed to save connection: %v", err)
	}

	// Vérifier les attaques par force brute
	if !success {
		go s.checkBruteForce(remoteAddr)
	}

	if success {
		s.logger.Infof("Successful login from %s: %s", remoteAddr, username)
		return &ssh.Permissions{}, nil
	}

	s.logger.Warnf("Failed login attempt from %s: %s", remoteAddr, username)
	return nil, fmt.Errorf("authentication failed")
}

// validateCredentials valide les identifiants contre la liste des utilisateurs factices
func (s *SSHServer) validateCredentials(username, password string) bool {
	for _, user := range s.config.Auth.FakeUsers {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

// checkBruteForce vérifie s'il y a une attaque par force brute
func (s *SSHServer) checkBruteForce(remoteAddr string) {
	isBruteForce, attempts, err := database.CheckBruteForce(
		remoteAddr, 
		s.config.Alerts.Thresholds.TimeWindow,
		s.config.Alerts.Thresholds.FailedAttempts,
	)
	
	if err != nil {
		s.logger.Errorf("Failed to check brute force: %v", err)
		return
	}

	if isBruteForce {
		s.logger.Warnf("Brute force attack detected from %s: %d attempts", remoteAddr, attempts)
		
		// Créer une alerte
		alert := &models.Alert{
			Type:       "brute_force",
			Severity:   "high",
			Message:    fmt.Sprintf("Brute force attack detected from %s", remoteAddr),
			RemoteAddr: remoteAddr,
			Details:    fmt.Sprintf("%d failed attempts in %d seconds", attempts, s.config.Alerts.Thresholds.TimeWindow),
			CreatedAt:  time.Now(),
		}

		if err := database.SaveAlert(alert); err != nil {
			s.logger.Errorf("Failed to save alert: %v", err)
		}

		// Envoyer une alerte par email si configuré
		if s.config.Alerts.Enabled {
			go s.sendEmailAlert(alert)
		}
	}
}

// sendEmailAlert envoie une alerte par email
func (s *SSHServer) sendEmailAlert(alert *models.Alert) {
	// TODO: Implémenter l'envoi d'email
	s.logger.Infof("Email alert would be sent: %s", alert.Message)
}

// handleConnection gère une connexion SSH
func (s *SSHServer) handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	// Timeout de connexion
	conn.SetDeadline(time.Now().Add(time.Duration(s.config.Server.ConnectionTimeout) * time.Second))

	// Handshake SSH
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		s.logger.Errorf("Failed to handshake: %v", err)
		return
	}
	defer sshConn.Close()

	s.logger.Infof("New SSH connection from %s for user %s", 
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
			s.logger.Errorf("Could not accept channel: %v", err)
			continue
		}

		// Créer une session factice
		session := &FakeSession{
			channel:   channel,
			requests:  requests,
			config:    s.config,
			logger:    s.logger,
			username:  sshConn.User(),
			remoteAddr: sshConn.RemoteAddr().String(),
		}

		// Démarrer la session
		go session.handle()
	}
}

// Stop arrête le serveur SSH
func (s *SSHServer) Stop() {
	s.cancel()
	if s.listener != nil {
		s.listener.Close()
	}
}
