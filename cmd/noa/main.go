package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	noa "github.com/brumhard/NoA/internal/noa"
	"github.com/brumhard/alligotor"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

const httpTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// TODO: add graceful shutdown
func run() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}

	cfg := struct {
		Port int
		TLS  struct {
			Cert string
			Key  string
		}
		Annotations []string
	}{
		Port: 8443,
	}

	cfgReader := alligotor.New(alligotor.NewEnvSource("NOA"))

	if err := cfgReader.Get(&cfg); err != nil {
		return err
	}

	logger.Info("got config", zap.Any("value", cfg))

	admissionHandler := noa.NewHandler(logger, cfg.Annotations)

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      admissionHandler,
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout,
	}

	logger.Info("starting application", zap.Int("port", cfg.Port))
	return s.ListenAndServeTLS(cfg.TLS.Cert, cfg.TLS.Key)
}
