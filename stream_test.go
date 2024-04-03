package json

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestReadChunk(t *testing.T) {
	testData := []byte("Hello, world! This is a test string.")
	buffer := bytes.NewBuffer(testData)

	reader := NewStreamReader(buffer, chunkSize)

	chunk, err := reader.ReadChunk()
	if err != nil && err != io.EOF {
		t.Errorf("Unexpected error: %v", err)
	}

	if !bytes.Equal(chunk, testData) {
		t.Errorf("ReadChunk() = %v, want %v", chunk, testData)
	}

	_, err = reader.ReadChunk()
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestReadChunk_SizeBasedSplitting(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		chunkSize      int
		expectedChunks []string
		expectedError  error
	}{
		{
			name:      "Exact Chunk Size",
			input:     "abcdefghijklmnop",
			chunkSize: 4,
			expectedChunks: []string{
				"abcd",
				"efgh",
				"ijkl",
				"mnop",
			},
			expectedError: nil,
		},
		{
			name:      "Smaller Than Chunk Size",
			input:     "abcdefg",
			chunkSize: 10,
			expectedChunks: []string{
				"abcdefg",
			},
			expectedError: nil,
		},
		{
			name:      "Larger Than Chunk Size",
			input:     "abcdefghijklmnopqrstuvwxyz",
			chunkSize: 10,
			expectedChunks: []string{
				"abcdefghij",
				"klmnop",
				"qrstuvwxyz",
			},
			expectedError: nil,
		},
		{
			name:           "Empty Input",
			input:          "",
			chunkSize:      4,
			expectedChunks: []string{},
			expectedError:  io.EOF,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.input)
			streamReader := NewStreamReader(reader, test.chunkSize)

			var chunks []string
			for {
				chunk, err := streamReader.ReadChunk()
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Fatalf("Unexpected error: %v", err)
				}
				chunks = append(chunks, string(chunk))
			}

			if len(chunks) != len(test.expectedChunks) {
				t.Fatalf("Expected %d chunks, got %d", len(test.expectedChunks), len(chunks))
			}

			if !reflect.DeepEqual(chunks, test.expectedChunks) {
				t.Errorf("ReadChunk() = %v, want %v", chunks, test.expectedChunks)
			}
		})
	}
}