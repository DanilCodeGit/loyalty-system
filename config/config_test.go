package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"loyalty/config"
)

func TestMustLoadConfig(t *testing.T) {
	err := os.Setenv("RUN_ADDRESS", "example.com:8080")
	assert.NoError(t, err)
	err = os.Setenv("DATABASE_URI", "host=example.com port=5432")
	assert.NoError(t, err)
	err = os.Setenv("SECRET_KEY", "testSecretKey")
	assert.NoError(t, err)

	cfg := config.MustLoadConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "example.com:8080", cfg.Addr)
	assert.Equal(t, "host=example.com port=5432", cfg.DB)
	assert.Equal(t, "testSecretKey", cfg.SecretKey)
}
