package ftp

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net"
	"path"
	"strings"
	"sync"
	"time"

	"honey/internal/config"
	"honey/internal/logger"
	"honey/internal/models"
)

// FTPServer représente le serveur honeypot FTP
type FTPServer struct {
	config       *config.FTPConfig
	logger       logger.Logger
	db           *sql.DB
	listener     net.Listener
	mu           sync.RWMutex
	sessions     map[string]*FTPSession
	alertManager AlertManager
}

// AlertManager interface pour envoyer des alertes
type AlertManager interface {
	SendFTPAlert(alertType, severity, message, remoteAddr, username, details string)
}

// FTPSession représente une session FTP active
type FTPSession struct {
	ID           int
	Conn         net.Conn
	RemoteAddr   string
	Username     string
	Password     string
	Authenticated bool
	CurrentDir   string
	LoginAttempts int
	StartTime    time.Time
	LastActivity time.Time
	Commands     []string
}

// New crée un nouveau serveur FTP honeypot
func New(cfg *config.FTPConfig, log logger.Logger, database *sql.DB) *FTPServer {
	return &FTPServer{
		config:   cfg,
		logger:   log,
		db:       database,
		sessions: make(map[string]*FTPSession),
	}
}

// SetAlertManager définit le gestionnaire d'alertes
func (f *FTPServer) SetAlertManager(am AlertManager) {
	f.alertManager = am
}

// Start démarre le serveur FTP honeypot
func (f *FTPServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", f.config.Host, f.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("échec du démarrage du serveur FTP: %w", err)
	}

	f.listener = listener
	f.logger.Infof("Serveur FTP honeypot démarré sur %s", addr)

	// Accepter les connexions
	go f.acceptConnections(ctx)

	<-ctx.Done()
	return f.listener.Close()
}

// acceptConnections accepte les nouvelles connexions FTP
func (f *FTPServer) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := f.listener.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					f.logger.Errorf("Erreur lors de l'acceptation de connexion FTP: %v", err)
				}
				continue
			}

			go f.handleConnection(ctx, conn)
		}
	}
}

// handleConnection gère une connexion FTP
func (f *FTPServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	f.logger.Infof("Nouvelle connexion FTP depuis %s", remoteAddr)

	// Créer une session
	session := &FTPSession{
		Conn:         conn,
		RemoteAddr:   remoteAddr,
		CurrentDir:   "/",
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Commands:     make([]string, 0),
	}

	// Enregistrer la connexion dans la base de données
	sessionID, err := f.logConnection(session)
	if err != nil {
		f.logger.Errorf("Erreur lors de l'enregistrement de la connexion FTP: %v", err)
	}
	session.ID = sessionID

	// Ajouter à la liste des sessions
	f.mu.Lock()
	f.sessions[remoteAddr] = session
	f.mu.Unlock()

	defer func() {
		f.mu.Lock()
		delete(f.sessions, remoteAddr)
		f.mu.Unlock()

		// Mettre à jour la durée de la session
		duration := time.Since(session.StartTime).Seconds()
		f.updateConnectionDuration(sessionID, int64(duration))
	}()

	// Envoyer la bannière de bienvenue
	f.sendResponse(conn, 220, "FTP Server ready")

	// Scanner pour lire les commandes
	scanner := bufio.NewScanner(conn)

	// Timeout de 5 minutes
	conn.SetDeadline(time.Now().Add(5 * time.Minute))

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			session.LastActivity = time.Now()
			conn.SetDeadline(time.Now().Add(5 * time.Minute))

			f.logger.Debugf("FTP [%s]: %s", remoteAddr, line)

			// Parser et exécuter la commande
			if !f.handleCommand(session, line) {
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		f.logger.Debugf("Erreur de lecture FTP: %v", err)
	}
}

