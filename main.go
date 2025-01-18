package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/adwait-godbole/go-bank/api"
	db "github.com/adwait-godbole/go-bank/db/sqlc"
	_ "github.com/adwait-godbole/go-bank/doc/statik"
	"github.com/adwait-godbole/go-bank/gapi"
	"github.com/adwait-godbole/go-bank/pb"
	"github.com/adwait-godbole/go-bank/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("failed to connect to the db: ", err)
	}

	store := db.NewSQLStore(conn)
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
	// runGinServer(config, store)
}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("failed to create gRPC server: ", err)
	}

	// We do the below jsonOption(s) so that we get the SAME field names as we have defined
	// in the proto files back in the server responses
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterGoBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("failed to register handler server: ", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// The below approach of serving swagger docs will also force us to copy all static assets
	// in the Dockerfile. Hence to avoid this we are preferring to use "statik".
	// fs := http.FileServer(http.Dir("./doc/swagger"))

	statikFs, err := fs.New()
	if err != nil {
		log.Fatal("failed to create statik fs: ", err)
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFs))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("failed to create listener: ", err)
	}

	log.Printf("starting HTTP Gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("failed to start HTTP Gateway server: ", err)
	}
}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("failed to create gRPC server: ", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGoBankServer(grpcServer, server)
	reflection.Register(grpcServer) // helps the gRPC client to easily explore what RPCs are avilable on the server, and how to call them.

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("failed to create listener: ", err)
	}

	log.Printf("starting gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("failed to start gRPC server: ", err)
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("failed to create gin server: ", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("failed to start gin server: ", err)
	}
}
