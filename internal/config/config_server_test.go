package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	fileStoragePath := getDefaultStoragePath()
	databaseDSN := getDefaultDatabaseDSN()
	secretKey := getDefaultSecretKey()
	tlsCertFile := getDefaultTLSCertFile()
	tlsKeyFile := getDefaultTLSKeyFile()
	configFile := getDefaultConfigFile()

	tests := []struct {
		name string
		args []string
		env  map[string]string
		want Config
	}{
		{
			name: "Default",
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Host And Port",
			args: []string{"-h", "localhost", "-p", "8082"},
			want: Config{
				Host:                              "localhost",
				Port:                              8082,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Address",
			args: []string{"-a", "localhost:8083"},
			want: Config{
				Host:                              "localhost",
				Port:                              8083,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Base URL",
			args: []string{"-b", "http://localhost:8083"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://localhost:8083",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Database DSN",
			args: []string{"-d", "test"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       "test",
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "String options 1",
			args: []string{"-x", "8"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                8,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "String options 2",
			args: []string{"-x", "8,200"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                8,
				RandomStringMaxLength:             200,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "String options 3",
			args: []string{"-x", "8,200,5000"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                8,
				RandomStringMaxLength:             200,
				RandomStringMaxGenerationAttempts: 5000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Audit file",
			args: []string{"-z", "audit.json"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				AuditFile:                         "audit.json",
				AuditURL:                          "",
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Audit URL",
			args: []string{"-u", "http://127.0.0.1:8080/audit"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				AuditFile:                         "",
				AuditURL:                          "http://127.0.0.1:8080/audit",
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: all settings",
			env: map[string]string{
				"HOST":                                  "localhost",
				"PORT":                                  "8888",
				"BASE_URL":                              "https://shortener.com",
				"RANDOM_STRING_LENGTH":                  "7",
				"RANDOM_STRING_MAX_LENGTH":              "8",
				"RANDOM_STRING_MAX_GENERATION_ATTEMPTS": "9",
			},
			want: Config{
				Host:                              "localhost",
				Port:                              8888,
				BaseURL:                           "https://shortener.com",
				RandomStringLength:                7,
				RandomStringMaxLength:             8,
				RandomStringMaxGenerationAttempts: 9,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: server address and base url",
			env: map[string]string{
				"SERVER_ADDRESS": "127.0.0.2:8888",
				"BASE_URL":       "https://shortener.com",
			},
			want: Config{
				Host:                              "127.0.0.2",
				Port:                              8888,
				BaseURL:                           "https://shortener.com",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Settings priority",
			args: []string{"-a", "localhost:8083", "-b", "http://localhost:8083"},
			env: map[string]string{
				"SERVER_ADDRESS": "127.0.0.2:8888",
				"BASE_URL":       "https://shortener.com",
			},
			want: Config{
				Host:                              "127.0.0.2",
				Port:                              8888,
				BaseURL:                           "https://shortener.com",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Log level debug",
			args: []string{"-e", "debug"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "debug",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Enable HTTPS",
			args: []string{"-s", "true"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       true,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "TLS certificate file",
			args: []string{"--tls-cert-file", "tls-cert.pem"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       "tls-cert.pem",
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "TLS key file",
			args: []string{"--tls-key-file", "tls-key.pem"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        "tls-key.pem",
				Config:                            configFile,
			},
		},
		{
			name: "Config file",
			args: []string{"-c", "config.json"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            "config.json",
			},
		},
		{
			name: "Env: log level warning",
			env: map[string]string{
				"SERVER_ADDRESS": "127.0.0.2:8888",
				"BASE_URL":       "https://shortener.com",
				"LOG_LEVEL":      "warn",
			},
			want: Config{
				Host:                              "127.0.0.2",
				Port:                              8888,
				BaseURL:                           "https://shortener.com",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "warn",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: file storage path",
			env: map[string]string{
				"FILE_STORAGE_PATH": "test.json",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   "test.json",
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: database DSN",
			env: map[string]string{
				"DATABASE_DSN": "test",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       "test",
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: Secret Key",
			env: map[string]string{
				"SECRET_KEY": "test",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         "test",
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: Audit File",
			env: map[string]string{
				"AUDIT_FILE": "audit.json",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				AuditFile:                         "audit.json",
				AuditURL:                          "",
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: Audit URL",
			env: map[string]string{
				"AUDIT_URL": "http://127.0.0.1:8080/audit",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				AuditFile:                         "",
				AuditURL:                          "http://127.0.0.1:8080/audit",
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: Enable HTTPS",
			env: map[string]string{
				"ENABLE_HTTPS": "true",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       true,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: TLS certificate file",
			env: map[string]string{
				"TLS_CERT_FILE": "tls-cert.pem",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       "tls-cert.pem",
				TLSKeyFile:                        tlsKeyFile,
				Config:                            configFile,
			},
		},
		{
			name: "Env: TLS key file",
			env: map[string]string{
				"TLS_KEY_FILE": "tls-key.pem",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        "tls-key.pem",
				Config:                            configFile,
			},
		},
		{
			name: "Env: Config file",
			env: map[string]string{
				"CONFIG": "config.json",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
				SecretKey:                         secretKey,
				EnableHTTPS:                       false,
				TLSCertFile:                       tlsCertFile,
				TLSKeyFile:                        tlsKeyFile,
				Config:                            "config.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменные окружения для теста
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			// Сначала применяем настройки из флагов
			cfg, err := InitConfigWithArgs(tt.args)
			assert.NoError(t, err)

			// Затем применяем настройки из переменных окружения
			cfg, err = initConfigWithEnv(cfg)
			assert.NoError(t, err)

			assert.Equal(t, *cfg, tt.want)
		})
	}
}

// TestTrustedSubnetFromFlags проверяет, что trusted subnet устанавливается из флагов
func TestTrustedSubnetFromFlags(t *testing.T) {
	cfg, err := InitConfigWithArgs([]string{"-t", "192.168.0.0/16"})
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.0/16", cfg.TrustedSubnet)
}

// TestTrustedSubnetFromEnv проверяет, что trusted subnet устанавливается из переменных окружения
func TestTrustedSubnetFromEnv(t *testing.T) {
	t.Setenv("TRUSTED_SUBNET", "10.0.0.0/8")
	cfg, err := InitConfigWithArgs(nil)
	assert.NoError(t, err)
	cfg, err = initConfigWithEnv(cfg)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.0/8", cfg.TrustedSubnet)
}

// TestInitConfigInvalidTrustedSubnet проверяет, что ошибка возвращается при некорректном trusted subnet
func TestInitConfigInvalidTrustedSubnet(t *testing.T) {
	t.Setenv("TRUSTED_SUBNET", "not-a-cidr")
	_, err := InitConfig()
	assert.Error(t, err)
}
