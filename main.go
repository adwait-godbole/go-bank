package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/adwait-godbole/go-bank/api"
	db "github.com/adwait-godbole/go-bank/db/sqlc"
	_ "github.com/adwait-godbole/go-bank/doc/statik"
	"github.com/adwait-godbole/go-bank/gapi"
	"github.com/adwait-godbole/go-bank/pb"
	"github.com/adwait-godbole/go-bank/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // since our migration source is from the local fs
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	DEVELOPMENT = "development"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("failed to load config: ")
	}

	if config.Environment == DEVELOPMENT {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Msg("failed to connect to the db: ")
	}

	// run db migrations
	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewSQLStore(conn)
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
	// runGinServer(config, store)
}

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("failed to create a new golang-migrate instance: ")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migrate up: ")
	}

	log.Info().Msg("db migrations ran successfully")
}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("failed to create gRPC server: ")
	}

	// We do the below jsonOption(s) so that we get the SAME field names in the server responses
	// as we have defined in the proto files
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
		log.Fatal().Msg("failed to register handler server: ")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// The below approach of serving swagger docs will also force us to copy all static assets
	// in the Dockerfile. Hence to avoid this we are preferring to use "statik".
	// fs := http.FileServer(http.Dir("./doc/swagger"))

	statikFs, err := fs.New()
	if err != nil {
		log.Fatal().Msg("failed to create statik fs: ")
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFs))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("failed to create listener: ")
	}

	log.Info().Msgf("starting HTTP Gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal().Msg("failed to start HTTP Gateway server: ")
	}
}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("failed to create gRPC server: ")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger) // this logging only works for direct gRPC calls and does not work for HTTP Requests received via grpc-gateway
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterGoBankServer(grpcServer, server)
	reflection.Register(grpcServer) // helps the gRPC client to easily explore what RPCs are avilable on the server, and how to call them.

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msg("failed to create listener: ")
	}

	log.Info().Msgf("starting gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Msg("failed to start gRPC server: ")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("failed to create gin server: ")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("failed to start gin server: ")
	}
}
