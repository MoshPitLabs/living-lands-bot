package api

import (
	"context"
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log/slog"

	"living-lands-bot/internal/api/handlers"
	"living-lands-bot/internal/config"
	"living-lands-bot/internal/services"
)

type Server struct {
	app    *fiber.App
	addr   string
	config *config.Config
	logger *slog.Logger
}

func NewServer(cfg *config.Config, account *services.AccountService, logger *slog.Logger) *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           30 * time.Second,
	})

	app.Use(recover.New())

	s := &Server{
		app:    app,
		addr:   cfg.HTTP.Addr,
		config: cfg,
		logger: logger,
	}

	// Routes
	app.Get("/health", s.health)

	// API routes with auth
	verifyHandler := handlers.NewVerifyHandler(account, logger)
	app.Post("/api/v1/verify", s.authMiddleware, verifyHandler.Handle)

	return s
}

func (s *Server) Start() error {
	s.logger.Info("http server starting", "addr", s.addr)
	return s.app.Listen(s.addr)
}

func (s *Server) ShutdownWithContext(ctx context.Context) error {
	// Fiber's Shutdown uses a context-less shutdown. We can approximate by
	// racing Shutdown with ctx.
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.app.Shutdown()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (s *Server) health(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "healthy",
	})
}

func (s *Server) authMiddleware(c *fiber.Ctx) error {
	secret := c.Get("X-API-Secret")

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(secret), []byte(s.config.Hytale.APISecret)) != 1 {
		s.logger.Warn("unauthorized api request",
			"ip", c.IP(),
			"path", c.Path(),
		)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	return c.Next()
}
