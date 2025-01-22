package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/adwait-godbole/go-bank/db/sqlc"
	"github.com/adwait-godbole/go-bank/token"
	"github.com/adwait-godbole/go-bank/util"
	"github.com/adwait-godbole/go-bank/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return server
}

func newContextWithBearerToken(t testing.TB, tokenMaker token.Maker, username string, duration time.Duration) context.Context {
	t.Helper()

	accessToken, _, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	md := metadata.MD{
		authorizationHeaderKey: []string{
			fmt.Sprintf("%s %s", authorizationTypeBearer, accessToken),
		},
	}

	return metadata.NewIncomingContext(context.Background(), md)
}
