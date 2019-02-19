package soajsgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryPath_register(t *testing.T) {
	path := registryPath{address: "localhost"}
	assert.Equal(t, "http://localhost/register", path.register())
}
