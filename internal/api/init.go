package api

import (
	"log"
	"net"
	"os"

	"github.com/OnsagerHe/geoip-detector/pkg/retriever"
	pb "github.com/OnsagerHe/geoip-detector/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	portGRPC string
	ip       string
}

func initConfig() *Config {
	return &Config{
		ip: os.Getenv("IP"),
		//portGRPC: os.Getenv("PORT_GRPC"),
		portGRPC: "5001",
	}
}

type Frontend struct {
	pb.UnimplementedApiServer
	Retriever *retriever.Retriever
}

func InitServer(geoIP *retriever.Retriever) *Frontend {
	return &Frontend{
		Retriever: geoIP,
	}
}

func LaunchServer(srv *Frontend) error {
	config := initConfig()
	listener, err := net.Listen("tcp", config.ip+":"+config.portGRPC)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterApiServer(s, srv)

	log.Println("server listening at " + listener.Addr().String())

	return s.Serve(listener)

}
