package server_test

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/kazhuravlev/lrpc/ctypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	t.Parallel()

	var id ctypes.ID
	err := json.Unmarshal([]byte(`"123"`), &id)
	require.NoError(t, err)
	assert.Equal(t, ctypes.ID("123"), id)

	bodyRes, err := json.Marshal(id)
	require.NoError(t, err)
	assert.Equal(t, []byte(`"123"`), bodyRes)
}
