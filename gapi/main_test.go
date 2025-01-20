package gapi

import (
	"testing"
	"time"

	db "github.com/adwait-godbole/go-bank/db/sqlc"
	"github.com/adwait-godbole/go-bank/util"
	"github.com/adwait-godbole/go-bank/worker"
	"github.com/stretchr/testify/require"
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
