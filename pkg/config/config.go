package config

import (
	"time"

	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/registry"
)

// ReconcilerConfig controls the background reconciler that enqueues messages for adapters.
type ReconcilerConfig struct {
	// Enabled turns the reconciler on or off (default: true).
	Enabled bool `mapstructure:"enabled" json:"enabled"`
	// PollInterval is how often the reconciler scans all resources (default: 30s).
	PollInterval time.Duration `mapstructure:"poll_interval" json:"poll_interval"`
	// StaleThreshold is the maximum time a resource can go without a successful
	// reconciliation before the reconciler re-triggers adapters (default: 30m).
	StaleThreshold time.Duration `mapstructure:"stale_threshold" json:"stale_threshold"`
}

func NewReconcilerConfig() *ReconcilerConfig {
	return &ReconcilerConfig{
		Enabled:        true,
		PollInterval:   5 * time.Second,
		StaleThreshold: 30 * time.Minute,
	}
}

// ApplicationConfig holds all application configuration
// Follows HyperFleet Configuration Standard with validation and structured marshaling
type ApplicationConfig struct {
	Server     *ServerConfig               `mapstructure:"server" json:"server" validate:"required"`
	Metrics    *MetricsConfig              `mapstructure:"metrics" json:"metrics" validate:"required"`
	Health     *HealthConfig               `mapstructure:"health" json:"health" validate:"required"`
	Database   *DatabaseConfig             `mapstructure:"database" json:"database" validate:"required"`
	Logging    *LoggingConfig              `mapstructure:"logging" json:"logging" validate:"required"`
	Entities   []registry.EntityDescriptor `mapstructure:"entities" json:"entities"`
	Reconciler *ReconcilerConfig           `mapstructure:"reconciler" json:"reconciler,omitempty"`
}

// NewApplicationConfig returns default ApplicationConfig with all sub-configs initialized
// These defaults can be overridden by config file, env vars, or CLI flags
func NewApplicationConfig() *ApplicationConfig {
	return &ApplicationConfig{
		Server:     NewServerConfig(),
		Metrics:    NewMetricsConfig(),
		Health:     NewHealthConfig(),
		Database:   NewDatabaseConfig(),
		Logging:    NewLoggingConfig(),
		Reconciler: NewReconcilerConfig(),
	}
}
