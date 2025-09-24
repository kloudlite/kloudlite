package logger_test

import (
	"testing"

	"github.com/kloudlite/api/v2/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		environment string
		shouldError bool
	}{
		{
			name:        "development environment with debug level",
			level:       "debug",
			environment: "development",
			shouldError: false,
		},
		{
			name:        "production environment with info level",
			level:       "info",
			environment: "production",
			shouldError: false,
		},
		{
			name:        "invalid log level",
			level:       "invalid",
			environment: "development",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.New(tt.level, tt.environment)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, log)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, log)
			}
		})
	}
}