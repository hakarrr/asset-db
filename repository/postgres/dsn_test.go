// Copyright © by Jeff Foley 2017-2026. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresConnectionString(t *testing.T) {
	for _, tc := range []struct{ name, dsn string }{
		{"no query", "postgres://amass:pass@localhost:5432/assetdb"},
		{"empty query", "postgres://amass:pass@localhost:5432/assetdb?"},
		{"testcontainer query", "postgres://amass:pass@localhost:5432/assetdb?timezone=UTC&sslmode=disable&connect_timeout=5&lock_timeout=5000&statement_timeout=60000&statement_cache_capacity=256&application_name=assetdbtests"},
		{"trailing separator", "postgres://amass:pass@localhost:5432/assetdb?search_path=public&"},
		{"encoded query value", "postgres://amass:pass@localhost:5432/assetdb?search_path=space%20name%2Cpublic"},
		{"encoded credentials", "postgresql://amass:p%3Fa%26s%3Ds@localhost:5432/assetdb"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dsn := postgresConnectionString(tc.dsn)
			cfg, err := pgx.ParseConfig(dsn)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := pgxpool.ParseConfig(dsn); err != nil {
				t.Fatal(err)
			}
			if cfg.Host != "localhost" || cfg.Port != 5432 || cfg.Database != "assetdb" || cfg.User != "amass" {
				t.Fatalf("connection identity changed")
			}
			if cfg.TLSConfig != nil || cfg.StatementCacheCapacity != 256 || cfg.ConnectTimeout != 5*time.Second {
				t.Fatal("repository connection options changed")
			}
			for key, want := range map[string]string{"timezone": "UTC", "application_name": "asset-db", "statement_timeout": "60000", "lock_timeout": "5000"} {
				if got := cfg.RuntimeParams[key]; got != want {
					t.Errorf("%s=%q, want %q", key, got, want)
				}
			}
			if tc.name == "encoded query value" && cfg.RuntimeParams["search_path"] != "space name,public" {
				t.Fatal("query value changed")
			}
			if tc.name == "encoded credentials" && cfg.Password != "p?a&s=s" {
				t.Fatal("password changed")
			}
		})
	}
}
