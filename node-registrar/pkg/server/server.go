package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
)

type Server struct {
	router      *gin.Engine
	db          db.DB
	network     string
	adminTwinID uint64
}

func NewServer(db db.DB, network string, adminTwinID uint64) Server {
	router := gin.Default()
	router.RedirectTrailingSlash = true

	server := Server{router, db, network, adminTwinID}
	server.SetupRoutes()

	return server
}

func (s Server) Run(addr string) error {
	server := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-quit

		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log.Info().Msg("server is shutting down")
		err := server.Shutdown(context)
		if err != nil {
			log.Error().Err(err).Msg("failed to shut down server gracefully")
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		quit <- syscall.SIGINT
		if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("server stopped gracefully")
		} else {
			log.Error().Err(err).Msg("server stopped unexpectedly")
		}
	}
	wg.Wait()

	return err
}
