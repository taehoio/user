package handler

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	userddlv1 "github.com/taehoio/ddl/gen/go/taehoio/ddl/services/user/v1"
	userv1 "github.com/taehoio/idl/gen/go/taehoio/idl/services/user/v1"
)

var (
	ErrNotSupportedProvier = status.Error(codes.Unimplemented, "not supported provider")
	ErrInvalidPassword     = status.Error(codes.InvalidArgument, "invalid password")
	ErrAlreadyExists       = status.Error(codes.AlreadyExists, "already exists")
)

type SignUpHandlerFunc func(ctx context.Context, req *userv1.SignUpRequest) (*userv1.SignUpResponse, error)

func SignUp(db *sql.DB) SignUpHandlerFunc {
	return func(ctx context.Context, req *userv1.SignUpRequest) (*userv1.SignUpResponse, error) {
		if err := validateRequest(req); err != nil {
			return nil, err
		}

		passwordHash, err := hashPassword(req.Password.GetValue())
		if err != nil {
			return nil, err
		}

		u := &userddlv1.User{
			Provider:   userddlv1.Provider(req.Provider),
			Identifier: req.Identifier,
			PasswordHash: &wrapperspb.StringValue{
				Value: passwordHash,
			},
		}

		if err := u.Save(db); err != nil {
			if err == userddlv1.ErrDuplicateEntry {
				return nil, ErrAlreadyExists
			}
			return nil, err
		}

		return &userv1.SignUpResponse{}, nil
	}
}

func validateRequest(req *userv1.SignUpRequest) error {
	if req.Provider != userv1.Provider_PROVIDER_EMAIL {
		return ErrNotSupportedProvier
	}
	if req.Password == nil || len(req.Password.GetValue()) == 0 {
		return ErrInvalidPassword
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
