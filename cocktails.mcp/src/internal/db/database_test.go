package db

import (
	"testing"

	"github.com/jackc/pgx/v5"

	"cezzis.com/cezzis-mcp-server/internal/config"
)

func Test_applyPostgresConnSettings_sets_tls_from_config(t *testing.T) {
	tests := []struct {
		name   string
		useTLS bool
		want   bool
	}{
		{
			name:   "tls enabled in config uses tls",
			useTLS: true,
			want:   true,
		},
		{
			name:   "tls disabled in config does not use tls",
			useTLS: false,
			want:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			connConfig, err := pgx.ParseConfig("")
			if err != nil {
				t.Fatalf("ParseConfig returned error: %v", err)
			}

			settings := &config.AppSettings{
				PostgresHost:     "psqlfs-vec-eus2-glo-shared-001.postgres.database.azure.com",
				PostgresPort:     5432,
				PostgresDBName:   "cezzis-cocktailsmcp-db-loc",
				PostgresUser:     "admin",
				PostgresPassword: "password",
				PostgresUseTLS:   test.useTLS,
			}

			applyPostgresConnSettings(connConfig, settings, settings.PostgresDBName)

			got := connConfig.TLSConfig != nil
			if got != test.want {
				t.Fatalf("TLS enabled = %t, want %t", got, test.want)
			}

			if test.want && connConfig.TLSConfig.ServerName != settings.PostgresHost {
				t.Fatalf("TLS server name = %q, want %q", connConfig.TLSConfig.ServerName, settings.PostgresHost)
			}
		})
	}
}
