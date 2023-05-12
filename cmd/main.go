package main

import (
	"github.com/cbotte21/hive-go/internal"
	"github.com/cbotte21/hive-go/pb"
	schema2 "github.com/cbotte21/hive-go/schema"
	judicial "github.com/cbotte21/judicial-go/pb"
	"github.com/cbotte21/microservice-common/pkg/datastore"
	"github.com/cbotte21/microservice-common/pkg/enviroment"
	"github.com/cbotte21/microservice-common/pkg/jwtParser"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	// Verify environment variables exist
	enviroment.VerifyEnvVariable("port")
	enviroment.VerifyEnvVariable("jwt_secret")
	enviroment.VerifyEnvVariable("judicial_port")

	port := enviroment.GetEnvVariable("port")

	// Setup tcp listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port: %d", port)
	}
	grpcServer := grpc.NewServer()

	// Register handlers to attach
	redisClient := datastore.RedisClient[schema2.ActiveUser]{}
	err = redisClient.Init()
	if err != nil {
		panic(err)
	}
	jwtRedeemer := jwtParser.JwtSecret(enviroment.GetEnvVariable("jwt_secret"))
	judicialClient := judicial.NewJudicialServiceClient(getJudicialConn())

	// Initialize hive
	hive := internal.NewHive(&jwtRedeemer, &judicialClient, &redisClient)
	pb.RegisterHiveServiceServer(grpcServer, &hive)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to initialize grpc server.")
	}
}

func getJudicialConn() *grpc.ClientConn {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial("judicial:"+enviroment.GetEnvVariable("judicial_port"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf(err.Error())
	}
	return conn
}
