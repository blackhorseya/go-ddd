// Package logx provides a structured logging wrapper around log/slog
// with configuration support and contextx integration.
package logx

// Format defines log output format.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Output defines log output destination.
type Output string

const (
	OutputStdout Output = "stdout"
	OutputStderr Output = "stderr"
)

// Level defines log level.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Config defines logger configuration.
// This struct is designed to be embedded in infrastructure/config.
type Config struct {
	// Level is the minimum log level: debug, info, warn, error.
	// Default: info
	Level string `mapstructure:"level" json:"level" yaml:"level"`

	// Format is the output format: json, text.
	// Default: json
	Format string `mapstructure:"format" json:"format" yaml:"format"`

	// Output is the output destination: stdout, stderr.
	// Default: stdout
	Output string `mapstructure:"output" json:"output" yaml:"output"`

	// AddSource adds source file and line number to log entries.
	// Default: false (disabled for performance in production)
	AddSource bool `mapstructure:"add_source" json:"add_source" yaml:"add_source"`
}

// Default values.
const (
	DefaultLevel  = "info"
	DefaultFormat = "json"
	DefaultOutput = "stdout"
)

// defaultConfig returns configuration with default values.
func defaultConfig() *Config {
	return &Config{
		Level:     DefaultLevel,
		Format:    DefaultFormat,
		Output:    DefaultOutput,
		AddSource: false,
	}
}
