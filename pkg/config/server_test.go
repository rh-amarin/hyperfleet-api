package config

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewServerConfig_DefaultOpenAPISchemaPath(t *testing.T) {
	RegisterTestingT(t)

	cfg := NewServerConfig()
	// Empty string means "not configured" — schema validation is disabled by default.
	// If a non-empty path is set, the API exits on startup if the file is missing or invalid.
	Expect(cfg.OpenAPISchemaPath).To(Equal(""))
}

func TestServerConfig_OpenAPISchemaPath_EnvVar(t *testing.T) {
	RegisterTestingT(t)

	t.Setenv("HYPERFLEET_SERVER_OPENAPI_SCHEMA_PATH", "/etc/partner/schema.yaml")

	cfg, err := LoadTestConfig(t)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg.Server.OpenAPISchemaPath).To(Equal("/etc/partner/schema.yaml"))
}
