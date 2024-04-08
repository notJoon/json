package json

import (
	"strconv"
	"testing"
)

func TestParseFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"123", 123},
		{"-123", -123},
		{"123.456", 123.456},
		{"-123.456", -123.456},
		{"12345678.1234567890", 12345678.1234567890},
		{"-12345678.09123456789", -12345678.09123456789},
		{"0.123", 0.123},
		{"-0.123", -0.123},
		{"", -1},
		{"abc", -1},
		{"123.45.6", -1},
		{"999999999999999999999", -1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, _ := ParseFloatLiteral([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("ParseFloatLiteral(%s): got %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseFloatWithScientificNotation(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1e6", 1000000},
		{"1E6", 1000000},
		{"1.23e10", 1.23e10},
		{"1.23E10", 1.23e10},
		{"-1.23e10", -1.23e10},
		{"-1.23E10", -1.23e10},
		{"2.45e-8", 2.45e-8},
		{"2.45E-8", 2.45e-8},
		{"-2.45e-8", -2.45e-8},
		{"-2.45E-8", -2.45e-8},
		{"5e0", 5},
		{"-5e0", -5},
		{"5E+0", 5},
		{"5e+1", 50},
		{"1e-1", 0.1},
		{"1E-1", 0.1},
		{"-1e-1", -0.1},
		{"-1E-1", -0.1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFloatLiteral([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("ParseFloatLiteral(%s): got %v, want %v", tt.input, got, tt.expected)
			}

			if err != nil {
				t.Errorf("ParseFloatLiteral(%s): got error %v", tt.input, err)
			}
		})
	}
}

func TestParseFloat_May_Interoperability_Problem(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
	}{
		{"3.141592653589793238462643383279", true},
		{"1E400", false}, // TODO: should error
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseFloatLiteral([]byte(tt.input))
			if tt.shouldErr && err == nil {
				t.Errorf("ParseFloatLiteral(%s): expected error, but not error", tt.input)
			}
		})
	}
}

func benchmarkParseFloatLiteral(b *testing.B, input []byte) {
	for i := 0; i < b.N; i++ {
		ParseFloatLiteral(input)
	}
}

func benchmarkStrconvParseFloat(b *testing.B, input string) {
	for i := 0; i < b.N; i++ {
		_, err := strconv.ParseFloat(input, 64)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFloatLiteral_Simple(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("123.456"))
}

func BenchmarkStrconvParseFloat_Simple(b *testing.B) {
	benchmarkStrconvParseFloat(b, "123.456")
}

func BenchmarkParseFloatLiteral_Long(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("123456.78901"))
}

func BenchmarkStrconvParseFloat_Long(b *testing.B) {
	benchmarkStrconvParseFloat(b, "123456.78901")
}

func BenchmarkParseFloatLiteral_Negative(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("-123.456"))
}

func BenchmarkStrconvParseFloat_Negative(b *testing.B) {
	benchmarkStrconvParseFloat(b, "-123.456")
}

func BenchmarkParseFloatLiteral_Negative_Long(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("-123456.78901"))
}

func BenchmarkStrconvParseFloat_Negative_Long(b *testing.B) {
	benchmarkStrconvParseFloat(b, "-123456.78901")
}

func BenchmarkParseFloatLiteral_Decimal(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("123456"))
}

func BenchmarkStrconvParseFloat_Decimal(b *testing.B) {
	benchmarkStrconvParseFloat(b, "123456")
}

func BenchmarkParseFloatLiteral_Float_No_Decimal(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("123."))
}

func BenchmarkStrconvParseFloat_Float_No_Decimal(b *testing.B) {
	benchmarkStrconvParseFloat(b, "123.")
}

func BenchmarkParseFloatLiteral_Complex(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("-123456.798012345"))
}

func BenchmarkStrconvParseFloat_Complex(b *testing.B) {
	benchmarkStrconvParseFloat(b, "-123456.798012345")
}

func BenchmarkParseFloatLiteral_Science_Notation(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("1e6"))
}

func BenchmarkStrconvParseFloat_Science_Notation(b *testing.B) {
	benchmarkStrconvParseFloat(b, "1e6")
}

func BenchmarkParseFloatLiteral_Science_Notation_Long(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("1.23e10"))
}

func BenchmarkStrconvParseFloat_Science_Notation_Long(b *testing.B) {
	benchmarkStrconvParseFloat(b, "1.23e10")
}

func BenchmarkParseFloatLiteral_Science_Notation_Negative(b *testing.B) {
	benchmarkParseFloatLiteral(b, []byte("-1.23e10"))
}

func BenchmarkStrconvParseFloat_Science_Notation_Negative(b *testing.B) {
	benchmarkStrconvParseFloat(b, "-1.23e10")
}
