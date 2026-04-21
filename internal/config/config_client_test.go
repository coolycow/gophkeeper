package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadFromFlagsAndEnvClient парсит флаги и затем применяет переменные окружения (как в ручных сценариях тестирования).
func loadFromFlagsAndEnvClient(t *testing.T, args []string) *ConfigClient {
	t.Helper()
	cfg, err := InitConfigClientWithArgs(args)
	require.NoError(t, err)
	cfg, err = initConfigClientWithEnv(cfg)
	require.NoError(t, err)
	return cfg
}

func TestInitConfigClientWithArgs_defaults(t *testing.T) {
	cfg, err := InitConfigClientWithArgs(nil)
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1:8080", cfg.ServerAddress)
	assert.Equal(t, "", cfg.Email)
	assert.Equal(t, getDefaultConfigFile(), cfg.Config)
}

func TestInitConfigClientWithArgs_flags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want func(*testing.T, *ConfigClient)
	}{
		{
			name: "server address",
			args: []string{"--server-address", "10.0.0.1:9090"},
			want: func(t *testing.T, c *ConfigClient) {
				assert.Equal(t, "10.0.0.1:9090", c.ServerAddress)
			},
		},
		{
			name: "server address short",
			args: []string{"-a", "192.168.1.1:443"},
			want: func(t *testing.T, c *ConfigClient) {
				assert.Equal(t, "192.168.1.1:443", c.ServerAddress)
			},
		},
		{
			name: "email",
			args: []string{"--email", "user@example.com"},
			want: func(t *testing.T, c *ConfigClient) {
				assert.Equal(t, "user@example.com", c.Email)
			},
		},
		{
			name: "email short",
			args: []string{"-m", "short@mail.test"},
			want: func(t *testing.T, c *ConfigClient) {
				assert.Equal(t, "short@mail.test", c.Email)
			},
		},
		{
			name: "config path",
			args: []string{"--config", "custom-client.json"},
			want: func(t *testing.T, c *ConfigClient) {
				assert.Equal(t, "custom-client.json", c.Config)
			},
		},
		{
			name: "config path short",
			args: []string{"-c", "other.json"},
			want: func(t *testing.T, c *ConfigClient) {
				assert.Equal(t, "other.json", c.Config)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := InitConfigClientWithArgs(tt.args)
			require.NoError(t, err)
			tt.want(t, cfg)
		})
	}
}

func TestInitConfigClientWithEnv_supportedVars(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want func(*testing.T, *ConfigClient)
	}{
		{
			name: "server address and email",
			env: map[string]string{
				"SERVER_ADDRESS": "grpc.example.com:50051",
				"EMAIL":          "env@example.com",
			},
			want: func(t *testing.T, c *ConfigClient) {
				assert.Equal(t, "grpc.example.com:50051", c.ServerAddress)
				assert.Equal(t, "env@example.com", c.Email)
			},
		},
		{
			name: "config path",
			env: map[string]string{
				"CONFIG": "from-env-client.json",
			},
			want: func(t *testing.T, c *ConfigClient) {
				assert.Equal(t, "from-env-client.json", c.Config)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			cfg := loadFromFlagsAndEnvClient(t, nil)
			tt.want(t, cfg)
		})
	}
}

func TestInitConfigClientEnvOverridesFlags(t *testing.T) {
	t.Setenv("SERVER_ADDRESS", "from-env:8080")
	t.Setenv("EMAIL", "from-env@mail")

	cfg := loadFromFlagsAndEnvClient(t, []string{"--server-address", "from-flag:9999", "--email", "from-flag@mail"})
	assert.Equal(t, "from-env:8080", cfg.ServerAddress)
	assert.Equal(t, "from-env@mail", cfg.Email)
}
