package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing_trainer/config"
	"testing_trainer/internal/app/auth"
	"testing_trainer/internal/app/habit"
	"testing_trainer/internal/usecase/user"
	"testing_trainer/middlewares"
	desc "testing_trainer/pkg/service"
	"testing_trainer/utils/grpc_server"
	"testing_trainer/utils/http_server"
)

func runGrpcProxyAuthService(httpCfg *http_server.Config, grpcCfg *grpc_server.Config) (*http_server.Server, error) {
	ctx := context.Background()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	mux := runtime.NewServeMux(runtime.WithForwardResponseOption(http_server.ResponseHeaderMatcher))

	if err := desc.RegisterServiceHandlerFromEndpoint(ctx, mux, grpcCfg.Host+":"+strconv.Itoa(grpcCfg.Port), opts); err != nil {
		return nil, err

	}
	httpServer, err := http_server.New(httpCfg, mux)
	if err != nil {
		return nil, err
	}
	log.Printf("started http server at %s:%s", httpCfg.Host, strconv.Itoa(httpCfg.Port))
	return httpServer, err
}

func runGRPCServer(implementation desc.ServiceServer, cfg *grpc_server.Config) (*grpc_server.Server, error) {
	grpcServer, err := grpc_server.NewGRPCServer(cfg)
	if err != nil {
		return nil, err
	}

	desc.RegisterServiceServer(grpcServer.Ser, implementation)
	grpcServer.Run()
	log.Printf("started grpc server at %s:%s", cfg.Host, strconv.Itoa(cfg.Port))
	return grpcServer, nil
}

func setupRouter(userUc user.UseCase, habitUc habit.UseCase) *gin.Engine {
	r := gin.Default()

	authHandlers := r.Group("/api/auth")
	auth.NewAuthHandler(authHandlers, userUc)

	protectedHabitHandlers := r.Group("/api/habit")
	protectedHabitHandlers.Use(middlewares.AuthMiddleware())
	habit.NewHabitHandler(protectedHabitHandlers, habitUc)
	return r
}

func initPostgreSQLConnection(config config.Postgres) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), config.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	return pool, nil
}
