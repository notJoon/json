package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStream_LoadString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, data string
		expected   []byte
		isErr 	   bool
	}{
		{
			name: "valid JSON string",
			data : `{"key": "value"}`,
			expected: []byte(`{"key": "value"}`),
			isErr: false,
		},
		{
			name: "empty JSON string",
			data: "",
			expected: nil,
			isErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func (t *testing.T)  {
			b := &buffer{}
			err := b.LoadString(tt.data)

			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, b.data)
				assert.Equal(t, len(tt.expected), b.length)
				assert.Equal(t, 0, b.index)
				assert.Equal(t, GO, b.last)
				assert.Equal(t, GO, b.state)
			}
		})
	}
}

func TestBuffer_ParseString(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid string",
			jsonStr: `"hello"`,
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "Empty string",
			jsonStr: `""`,
			want:    "",
			wantErr: false,
		},
		{
			name:    "String with escape characters",
			jsonStr: `"hello \"world\""`,
			want:    `hello "world"`,
			wantErr: false,
		},
		{
			name:    "Invalid string - missing closing quote",
			jsonStr: `"hello`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &buffer{}
			err := b.LoadString(tt.jsonStr)
			assert.NoError(t, err)

			got, err := b.parseString()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuffer_ParseNumber(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		want    float64
		wantErr bool
	}{
		{
			name:    "Valid integer",
			jsonStr: "42",
			want:    42,
			wantErr: false,
		},
		{
			name:    "Valid float",
			jsonStr: "3.14",
			want:    3.14,
			wantErr: false,
		},
		{
			name:    "Valid negative number",
			jsonStr: "-10",
			want:    -10,
			wantErr: false,
		},
		{
			name:    "Valid exponential notation",
			jsonStr: "1e+2",
			want:    100,
			wantErr: false,
		},
		{
			name:    "Invalid number",
			jsonStr: "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &buffer{}
			err := b.LoadString(tt.jsonStr)
			assert.NoError(t, err)

			got, err := b.parseNumber()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuffer_ParseTrue(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name:    "Valid true",
			jsonStr: "true",
			wantErr: false,
		},
		{
			name:    "Invalid true",
			jsonStr: "tru",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &buffer{}
			err := b.LoadString(tt.jsonStr)
			assert.NoError(t, err)

			err = b.parseTrue()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuffer_ParseFalse(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name:    "Valid false",
			jsonStr: "false",
			wantErr: false,
		},
		{
			name:    "Invalid false",
			jsonStr: "fals",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &buffer{}
			err := b.LoadString(tt.jsonStr)
			assert.NoError(t, err)

			err = b.parseFalse()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuffer_ParseNull(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name:    "Valid null",
			jsonStr: "null",
			wantErr: false,
		},
		{
			name:    "Invalid null",
			jsonStr: "nul",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &buffer{}
			err := b.LoadString(tt.jsonStr)
			assert.NoError(t, err)

			err = b.parseNull()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuffer_ParseValue(t *testing.T) {
    tests := []struct {
        name    string
        jsonStr string
        want    interface{}
        wantErr bool
    }{
        {
            name:    "Valid string",
            jsonStr: `"hello"`,
            want:    "hello",
            wantErr: false,
        },
        {
            name:    "Valid true",
            jsonStr: "true",
            want:    true,
            wantErr: false,
        },
        {
            name:    "Valid false",
            jsonStr: "false",
            want:    false,
            wantErr: false,
        },
        {
            name:    "Valid null",
            jsonStr: "null",
            want:    nil,
            wantErr: false,
        },
        {
            name:    "Valid number",
            jsonStr: "42",
            want:    42.0,
            wantErr: false,
        },
        {
            name:    "Invalid token",
            jsonStr: "abc",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            b := &buffer{}
            err := b.LoadString(tt.jsonStr)
            assert.NoError(t, err)

            got, err := b.parseValue()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
