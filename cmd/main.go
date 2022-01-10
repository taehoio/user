package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/profiler"
	cloudtrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"

	"github.com/taehoio/user/config"
	"github.com/taehoio/user/server"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logger := logrus.StandardLogger()

	setting := config.NewSetting()
	cfg := config.NewConfig(setting, logger)

	if err := runServer(cfg); err != nil {
		logger.Fatal(err)
	}
}

func runServer(cfg config.Config) error {
	log := cfg.Logger()

	if cfg.Setting().ShouldProfile {
		if err := setUpProfiler(cfg.Setting().ServiceName); err != nil {
			return err
		}
	}

	if cfg.Setting().ShouldTrace {
		tp, err := setUpTracing()
		if err != nil {
			return err
		}
		defer tp.ForceFlush(context.Background())
	}

	grpcServer, err := server.NewGRPCServer(cfg)
	if err != nil {
		return err
	}

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.Setting().GRPCServerPort)
		if err != nil {
			log.Fatal(err)
		}

		log.WithField("port", cfg.Setting().GRPCServerPort).Info("starting user gRPC server")
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	<-quit

	time.Sleep(time.Duration(cfg.Setting().GracefulShutdownTimeoutMs) * time.Millisecond)

	log.Info("Stopping user gRPC server")
	grpcServer.GracefulStop()

	return nil
}

func setUpProfiler(serviceName string) error {
	pc := profiler.Config{
		Service: serviceName,
	}
	if err := profiler.Start(pc); err != nil {
		return err
	}
	return nil
}

func setUpTracing() (*sdktrace.TracerProvider, error) {
	exporter, err := cloudtrace.New()
	if err != nil {
		return nil, nil
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)

	return tp, nil
}
