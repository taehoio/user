package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/taehoio/user/config"
	"github.com/taehoio/user/server"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	log := logrus.StandardLogger()

	setting := config.NewSetting()
	cfg := config.NewConfig(setting)

	grpcServer, err := server.NewGRPCServer(cfg)
	if err != nil {
		log.Fatal(err)
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
}
