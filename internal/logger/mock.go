package logger

type mockLogger struct{}

func NewMockLogger() Logger {
	return &mockLogger{}
}

func (m *mockLogger) Debug(msg string, args ...any) {
	// Mock implementation for Debug
}

func (m *mockLogger) Info(msg string, args ...any) {
	// Mock implementation for Info
}

func (m *mockLogger) Warn(msg string, args ...any) {
	// Mock implementation for Warn
}

func (m *mockLogger) Error(msg string, args ...any) {
	// Mock implementation for Error
}

func (m *mockLogger) Fatal(msg string, args ...any) {
	// Mock implementation for Fatal
}

func (m *mockLogger) With(args ...any) Logger {
	return &mockLogger{}
}
