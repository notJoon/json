package json

import (
	"bufio"
	"io"
)

const chunkSize = 4096 // 4KB

type StreamReader struct {
	reader *bufio.Reader
	size   int
}

// NewStreamReader returns a new StreamReader that reads data from the given io.Reader.
// The size parameter specifies the size of the chunks to read. If size is not specified
// or is less than or equal to 0, a default size of 4KB is used.
func NewStreamReader(r io.Reader, size int) *StreamReader {
	if size <= 0 {
		size = chunkSize
	}

	return &StreamReader{
		reader: bufio.NewReaderSize(r, size),
		size:   size,
	}
}

// ReadChunk reads data in chunks of the specified size (chunkSize).
// It returns the read data as a byte slice and an error if any occurs.
// When the end of the data is reached, it returns an io.EOF error.
func (sr *StreamReader) ReadChunk() ([]byte, error) {
	chunk := make([]byte, sr.size)
	n, err := sr.reader.Read(chunk)
	if err != nil && err != io.EOF {
		return nil, err
	}

	if n == 0 {
		return nil, io.EOF
	}

	return chunk[:n], nil
}
