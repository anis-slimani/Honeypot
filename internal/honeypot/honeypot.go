package honeypot

import (
	"context"
	"database/sql"
	"honey/internal/config"
	"honey/internal/logger"
)

// Honeypot représente le honeypot principal
type Honeypot struct {
	config *config.Config
	logger logger.Logger
	db     *sql.DB
	server *SSHServer
}

// New crée une nouvelle instance du honeypot
func New(cfg *config.Config, log logger.Logger, db *sql.DB) *Honeypot {
	return &Honeypot{
		config: cfg,
		logger: log,
		db:     db,
		server: NewSSHServer(cfg, log, db),
	}
}

// Start démarre le honeypot
func (h *Honeypot) Start(ctx context.Context) error {
	return h.server.Start(ctx)
}

// Stop arrête le honeypot
func (h *Honeypot) Stop() {
	h.server.Stop()
}
