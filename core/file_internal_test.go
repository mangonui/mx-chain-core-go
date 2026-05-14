package core

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type fixedSizeReader struct {
	remaining int
}

func (reader *fixedSizeReader) Read(p []byte) (int, error) {
	if reader.remaining == 0 {
		return 0, io.EOF
	}

	if len(p) > reader.remaining {
		p = p[:reader.remaining]
	}
	for idx := range p {
		p[idx] = 'a'
	}

	reader.remaining -= len(p)
	return len(p), nil
}

func TestReadBoundedTomlMap_GrowAfterStatShouldErr(t *testing.T) {
	t.Parallel()

	reader := &fixedSizeReader{remaining: MaxTomlMapFileSize + 1}

	buffer, err := readBoundedTomlMap("test.toml", 0, reader)

	require.Nil(t, buffer)
	require.Error(t, err)
	require.Contains(t, err.Error(), "grew beyond the maximum allowed")
}
