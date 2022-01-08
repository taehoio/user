package handler

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userddlv1 "github.com/taehoio/ddl/gen/go/taehoio/ddl/services/user/v1"
	userv1 "github.com/taehoio/idl/gen/go/taehoio/idl/services/user/v1"
)

var (
	ErrMismatchedHashAndPassword = status.Error(codes.Unauthenticated, "unauthorized")
)

type SignInHandlerFunc func(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error)

func SignIn(db *sql.DB) SignInHandlerFunc {
	return func(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error) {
		um := &userddlv1.User{}
		u, err := um.FindOneByProvideAndIdentifier(
			db,
			userddlv1.Provider_PROVIDER_EMAIL,
			req.GetEmail(),
		)
		if err != nil {
			return nil, err
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

		return &userv1.SignInResponse{
			AccessToken: u.Identifier,
		}, nil
	}
}
