package client

import (
	context "context"
	"fmt"

	googlemetadata "cloud.google.com/go/compute/metadata"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	authv1 "github.com/taehoio/idl/gen/go/taehoio/idl/services/auth/v1"
)

//go:generate mockgen -package client -destination ./auth_service_client_mock.go -mock_names AuthServuceClient=MockAuthServiceClient github.com/taehoio/idl/gen/go/taehoio/idl/services/auth/v1 AuthServiceClient
const serviceConfig = `{"loadBalancingPolicy":"round_robin"}`

var (
	cli authv1.AuthServiceClient

	_ authv1.AuthServiceClient = (*MockAuthServiceClient)(nil)
)

func NewAuthServiceClient(
	serviceHost string,
	shouldUseTLS bool,
	certFile string,
	isInGCP bool,
	serviceURL string,
) authv1.AuthServiceClient {
	creds := insecure.NewCredentials()
	if shouldUseTLS {
		creds, _ = credentials.NewClientTLSFromFile(certFile, "")
	}

	conn, _ := grpc.Dial(
		serviceHost,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				addIDTokenHeaderInterceptor(isInGCP, serviceURL),
			),
		),
	)

	cli = authv1.NewAuthServiceClient(conn)

	return cli
}

func getIDTokenInGCP(serviceURL string) (string, error) {
	tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceURL)
	return googlemetadata.Get(tokenURL)
}

func addIDTokenHeaderInterceptor(isInGCP bool, serviceURL string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if isInGCP {
			idToken, err := getIDTokenInGCP(serviceURL)
			if err != nil {
				return err
			}

			ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", "Bearer "+idToken)
		}
		return invoker(ctx, method, req, resp, cc, opts...)
	}
}
