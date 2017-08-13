package config

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	cfg, err := FromFile("testdata/config.yml")

	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "master", cfg.StartPage)
	assert.Equal(t, 8080, cfg.AppPort)

	assert.NotNil(t, cfg.Hosts["master"])
	assert.Equal(t, "1.1.1.1:8080", cfg.Hosts["master"].Address)
	assert.Nil(t, cfg.Hosts["master"].Forwarding)

	assert.NotNil(t, cfg.Hosts["worker-1"])
	assert.Equal(t, "2.2.2.2:8081", cfg.Hosts["worker-1"].Address)
	assert.NotNil(t, cfg.Hosts["worker-1"].Forwarding)
	assert.Equal(t, "/path/to/private_key.pem", *cfg.Hosts["worker-1"].Forwarding.PrivateKey)
	assert.Nil(t, cfg.Hosts["worker-1"].Forwarding.Password)
	assert.Equal(t, "3.3.3.3:22", cfg.Hosts["worker-1"].Forwarding.Server)
	assert.Equal(t, "ssh-user", cfg.Hosts["worker-1"].Forwarding.User)

	assert.NotNil(t, cfg.Hosts["worker-2"])
	assert.Equal(t, "4.4.4.4:4040", cfg.Hosts["worker-2"].Address)
	assert.Nil(t, cfg.Hosts["worker-2"].Forwarding)
}

func TestMissingConfig(t *testing.T) {
	_, err := FromFile("config_missing.yml")
	assert.Error(t, err)
}

func TestInvalidConfig(t *testing.T) {
	_, err := FromFile("testdata/invalid_config.yml")
	assert.Error(t, err)
}

func TestEmptyConfig(t *testing.T) {
	var bytes []byte
	_, err := NewConfig(bytes)
	assert.NoError(t, err)
}