// handleCommand traite une commande FTP
func (f *FTPServer) handleCommand(session *FTPSession, cmdLine string) bool {
	parts := strings.SplitN(cmdLine, " ", 2)
	cmd := strings.ToUpper(parts[0])
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	// Enregistrer la commande
	session.Commands = append(session.Commands, cmdLine)
	f.logCommand(session.ID, cmdLine)

	// Détecter les commandes suspectes
	f.detectSuspiciousCommand(session, cmd, args)

	switch cmd {
	case "USER":
		return f.handleUSER(session, args)
	case "PASS":
		return f.handlePASS(session, args)
	case "SYST":
		return f.handleSYST(session)
	case "PWD", "XPWD":
		return f.handlePWD(session)
	case "CWD":
		return f.handleCWD(session, args)
	case "CDUP":
		return f.handleCDUP(session)
	case "LIST", "NLST":
		return f.handleLIST(session, args)
	case "RETR":
		return f.handleRETR(session, args)
	case "STOR", "STOU":
		return f.handleSTOR(session, args)
	case "DELE":
		return f.handleDELE(session, args)
	case "MKD", "XMKD":
		return f.handleMKD(session, args)
	case "RMD", "XRMD":
		return f.handleRMD(session, args)
	case "TYPE":
		return f.handleTYPE(session, args)
	case "MODE":
		return f.handleMODE(session, args)
	case "PASV":
		return f.handlePASV(session)
	case "PORT":
		return f.handlePORT(session, args)
	case "NOOP":
		return f.handleNOOP(session)
	case "FEAT":
		return f.handleFEAT(session)
	case "OPTS":
		return f.handleOPTS(session, args)
	case "HELP":
		return f.handleHELP(session)
	case "QUIT":
		return f.handleQUIT(session)
	default:
		f.sendResponse(session.Conn, 502, "Command not implemented")
		return true
	}
}

// Commandes FTP

func (f *FTPServer) handleUSER(session *FTPSession, username string) bool {
	if username == "" {
		f.sendResponse(session.Conn, 501, "Syntax error in parameters or arguments")
		return true
	}

	session.Username = username

	// Détecter anonymous login
	if strings.ToLower(username) == "anonymous" || strings.ToLower(username) == "ftp" {
		f.logger.Warnf("Tentative de connexion anonymous depuis %s", session.RemoteAddr)
		f.createAlert(session, "anonymous_login", "WARNING", fmt.Sprintf("Anonymous FTP login attempt: %s", username))
	}

	f.sendResponse(session.Conn, 331, "Password required for "+username)
	return true
}

func (f *FTPServer) handlePASS(session *FTPSession, password string) bool {
	session.Password = password
	session.LoginAttempts++

	// Mettre à jour la connexion avec les credentials
	f.updateConnectionCredentials(session.ID, session.Username, session.Password)

	// Simuler l'authentification (toujours accepter pour honeypot)
	if session.LoginAttempts <= 3 {
		session.Authenticated = true
		f.sendResponse(session.Conn, 230, "User "+session.Username+" logged in")
		f.logger.Infof("FTP Login successful: %s@%s (pass: %s)", session.Username, session.RemoteAddr, password)

		// Détecter brute force
		if session.LoginAttempts > 1 {
			f.createAlert(session, "ftp_brute_force", "WARNING",
				fmt.Sprintf("Multiple login attempts: %d attempts", session.LoginAttempts))
		}

		return true
	}

	f.sendResponse(session.Conn, 530, "Login incorrect")
	return true
}

func (f *FTPServer) handleSYST(session *FTPSession) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}
	f.sendResponse(session.Conn, 215, "UNIX Type: L8")
	return true
}

func (f *FTPServer) handlePWD(session *FTPSession) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}
	f.sendResponse(session.Conn, 257, fmt.Sprintf("\"%s\" is current directory", session.CurrentDir))
	return true
}

func (f *FTPServer) handleCWD(session *FTPSession, dir string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	if dir == "" {
		f.sendResponse(session.Conn, 501, "Syntax error")
		return true
	}

	// Simuler le changement de répertoire
	if strings.HasPrefix(dir, "/") {
		session.CurrentDir = path.Clean(dir)
	} else {
		session.CurrentDir = path.Clean(path.Join(session.CurrentDir, dir))
	}

	f.sendResponse(session.Conn, 250, "Directory changed to "+session.CurrentDir)
	return true
}

