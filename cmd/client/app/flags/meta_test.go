package flags

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaFlag_Set_Valid(t *testing.T) {
	var m MetaFlag

	err := m.Set("key=value")
	require.NoError(t, err)
	assert.Equal(t, "value", m["key"])

	err = m.Set("another=pair")
	require.NoError(t, err)
	assert.Equal(t, "pair", m["another"])

	err = m.Set(" spacedkey = spaced value ")
	require.NoError(t, err)
	assert.Equal(t, "spaced value", m["spacedkey"])
}

func TestMetaFlag_Set_Invalid(t *testing.T) {
	var m MetaFlag

	err := m.Set("novalue")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid meta format")

	err = m.Set("=novalue")
	require.Error(t, err)
	assert.True(t,
		strings.Contains(err.Error(), "invalid meta format") ||
			strings.Contains(err.Error(), "key is empty"),
		"unexpected error message: %s", err.Error())
}

func TestMetaFlag_String(t *testing.T) {
	m := MetaFlag{
		"key1": "value1",
		"key2": "value2",
	}

	s := m.String()
	assert.Contains(t, s, "key1=value1")
	assert.Contains(t, s, "key2=value2")
	assert.Contains(t, s, ", ")
}

func TestMetaFlag_Type(t *testing.T) {
	var m MetaFlag
	assert.Equal(t, "meta", (&m).Type())
}
