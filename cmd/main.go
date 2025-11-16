package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/di"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func main() {
	config := config.MustLoadConfig()

	var log zerolog.Logger
	if config.Env == "develop" {
		log = logger.NewDevelop()
	} else {
		log = logger.NewProduction()
	}

	log.Debug().Msg("debug messages on")

	router := gin.New()

	close := di.MustConfigureApp(router, config, log)
	defer close()

	server := listenRESTServer(router, log, config.RestConfig.Port)

	GracefullShutdown(server, log)
}

func listenRESTServer(r *gin.Engine, log zerolog.Logger, port int) *http.Server {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		log.Info().
			Int("port", port).
			Msg("Starting REST server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().
				Err(err).
				Msg("Failed to start REST server")
		}
	}()

	return server
}

func GracefullShutdown(server *http.Server, log zerolog.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().
		Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().
			Err(err).
			Msg("Server forced to shutdown")
	}

	log.Info().
		Msg("Server stopped")
}
