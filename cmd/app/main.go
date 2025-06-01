package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing_trainer/config"
	"testing_trainer/internal/storage"
	"testing_trainer/internal/storage/transactor"
	"testing_trainer/internal/usecase/goals_checker"
	"testing_trainer/internal/usecase/habit"
	"testing_trainer/internal/usecase/progress_adder"
	"testing_trainer/internal/usecase/progress_getter"
	"testing_trainer/internal/usecase/progress_recalculator"
	"testing_trainer/internal/usecase/time_manager"
	"testing_trainer/internal/usecase/time_switcher"
	"testing_trainer/internal/usecase/user"
	"testing_trainer/scripts/migrations"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// @BasePath  /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	defer func() {
		log.Print(recover())
	}()

	err := config.InitConfigWithEnvs()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := initPostgreSQLConnection(config.ConfigStruct.PG)
	if err != nil {
		log.Fatal(err)
	}

	db := GetSqlDBFromPgxPool(pool)
	// Apply migrations
	if err := migrations.ApplyMigrations(db); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	// storages
	var (
		store = storage.NewStorage(pool)
	)

	tx, err := transactor.New(pool)
	if err != nil {
		log.Fatal(err.Error())
	}

	var (
		timeManager = time_manager.New(store)

		authUc               = user.New(store)
		progressGetter       = progress_getter.NewGetter(authUc, store, tx, timeManager)
		progressRecalculator = progress_recalculator.NewRecalculator(authUc, store, progressGetter, tx, timeManager)
		processUc            = progress_adder.New(authUc, store, progressGetter, tx, timeManager, progressRecalculator)

		habitUc        = habit.New(store, authUc, tx, timeManager, progressGetter, progressRecalculator)
		goalsCheckerUC = goals_checker.NewChecker(store, tx, timeManager, progressGetter)
		timeSwitcherUC = time_switcher.New(timeManager)
	)

	scheduler, err := runCheckGoalsScheduler(goalsCheckerUC)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		scheduler.Shutdown()
	}()
	scheduler.Start()

	router := setupRouter(authUc, habitUc, processUc, progressGetter, progressRecalculator, timeSwitcherUC)

	// Start both HTTP and HTTPS servers
	startDualServers(router)
}

func startDualServers(router http.Handler) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// HTTP Server
	httpAddr := fmt.Sprintf("%s:%d", config.ConfigStruct.HTTP.Host, config.ConfigStruct.HTTP.Port)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	// HTTPS Server (port + 1000, e.g., if HTTP is 7001, HTTPS will be 8001)
	httpsPort := config.ConfigStruct.HTTP.Port + 1000
	httpsAddr := fmt.Sprintf("%s:%d", config.ConfigStruct.HTTP.Host, httpsPort)

	httpsServer := &http.Server{
		Addr:    httpsAddr,
		Handler: router,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server
	go func() {
		log.Printf("HTTP Server starting on %s", httpAddr)
		log.Printf("Swagger is available on http://%s/swagger/index.html", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start HTTPS server
	go func() {
		certFile := getCertFile()
		keyFile := getKeyFile()

		if certFile != "" && keyFile != "" {
			log.Printf("HTTPS Server starting on %s", httpsAddr)
			log.Printf("Swagger is available on https://%s/swagger/index.html", httpsAddr)
			if err := httpsServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS server failed: %v", err)
			}
		} else {
			log.Printf("HTTPS Server starting on %s with self-signed certificate", httpsAddr)
			log.Printf("Swagger is available on https://%s/swagger/index.html", httpsAddr)
			log.Println("WARNING: Using self-signed certificate. Not recommended for production!")

			// Generate self-signed certificate
			cert, key, err := generateSelfSignedCert()
			if err != nil {
				log.Fatalf("Failed to generate self-signed certificate: %v", err)
			}

			tlsCert, err := tls.X509KeyPair(cert, key)
			if err != nil {
				log.Fatalf("Failed to load certificate: %v", err)
			}

			httpsServer.TLSConfig.Certificates = []tls.Certificate{tlsCert}

			if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS server failed: %v", err)
			}
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down servers...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if err := httpsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTPS server shutdown error: %v", err)
	}

	log.Println("Servers stopped")
}

func getCertFile() string {
	// Check environment variable first
	if cert := os.Getenv("TLS_CERT_FILE"); cert != "" {
		return cert
	}
	// Check common locations
	commonCertPaths := []string{
		"./certs/server.crt",
		"./ssl/server.crt",
		"/etc/ssl/certs/server.crt",
	}
	for _, path := range commonCertPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func getKeyFile() string {
	// Check environment variable first
	if key := os.Getenv("TLS_KEY_FILE"); key != "" {
		return key
	}
	// Check common locations
	commonKeyPaths := []string{
		"./certs/server.key",
		"./ssl/server.key",
		"/etc/ssl/private/server.key",
	}
	for _, path := range commonKeyPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func generateSelfSignedCert() ([]byte, []byte, error) {
	// Generate a private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Testing Trainer"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost"},
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return certPEM, privKeyPEM, nil
}

func GetSqlDBFromPgxPool(pool *pgxpool.Pool) *sql.DB {
	return stdlib.OpenDB(*pool.Config().ConnConfig)
}