func (f *FTPServer) handleCDUP(session *FTPSession) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	session.CurrentDir = path.Dir(session.CurrentDir)
	f.sendResponse(session.Conn, 250, "Directory changed to "+session.CurrentDir)
	return true
}

func (f *FTPServer) handleLIST(session *FTPSession, args string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	// Le honeypot simule simplement une réponse réussie
	// Note: Un vrai serveur FTP ouvrirait une data connection ici
	f.sendResponse(session.Conn, 226, "Directory send OK")
	return true
}

func (f *FTPServer) handleRETR(session *FTPSession, filename string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	if filename == "" {
		f.sendResponse(session.Conn, 501, "Syntax error")
		return true
	}

	f.logger.Warnf("Tentative de téléchargement de fichier: %s depuis %s", filename, session.RemoteAddr)
	f.createAlert(session, "ftp_download_attempt", "INFO",
		fmt.Sprintf("File download attempt: %s", filename))

	// Simuler le téléchargement (fichier inexistant)
	f.sendResponse(session.Conn, 550, "File not found")
	return true
}

func (f *FTPServer) handleSTOR(session *FTPSession, filename string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	if filename == "" {
		f.sendResponse(session.Conn, 501, "Syntax error")
		return true
	}

	f.logger.Warnf("Tentative d'upload de fichier: %s depuis %s", filename, session.RemoteAddr)

	// Détecter les fichiers malveillants potentiels
	suspiciousExtensions := []string{".php", ".jsp", ".asp", ".exe", ".sh", ".bat", ".cmd", ".py", ".pl"}
	for _, ext := range suspiciousExtensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			f.createAlert(session, "malicious_upload_attempt", "CRITICAL",
				fmt.Sprintf("Suspicious file upload attempt: %s", filename))
			break
		}
	}

	// Le honeypot simule un upload réussi
	f.sendResponse(session.Conn, 226, "Transfer complete")
	return true
}

func (f *FTPServer) handleDELE(session *FTPSession, filename string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	f.logger.Warnf("Tentative de suppression: %s depuis %s", filename, session.RemoteAddr)
	f.createAlert(session, "ftp_delete_attempt", "WARNING",
		fmt.Sprintf("File deletion attempt: %s", filename))

	f.sendResponse(session.Conn, 250, "File deleted")
	return true
}

func (f *FTPServer) handleMKD(session *FTPSession, dirname string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	f.sendResponse(session.Conn, 257, fmt.Sprintf("\"%s\" directory created", dirname))
	return true
}

func (f *FTPServer) handleRMD(session *FTPSession, dirname string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	f.logger.Warnf("Tentative de suppression de répertoire: %s depuis %s", dirname, session.RemoteAddr)
	f.sendResponse(session.Conn, 250, "Directory removed")
	return true
}

func (f *FTPServer) handleTYPE(session *FTPSession, typeCode string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	f.sendResponse(session.Conn, 200, "Type set to "+typeCode)
	return true
}

func (f *FTPServer) handleMODE(session *FTPSession, mode string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	f.sendResponse(session.Conn, 200, "Mode set to "+mode)
	return true
}

func (f *FTPServer) handlePASV(session *FTPSession) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	// Simuler le mode passif (ne pas vraiment ouvrir de port)
	f.sendResponse(session.Conn, 227, "Entering Passive Mode (127,0,0,1,195,149)")
	return true
}

func (f *FTPServer) handlePORT(session *FTPSession, args string) bool {
	if !session.Authenticated {
		f.sendResponse(session.Conn, 530, "Please login with USER and PASS")
		return true
	}

	f.sendResponse(session.Conn, 200, "PORT command successful")
	return true
}

func (f *FTPServer) handleNOOP(session *FTPSession) bool {
	f.sendResponse(session.Conn, 200, "NOOP ok")
	return true
}

func (f *FTPServer) handleFEAT(session *FTPSession) bool {
	features := []string{
		"211-Features:",
		" SIZE",
		" MDTM",
		" UTF8",
		"211 End",
	}
	for _, feat := range features {
		session.Conn.Write([]byte(feat + "\r\n"))
	}
	return true
}

