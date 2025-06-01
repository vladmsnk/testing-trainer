package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"testing_trainer/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Database Cleaner Cron Service...")

	err := config.InitConfigWithEnvs()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Initialize PostgreSQL connection
	pool, err := initPostgreSQLConnection(config.ConfigStruct.PG)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(log.New(os.Stdout, "CRON: ", log.LstdFlags))))

	jobID, err := c.AddFunc("0 0 */2 * *", func() {
		cleanDatabase(pool)
	})
	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}

	log.Printf("Database cleaner job scheduled with ID: %d", jobID)
	log.Println("Cron schedule: Every 2 days at midnight (0 0 */2 * *)")

	c.Start()
	log.Println("Cron scheduler started successfully")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Database Cleaner is running. Press Ctrl+C to stop.")
	<-quit

	log.Println("Shutting down Database Cleaner...")
	ctx := c.Stop()
	<-ctx.Done()
	log.Println("Database Cleaner stopped")
}

func cleanDatabase(pool *pgxpool.Pool) {
	log.Println("🧹 Starting database cleanup process...")
	startTime := time.Now()

	tablesToClean := []string{
		"goal_logs",
		"goal_stats",
		"goals",
		"habits",
		"progress_snapshots",
		"users_time",
		"tokens",
	}

	var totalRowsDeleted int64

	for _, tableName := range tablesToClean {
		rowsDeleted, err := cleanTable(pool, tableName)
		if err != nil {
			continue
		}
		totalRowsDeleted += rowsDeleted
	}

	duration := time.Since(startTime)
	log.Printf("Database cleanup completed in %v", duration)
}

func cleanTable(pool *pgxpool.Pool, tableName string) (int64, error) {
	log.Printf("🔄 Cleaning table: %s", tableName)
	ctx := context.Background()

	query := "TRUNCATE TABLE " + tableName + " RESTART IDENTITY CASCADE"

	result, err := pool.Exec(ctx, query)
	if err != nil {
		deleteQuery := "DELETE FROM " + tableName
		result, err = pool.Exec(ctx, deleteQuery)
		if err != nil {
			return 0, err
		}
	}

	rowsAffected := result.RowsAffected()
	return rowsAffected, nil
}

func initPostgreSQLConnection(cfg config.Postgres) (*pgxpool.Pool, error) {
	log.Printf("🔌 Connecting to PostgreSQL at %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.GetConnectionString())
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
