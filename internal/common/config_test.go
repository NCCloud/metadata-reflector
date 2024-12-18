package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	assert.NotEqual(t, *config, Config{})
}
