package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("", "owner/repo", nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.github)
	require.NotNil(t, client.github.BaseURL)
	assert.Equal(t, "owner", client.Owner)
	assert.Equal(t, "repo", client.Name)
}

func TestNewClient_InvalidRepository(t *testing.T) {
	client, err := NewClient("", "invalid", nil)
	require.Error(t, err)
	assert.Nil(t, client)
}
