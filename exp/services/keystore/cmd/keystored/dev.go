// +build !aws

package main

import (
	"github.com/stellar/go/exp/services/keystore"
	"github.com/stellar/go/support/env"
)

func getConfig() *keystore.Config {
	return &keystore.Config{
		DBURL:          env.String("KEYSTORE_DATABASE_URL", "postgres:///keystore?sslmode=disable"),
		MaxIdleDBConns: env.Int("DB_MAX_IDLE_CONNS", 5),
		MaxOpenDBConns: env.Int("DB_MAX_OPEN_CONNS", 5),
	}
}
