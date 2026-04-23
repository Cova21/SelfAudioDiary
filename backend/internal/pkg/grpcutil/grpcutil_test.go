package grpcutil

import "testing"

func TestCloseConnectionNil_NoPanic(t *testing.T) {
	// Act: вызываем закрытие nil-соединения.
	CloseConnection(nil)
	// Assert: отсутствие panic считается успешным результатом теста.
}
