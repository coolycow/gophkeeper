package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/coolycow/gophkeeper/internal/logger"
	flag "github.com/spf13/pflag"
)

// ConfigClient Структура для хранения конфигурации клиента
type ConfigClient struct {
	// ServerAddress — адрес gRPC (host:port), не HTTP-порт сервера.
	ServerAddress string `env:"SERVER_ADDRESS" json:"server_address,omitempty"`
	Email         string `env:"EMAIL" json:"email,omitempty"`                 // Email по умолчанию
	Config        string `env:"CONFIG" json:"config,omitempty"`               // Путь к файлу конфигурации
	TLSCAFile     string `env:"TLS_CA_FILE" json:"tls_ca_file,omitempty"`     // TLSCAFile — путь к PEM корневого/промежуточного CA для проверки сертификата gRPC (пусто — системное хранилище).
	GRPCInsecure  bool   `env:"GRPC_INSECURE" json:"grpc_insecure,omitempty"` // GRPCInsecure — не использовать TLS (только для разработки; на проде нужно использовать TLS на сервере).
	ShowVersion   bool   `env:"SHOW_VERSION" json:"show_version,omitempty"`   // ShowVersion — вывести версию сборки и завершить работу (обрабатывается в main).
}

// fileConfigClient — JSON-файл; указатели задают поля, явно присутствующие в файле.
type fileConfigClient struct {
	ServerAddress *string `json:"server_address"`
	Email         *string `json:"email"`
	Config        *string `json:"config"`
	TLSCAFile     *string `json:"tls_ca_file"`
	GRPCInsecure  *bool   `json:"grpc_insecure"`
	ShowVersion   *bool   `json:"show_version"`
}

// PrintConfig выводит настройки в лог
func (c *ConfigClient) PrintConfig() {
	var b strings.Builder
	fmt.Fprintf(&b, "config: ServerAddress=%s", c.ServerAddress)
	fmt.Fprintf(&b, "config: Email=%s", c.Email)
	fmt.Fprintf(&b, "config: Config=%s", c.Config)
	logger.Log.Info(b.String())
}

// initConfigClientWithEnv получение настроек из переменных окружения.
func initConfigClientWithEnv(config *ConfigClient) (*ConfigClient, error) {
	return applyEnvToConfigClient(config, false)
}

// InitConfigClient инициализирует конфигурацию клиента
func InitConfigClient() (*ConfigClient, error) {
	args := os.Args[1:]

	// Получаем настройки из флагов
	flagCfg, fs, err := parseClientFlags(args)
	if err != nil {
		return nil, err
	}

	// Получаем настройки из переменных окружения
	configPath := strings.TrimSpace(flagCfg.Config)
	if !fs.Changed("config") {
		if v, ok := os.LookupEnv("CONFIG"); ok {
			configPath = strings.TrimSpace(v)
		}
	}

	// Получаем настройки из файла конфигурации
	cfg := defaultConfigClient()
	_, configEnvSet := os.LookupEnv("CONFIG")
	explicitConfigPath := fs.Changed("config") || configEnvSet

	if configPath != "" {
		if err := mergeConfigClientFromFile(&cfg, configPath); err != nil {
			if errors.Is(err, os.ErrNotExist) && !explicitConfigPath {
				cfg = defaultConfigClient()
			} else {
				return nil, err
			}
		}
	}

	cfg.Config = configPath

	// Применяем настройки из переменных окружения
	if _, err := applyEnvToConfigClient(&cfg, fs.Changed("config")); err != nil {
		return nil, err
	}

	applyExplicitClientFlags(&cfg, flagCfg, fs)

	// Проверяем настройки на корректность
	var errs []error

	return &cfg, errors.Join(errs...)
}

// InitConfigClientWithArgs инициализация с переданными аргументами (только флаги; как раньше для тестов).
func InitConfigClientWithArgs(args []string) (*ConfigClient, error) {
	cfg, _, err := parseClientFlags(args)
	return cfg, err
}

