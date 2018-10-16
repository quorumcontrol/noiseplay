package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestP2P(t *testing.T) {
	err := RunP2P()
	assert.Nil(t, err)
}
