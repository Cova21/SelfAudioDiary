package config

import "testing"

func TestLoad_UsesDefaults(t *testing.T) {
	// Arrange: вызываем загрузку конфигурации без дополнительных env-переопределений.
	cfg, err := Load()
	// Assert: загрузка должна завершиться без ошибки.
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Assert: проверяем ожидаемые значения по умолчанию.
	if cfg.ServiceName != "voice-diary-service" {
		t.Fatalf("expected default ServiceName, got %q", cfg.ServiceName)
	}
	if cfg.GRPCPort != "50051" {
		t.Fatalf("expected default GRPCPort, got %q", cfg.GRPCPort)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("expected default HTTPPort, got %q", cfg.HTTPPort)
	}
}

func TestLoad_ReadsCustomEnv(t *testing.T) {
	// Arrange: задаём env-переменные для переопределения значений конфигурации.
	t.Setenv("SERVICE_NAME", "custom")
	t.Setenv("RABBITMQ_USER", "rabbit_user")
	t.Setenv("RABBITMQ_PASS", "rabbit_pass")
	t.Setenv("MINIO_USE_SSL", "true")

	// Act: загружаем конфигурацию после установки env.
	cfg, err := Load()
	// Assert: ошибок загрузки быть не должно.
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Assert: проверяем, что значения взяты из env-переменных.
	if cfg.ServiceName != "custom" {
		t.Fatalf("expected ServiceName=custom, got %q", cfg.ServiceName)
	}
	if cfg.RabbitMQUser != "rabbit_user" {
		t.Fatalf("expected RabbitMQUser=rabbit_user, got %q", cfg.RabbitMQUser)
	}
	if cfg.RabbitMQPass != "rabbit_pass" {
		t.Fatalf("expected RabbitMQPass=rabbit_pass, got %q", cfg.RabbitMQPass)
	}
	if !cfg.MinIOUseSSL {
		t.Fatal("expected MinIOUseSSL=true")
	}
}

func TestHelpers(t *testing.T) {
	// Arrange: создаём тестовую конфигурацию с известными значениями.
	cfg := &Config{
		DBHost:       "localhost",
		DBPort:       "5432",
		DBUser:       "postgres",
		DBPassword:   "postgres",
		DBName:       "voice_diary",
		RedisHost:    "redis",
		RedisPort:    "6379",
		RabbitMQUser: "guest",
		RabbitMQPass: "guest",
		RabbitMQHost: "rabbitmq",
		RabbitMQPort: "5672",
	}

	// Act: вызываем helper-методы.
	dsn := cfg.GetDSN()
	// Assert: DSN должен быть непустым.
	if dsn == "" {
		t.Fatal("GetDSN() returned empty string")
	}
	// Assert: Redis-адрес должен быть сформирован корректно.
	if got := cfg.GetRedisAddr(); got != "redis:6379" {
		t.Fatalf("unexpected redis addr: %q", got)
	}
	// Assert: RabbitMQ URL должен соответствовать ожидаемому шаблону.
	if got := cfg.GetRabbitMQURL(); got != "amqp://guest:guest@rabbitmq:5672/" {
		t.Fatalf("unexpected rabbitmq url: %q", got)
	}
}
