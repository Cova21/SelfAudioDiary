package rabbitmq

import "testing"

func TestHealthCheck_ClosedConnectionReturnsError(t *testing.T) {
	// Arrange: создаём клиент с пустым соединением.
	c := &Client{}
	// Act: вызываем HealthCheck.
	if err := c.HealthCheck(); err == nil {
		// Assert: должна вернуться ошибка, так как соединение отсутствует.
		t.Fatal("expected health check error for nil connection")
	}
}