func (f *FTPServer) handleOPTS(session *FTPSession, args string) bool {
	if strings.HasPrefix(strings.ToUpper(args), "UTF8") {
		f.sendResponse(session.Conn, 200, "UTF8 set to on")
	} else {
		f.sendResponse(session.Conn, 451, "Option not understood")
	}
	return true
}

func (f *FTPServer) handleHELP(session *FTPSession) bool {
	help := []string{
		"214-The following commands are recognized:",
		" USER PASS SYST PWD CWD CDUP LIST NLST RETR STOR",
		" DELE MKD RMD TYPE MODE PASV PORT NOOP FEAT OPTS HELP QUIT",
		"214 Help OK",
	}
	for _, line := range help {
		session.Conn.Write([]byte(line + "\r\n"))
	}
	return true
}

func (f *FTPServer) handleQUIT(session *FTPSession) bool {
	f.sendResponse(session.Conn, 221, "Goodbye")
	return false // Fermer la connexion
}

// Fonctions utilitaires

func (f *FTPServer) sendResponse(conn net.Conn, code int, message string) {
	response := fmt.Sprintf("%d %s\r\n", code, message)
	conn.Write([]byte(response))
}

func (f *FTPServer) detectSuspiciousCommand(session *FTPSession, cmd, args string) {
	// Détecter les commandes suspectes typiques des attaques FTP
	suspiciousCommands := map[string]string{
		"SITE EXEC": "command_injection",
		"SITE CHMOD": "privilege_escalation",
		"QUOTE": "protocol_manipulation",
	}

	fullCmd := cmd + " " + args
	for pattern, attackType := range suspiciousCommands {
		if strings.Contains(strings.ToUpper(fullCmd), pattern) {
			f.logger.Warnf("Commande FTP suspecte détectée: %s depuis %s", fullCmd, session.RemoteAddr)
			f.createAlert(session, attackType, "CRITICAL",
				fmt.Sprintf("Suspicious FTP command: %s", fullCmd))
		}
	}
}

// Fonctions de base de données

func (f *FTPServer) logConnection(session *FTPSession) (int, error) {
	query := `INSERT INTO ftp_connections (remote_addr, connected_at, current_dir)
			  VALUES (?, ?, ?)`

	result, err := f.db.Exec(query, session.RemoteAddr, session.StartTime, session.CurrentDir)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	return int(id), err
}

func (f *FTPServer) updateConnectionCredentials(sessionID int, username, password string) {
	query := `UPDATE ftp_connections SET username = ?, password = ?, authenticated = 1 WHERE id = ?`
	f.db.Exec(query, username, password, sessionID)
}

func (f *FTPServer) updateConnectionDuration(sessionID int, duration int64) {
	query := `UPDATE ftp_connections SET disconnected_at = ?, duration = ? WHERE id = ?`
	f.db.Exec(query, time.Now(), duration, sessionID)
}

func (f *FTPServer) logCommand(sessionID int, command string) {
	query := `INSERT INTO ftp_commands (connection_id, command, executed_at) VALUES (?, ?, ?)`
	f.db.Exec(query, sessionID, command, time.Now())
}

func (f *FTPServer) createAlert(session *FTPSession, alertType, severity, message string) {
	details := fmt.Sprintf(`{"username": "%s", "commands": %d, "session_duration": "%s"}`,
		session.Username, len(session.Commands), time.Since(session.StartTime).String())

	alert := &models.Alert{
		Type:       alertType,
		Severity:   severity,
		Message:    message,
		RemoteAddr: session.RemoteAddr,
		Details:    details,
		CreatedAt:  time.Now(),
		Sent:       false,
	}

	query := `INSERT INTO alerts (type, severity, message, remote_addr, details, created_at, sent)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	f.db.Exec(query, alert.Type, alert.Severity, alert.Message, alert.RemoteAddr,
		alert.Details, alert.CreatedAt, alert.Sent)
	
	// Envoyer l'alerte par email si l'AlertManager est configuré
	if f.alertManager != nil {
		go f.alertManager.SendFTPAlert(alertType, severity, message, session.RemoteAddr, session.Username, details)
	}
}
