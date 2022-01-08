package server

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	userv1 "github.com/taehoio/idl/gen/go/services/user/v1"
	"github.com/taehoio/user/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type UserServiceServer struct {
	userv1.UserServiceServer

	cfg config.Config
}

func NewUserServiceServer(cfg config.Config) (*UserServiceServer, error) {
	return &UserServiceServer{
		cfg: cfg,
	}, nil
}

func (s *UserServiceServer) HealthCheck(ctx context.Context, in *userv1.HealthCheckRequest) (*userv1.HealthCheckResponse, error) {
	return &userv1.HealthCheckResponse{}, nil
}

func (s *UserServiceServer) SignUp(ctx context.Context, in *userv1.SignUpRequest) (*userv1.SignUpResponse, error) {
	return &userv1.SignUpResponse{}, nil
}

func (s *UserServiceServer) SignIn(ctx context.Context, in *userv1.SignInRequest) (*userv1.SignInResponse, error) {
	return &userv1.SignInResponse{}, nil
}

func NewGRPCServer(cfg config.Config) (*grpc.Server, error) {
	logrus.ErrorKey = "grpc.error"
	logrusEntry := logrus.NewEntry(logrus.StandardLogger())

	grpcServer := grpc.NewServer(
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
