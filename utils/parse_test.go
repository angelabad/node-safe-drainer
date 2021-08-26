package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	node := "node1"
	nodes := "node1,node2,node3"

	assert.Equal(t, ParseArgs(nodes), []string{"node1", "node2", "node3"})
	assert.Equal(t, ParseArgs(node), []string{"node1"})
}