// applyEnvToConfigClient применяет переменные окружения. Если skipConfigFromEnv, CONFIG не трогаем
// (путь к файлу задан явно флагом -c).
func applyEnvToConfigClient(config *ConfigClient, skipConfigFromEnv bool) (*ConfigClient, error) {
	if serverAddress, present := os.LookupEnv("SERVER_ADDRESS"); present {
		config.ServerAddress = serverAddress
	}

	if Email, present := os.LookupEnv("EMAIL"); present {
		config.Email = Email
	}

	if v, present := os.LookupEnv("TLS_CA_FILE"); present {
		config.TLSCAFile = strings.TrimSpace(v)
	}

	if v, present := os.LookupEnv("GRPC_INSECURE"); present {
		config.GRPCInsecure = parseBoolEnv(v)
	}

	if v, present := os.LookupEnv("SHOW_VERSION"); present {
		config.ShowVersion = parseBoolEnv(v)
	}

	if !skipConfigFromEnv {
		if configFile, present := os.LookupEnv("CONFIG"); present {
			config.Config = strings.TrimSpace(configFile)
		}
	}

	return config, nil
}

// parseClientFlags парсит флаги для клиента
func parseClientFlags(args []string) (*ConfigClient, *flag.FlagSet, error) {
	var config ConfigClient

	flagSet := flag.NewFlagSet("main", flag.ContinueOnError)

	// Флаги для клиента (по умолчанию gRPC на 8081 рядом с HTTP 8080 — см. README).
	flagSet.StringVarP(&config.ServerAddress, "server-address", "a", "127.0.0.1:8081", "gRPC server address (host:port)")
	flagSet.StringVarP(&config.Email, "email", "m", "", "email")
	flagSet.StringVarP(&config.Config, "config", "c", getDefaultClientConfigFile(), "config file")
	flagSet.StringVar(&config.TLSCAFile, "tls-ca-file", "", "PEM CA file for gRPC TLS (optional)")
	flagSet.BoolVar(&config.GRPCInsecure, "grpc-insecure", false, "use insecure gRPC (no TLS); dev only")
	flagSet.BoolVarP(&config.ShowVersion, "version", "v", false, "print build version and exit")

	// Парсим флаги
	err := flagSet.Parse(args)
	if err != nil {
		return nil, nil, err
	}

	return &config, flagSet, nil
}

// defaultConfigClient возвращает конфигурацию клиента по умолчанию
func defaultConfigClient() ConfigClient {
	return ConfigClient{
		ServerAddress: "127.0.0.1:8081",
		Config:        "",
	}
}

// mergeConfigClientFromFile объединяет конфигурацию клиента с данными из файла
func mergeConfigClientFromFile(cfg *ConfigClient, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var fc fileConfigClient
	if err := json.Unmarshal(data, &fc); err != nil {
		return fmt.Errorf("invalid config file %s: %w", path, err)
	}

	if fc.ServerAddress != nil {
		cfg.ServerAddress = *fc.ServerAddress
	}

	if fc.Email != nil {
		cfg.Email = *fc.Email
	}

	if fc.TLSCAFile != nil {
		cfg.TLSCAFile = strings.TrimSpace(*fc.TLSCAFile)
	}

	if fc.GRPCInsecure != nil {
		cfg.GRPCInsecure = *fc.GRPCInsecure
	}

	if fc.ShowVersion != nil {
		cfg.ShowVersion = *fc.ShowVersion
	}

	return nil
}

// applyExplicitClientFlags применяет переданные флаги к конфигурации клиента
func applyExplicitClientFlags(dst *ConfigClient, src *ConfigClient, fs *flag.FlagSet) {
	if fs.Changed("server-address") {
		dst.ServerAddress = src.ServerAddress
	}
	if fs.Changed("email") {
		dst.Email = src.Email
	}
	if fs.Changed("config") {
		dst.Config = strings.TrimSpace(src.Config)
	}
	if fs.Changed("tls-ca-file") {
		dst.TLSCAFile = strings.TrimSpace(src.TLSCAFile)
	}
	if fs.Changed("grpc-insecure") {
		dst.GRPCInsecure = src.GRPCInsecure
	}
	if fs.Changed("version") {
		dst.ShowVersion = src.ShowVersion
	}
}

// parseBoolEnv распознаёт true/false/1/0 из переменной окружения.
func parseBoolEnv(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// getDefaultClientConfigFile имя JSON-конфига клиента по умолчанию (рядом с бинарником или в текущей директории).
func getDefaultClientConfigFile() string {
	return "config_client.json"
}
