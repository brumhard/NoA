package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	noa "github.com/brumhard/NoA/internal/noa"
	"go.uber.org/zap"
)

const httpTimeout = 10 * time.Second

// TODO: add control loop to check for existing secrets with the annotation
// TODO: make proper TODO file, define stuff
// TODO: add kubernetesx deployment files
// TODO: add graceful shutdown
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := run(context.TODO(), logger.Sugar()); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *zap.SugaredLogger) error {
	// TODO: move to config struct, read in main, validate key and cert paths
	port := flag.Int("port", 8443, "Webhook server port.")
	cert := flag.String("certFile", "/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	key := flag.String("keyFile", "/certs/tls.key", "File containing the x509 private key to --certFile.")
	flag.Parse()

	admissionHandler, err := noa.NewHandler(logger)
	if err != nil {
		return err
	}

	logger.Infow("starting server", "port", *port)

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      admissionHandler,
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout,
	}

	return s.ListenAndServeTLS(*cert, *key)
}
