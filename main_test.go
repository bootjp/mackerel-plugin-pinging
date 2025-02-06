package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRound(t *testing.T) {
	assert.Equal(t, int64(0), round(0.5))
	assert.Equal(t, int64(1), round(1.5))
	assert.Equal(t, int64(2), round(2.5))
}

func TestResolveHost(t *testing.T) {
	t.Run("localhost", func(t *testing.T) {
		pinger, err := resolveHost("localhost")
		assert.NoError(t, err)
		assert.NotNil(t, pinger)
		defer pinger.Stop()
	})

	t.Run("invalid", func(t *testing.T) {
		pinger, err := resolveHost("invalid")
		assert.Error(t, err)
		defer pinger.Stop()
	})

	t.Run("example.com", func(t *testing.T) {
		pinger, err := resolveHost("example.com")
		assert.NoError(t, err)
		assert.NotNil(t, pinger)
		defer pinger.Stop()
	})
}

func TestRttMilliSec(t *testing.T) {
	duration := time.Duration(1500 * time.Millisecond)
	assert.Equal(t, 1500.0, rttMilliSec(duration))
}

func TestGetStats(t *testing.T) {
	opts := cmdOpts{
		Host:      "localhost",
		Timeout:   1000,
		Interval:  10,
		Count:     10,
		Size:      56,
		KeyPrefix: "test",
	}

	err := getStats(opts)
	assert.NoError(t, err)

}

func TestMainFunction(t *testing.T) {
	// Mock os.Args for testing
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"cmd", "--host=localhost", "--key-prefix=test"}

	exitCode := _main()
	assert.Equal(t, 0, exitCode)
}
