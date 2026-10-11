package epubzip

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloseReportsWriteError(t *testing.T) {
	z, err := New(filepath.Join(t.TempDir(), "comic.epub"))
	require.NoError(t, err)
	require.NoError(t, z.WriteMagic())

	// the end of the zip cannot be written
	require.NoError(t, z.w.Close())
	assert.Error(t, z.Close())
}
