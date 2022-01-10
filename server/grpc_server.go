package server

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	authv1 "github.com/taehoio/idl/gen/go/taehoio/idl/services/auth/v1"
	userv1 "github.com/taehoio/idl/gen/go/taehoio/idl/services/user/v1"
	"github.com/taehoio/user/client"
	"github.com/taehoio/user/config"
	"github.com/taehoio/user/server/handler"
)

type UserServiceServer struct {
	userv1.UserServiceServer

	cfg     config.Config
	db      *sql.DB
	authCli authv1.AuthServiceClient
}

func NewUserServiceServer(cfg config.Config) (*UserServiceServer, error) {
	db, err := newMySQLDB(cfg)
	if err != nil {
		return nil, err
	}

	//authCli := client.NewAuthServiceClient(cfg.Setting().AuthGRPCServiceEndpoint, false, "/etc/ssl/cert.pem")
	authCli := client.NewAuthServiceClient(
		cfg.Setting().AuthGRPCServiceEndpoint,
		cfg.Setting().ShouldUseGRPCClientTLS,
		cfg.Setting().CACertFile,
		cfg.Setting().IsInGCP,
		cfg.Setting().AuthGRPCServiceURL,
	)

	return &UserServiceServer{
		cfg:     cfg,
		db:      db,
		authCli: authCli,
	}, nil
}

func newMySQLDB(cfg config.Config) (*sql.DB, error) {
	mysqlCfg := mysql.Config{
		Net:                  cfg.Setting().MysqlNetworkType,
		Addr:                 cfg.Setting().MysqlAddress,
		User:                 cfg.Setting().MysqlUser,
		Passwd:               cfg.Setting().MysqlPassword,
		DBName:               cfg.Setting().MysqlDatabaseName,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func (s *UserServiceServer) HealthCheck(ctx context.Context, req *userv1.HealthCheckRequest) (*userv1.HealthCheckResponse, error) {
	return &userv1.HealthCheckResponse{}, nil
}

func (s *UserServiceServer) SignUp(ctx context.Context, req *userv1.SignUpRequest) (*userv1.SignUpResponse, error) {
	return handler.SignUp(s.db)(ctx, req)
}

func (s *UserServiceServer) SignIn(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error) {
	return handler.SignIn(s.db, s.authCli)(ctx, req)
}

func NewGRPCServer(cfg config.Config) (*grpc.Server, error) {
	logrus.ErrorKey = "grpc.error"
	logrusEntry := logrus.NewEntry(logrus.StandardLogger())

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(
				grpc_ctxtags.WithFieldExtractor(
					grpc_ctxtags.CodeGenRequestFieldExtractor,
				),
			),
			grpc_logrus.UnaryServerInterceptor(logrusEntry),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: 30 * time.Second,
		}),
	)

	userServiceServer, err := NewUserServiceServer(cfg)
	if err != nil {
		return nil, err
	}

	userv1.RegisterUserServiceServer(grpcServer, userServiceServer)
	reflection.Register(grpcServer)

	return grpcServer, nil
}
