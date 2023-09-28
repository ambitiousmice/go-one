package log

import "runtime/debug"

// Info is info level
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Warn is warning level
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error is error level
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Debug is debug level
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Panic logs a templated message, then calls os.Exit.
func Panic(args ...interface{}) {
	Error(string(debug.Stack()))
	logger.Panic(args...)
}

// Infof is format info level
func Infof(fmt string, args ...interface{}) {
	logger.Infof(fmt, args...)
}

// Warnf is format warning level
func Warnf(fmt string, args ...interface{}) {
	logger.Warnf(fmt, args...)
}

// Errorf is format error level
func Errorf(fmt string, args ...interface{}) {
	Error(string(debug.Stack()))
	logger.Errorf(fmt, args...)
}

// Debugf is format debug level
func Debugf(fmt string, args ...interface{}) {
	logger.Debugf(fmt, args...)
}

// Fatal logs a message, then calls os.Exit.
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalf logs a templated message, then calls os.Exit.
func Fatalf(fmt string, args ...interface{}) {
	Error(string(debug.Stack()))
	logger.Fatalf(fmt, args...)
}

// Panicf logs a templated message, then calls os.Exit.
func Panicf(fmt string, args ...interface{}) {
	Error(string(debug.Stack()))
	logger.Panicf(fmt, args...)
}
