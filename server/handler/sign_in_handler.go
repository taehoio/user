package handler

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userddlv1 "github.com/taehoio/ddl/gen/go/taehoio/ddl/services/user/v1"
	authv1 "github.com/taehoio/idl/gen/go/taehoio/idl/services/auth/v1"
	userv1 "github.com/taehoio/idl/gen/go/taehoio/idl/services/user/v1"
)

var (
	ErrMismatchedHashAndPassword = status.Error(codes.Unauthenticated, "unauthorized")
	ErrUserNotFound              = status.Error(codes.NotFound, "user not found")
)

type SignInHandlerFunc func(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error)

func SignIn(db *sql.DB, authCli authv1.AuthServiceClient) SignInHandlerFunc {
	return func(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error) {
		um := &userddlv1.User{}
		tracer := otel.GetTracerProvider()
		ctx, span := tracer.Tracer("github.com/taehoio/user").Start(ctx, "SignIn.FindOneByProvideAndIdentifier")
		u, err := um.FindOneByProvideAndIdentifier(
			db,
			userddlv1.Provider_PROVIDER_EMAIL,
			req.GetEmail(),
		)
		span.End()
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, ErrUserNotFound
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(u.PasswordHash.GetValue()),
			[]byte(req.GetPassword()),
		); err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				return nil, ErrMismatchedHashAndPassword
			}
			return nil, err
		}

		authReq := &authv1.AuthRequest{
			Provider:   userv1.Provider_PROVIDER_EMAIL,
			Identifier: req.GetEmail(),
		}
		authResp, err := authCli.Auth(ctx, authReq)
		if err != nil {
			return nil, err
		}

		return &userv1.SignInResponse{
			AccessToken:  authResp.AccessToken,
			RefreshToken: authResp.RefreshToken,
		}, nil
	}
}
