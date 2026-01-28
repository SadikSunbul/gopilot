package gopilot

import (
	"io"
	"log"
	"time"
)

// Logger defines the interface for logging.
type Logger interface {
	// Debug logs a debug message.
	Debug(msg string, args ...any)
	// Info logs an info message.
	Info(msg string, args ...any)
	// Warn logs a warning message.
	Warn(msg string, args ...any)
	// Error logs an error message.
	Error(msg string, args ...any)
}

// defaultLogger is a simple logger that wraps the standard library logger.
type defaultLogger struct {
	logger *log.Logger
	level  LogLevel
}

// LogLevel represents the logging level.
type LogLevel int

const (
	// LogLevelDebug logs everything.
	LogLevelDebug LogLevel = iota
	// LogLevelInfo logs info, warn, and error.
	LogLevelInfo
	// LogLevelWarn logs warn and error.
	LogLevelWarn
	// LogLevelError logs only errors.
	LogLevelError
	// LogLevelNone disables logging.
	LogLevelNone
)

func (l *defaultLogger) Debug(msg string, args ...any) {
	if l.level <= LogLevelDebug {
		l.logger.Printf("[DEBUG] "+msg, args...)
	}
}

func (l *defaultLogger) Info(msg string, args ...any) {
	if l.level <= LogLevelInfo {
		l.logger.Printf("[INFO] "+msg, args...)
	}
}

func (l *defaultLogger) Warn(msg string, args ...any) {
	if l.level <= LogLevelWarn {
		l.logger.Printf("[WARN] "+msg, args...)
	}
}

func (l *defaultLogger) Error(msg string, args ...any) {
	if l.level <= LogLevelError {
		l.logger.Printf("[ERROR] "+msg, args...)
	}
}

// newDefaultLogger creates a new default logger.
func newDefaultLogger() *defaultLogger {
	return &defaultLogger{
		logger: log.New(io.Discard, "", 0),
		level:  LogLevelNone,
	}
}

// Option is a functional option for configuring Gopilot.
type Option func(*Gopilot)

// WithLogger sets a custom logger.
func WithLogger(logger Logger) Option {
	return func(g *Gopilot) {
		if logger != nil {
			g.logger = logger
		}
	}
}

// WithStdLogger enables standard library logging with the specified level.
func WithStdLogger(level LogLevel) Option {
	return func(g *Gopilot) {
		g.logger = &defaultLogger{
			logger: log.Default(),
			level:  level,
		}
	}
}

// WithTimeout sets the default timeout for operations.
func WithTimeout(d time.Duration) Option {
	return func(g *Gopilot) {
		if d > 0 {
			g.timeout = d
		}
	}
}

// WithSystemPromptRules sets custom rules for the system prompt.
func WithSystemPromptRules(rules []string) Option {
	return func(g *Gopilot) {
		g.systemPromptRules = rules
	}
}

// Middleware is a function that wraps function execution.
type Middleware func(next ExecuteFunc) ExecuteFunc

// ExecuteFunc is the function signature for execution.
type ExecuteFunc func(name string, params map[string]any) (any, error)

// WithMiddleware adds middleware to the execution chain.
func WithMiddleware(middlewares ...Middleware) Option {
	return func(g *Gopilot) {
		g.middlewares = append(g.middlewares, middlewares...)
	}
}

// WithUnsupportedHandler sets a custom handler for unsupported requests.
func WithUnsupportedHandler(fn FunctionWrapper) Option {
	return func(g *Gopilot) {
		if fn != nil {
			g.unsupportedHandler = fn
		}
	}
}

// RetryConfig holds retry configuration.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts.
	MaxAttempts int
	// InitialDelay is the initial delay between retries.
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration
	// Multiplier is the multiplier for exponential backoff.
	Multiplier float64
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// WithRetry enables retry with the specified configuration.
func WithRetry(config RetryConfig) Option {
	return func(g *Gopilot) {
		g.retryConfig = &config
	}
}
