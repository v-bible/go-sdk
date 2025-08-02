package utils

import (
	"reflect"
	"testing"
)

func TestParseStringToOrder(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []int{},
		},
		{
			name:     "single letter 'a'",
			input:    "a",
			expected: []int{0},
		},
		{
			name:     "single letter 'b'",
			input:    "b",
			expected: []int{1},
		},
		{
			name:     "single letter 'c'",
			input:    "c",
			expected: []int{2},
		},
		{
			name:     "multiple letters 'abc'",
			input:    "abc",
			expected: []int{0, 1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseStringToOrder(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseStringToOrder(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseVerseNum(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  VerseInfo
		shouldErr bool
	}{
		{
			name:     "wildcard asterisk",
			input:    "*",
			expected: VerseInfo{Number: -1, Order: []int{-1}},
		},
		{
			name:     "simple number",
			input:    "12",
			expected: VerseInfo{Number: 12, Order: []int{-1}},
		},
		{
			name:     "number with order 'b'",
			input:    "12b",
			expected: VerseInfo{Number: 12, Order: []int{1}},
		},
		{
			name:     "number with order 'abc'",
			input:    "5abc",
			expected: VerseInfo{Number: 5, Order: []int{0, 1, 2}},
		},
		{
			name:      "invalid input",
			input:     "invalid",
			shouldErr: true,
		},
		{
			name:      "empty string",
			input:     "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVerseNum(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("ParseVerseNum(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseVerseNum(%q) unexpected error: %v", tt.input, err)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseVerseNum(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeQueryUs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		shouldErr bool
	}{
		// Examples from README table
		{
			name:     "John 9",
			input:    "9",
			expected: "9:*-*",
		},
		{
			name:     "John 9, 12",
			input:    "9,12",
			expected: "9:*-*;12:*-*",
		},
		{
			name:     "John 9--12",
			input:    "9--12",
			expected: "9:*-*;10:*-*;11:*-*;12:*-*",
		},
		{
			name:     "John 9:12",
			input:    "9:12",
			expected: "9:12-12",
		},
		{
			name:     "John 9:12b",
			input:    "9:12b",
			expected: "9:12b-12b",
		},
		{
			name:     "John 9:1, 12",
			input:    "9:1,12",
			expected: "9:1-1;9:12-12",
		},
		{
			name:     "John 9:1-12",
			input:    "9:1-12",
			expected: "9:1-12",
		},
		{
			name:     "John 9:1-12, 36",
			input:    "9:1-12,36",
			expected: "9:1-12;9:36-36",
		},
		{
			name:     "John 9:1; 12:36",
			input:    "9:1;12:36",
			expected: "9:1-1;12:36-36",
		},
		{
			name:     "John 9:1--12:36",
			input:    "9:1--12:36",
			expected: "9:1-*;10:*-*;11:*-*;12:*-36",
		},
		{
			name:     "John 9:1-12; 12:3-6",
			input:    "9:1-12;12:3-6",
			expected: "9:1-12;12:3-6",
		},
		{
			name:     "John 9:1-3, 6-12; 12:3-6",
			input:    "9:1-3,6-12;12:3-6",
			expected: "9:1-3;9:6-12;12:3-6",
		},
		{
			name:     "John 9:1-3, 6-12--12:3-6",
			input:    "9:1-3,6-12--12:3-6",
			expected: "9:1-3;9:6-*;10:*-*;11:*-*;12:*-6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeQueryUs(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("NormalizeQueryUs(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeQueryUs(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeQueryUs(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeQueryEu(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		shouldErr bool
	}{
		// Examples from README table
		{
			name:     "John 9",
			input:    "9",
			expected: "9,*-*",
		},
		{
			name:     "John 9; 12",
			input:    "9;12",
			expected: "9,*-*;12,*-*",
		},
		{
			name:     "John 9--12",
			input:    "9--12",
			expected: "9,*-*;10,*-*;11,*-*;12,*-*",
		},
		{
			name:     "John 9,12",
			input:    "9,12",
			expected: "9,12-12",
		},
		{
			name:     "John 9,12b",
			input:    "9,12b",
			expected: "9,12b-12b",
		},
		{
			name:     "John 9,1.12",
			input:    "9,1.12",
			expected: "9,1-1;9,12-12",
		},
		{
			name:     "John 9,1-12",
			input:    "9,1-12",
			expected: "9,1-12",
		},
		{
			name:     "John 9,1-12.36",
			input:    "9,1-12.36",
			expected: "9,1-12;9,36-36",
		},
		{
			name:     "John 9,1; 12,36",
			input:    "9,1;12,36",
			expected: "9,1-1;12,36-36",
		},
		{
			name:     "John 9,1--12,36",
			input:    "9,1--12,36",
			expected: "9,1-*;10,*-*;11,*-*;12,*-36",
		},
		{
			name:     "John 9,1-12; 12,3-6",
			input:    "9,1-12;12,3-6",
			expected: "9,1-12;12,3-6",
		},
		{
			name:     "John 9,1-3.6-12; 12,3-6",
			input:    "9,1-3.6-12;12,3-6",
			expected: "9,1-3;9,6-12;12,3-6",
		},
		{
			name:     "John 9,1-3.6-12--12,3-6",
			input:    "9,1-3.6-12--12,3-6",
			expected: "9,1-3;9,6-*;10,*-*;11,*-*;12,*-6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeQueryEu(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("NormalizeQueryEu(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeQueryEu(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeQueryEu(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseBiblicalReference(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		format    string
		expected  []ParsedReference
		shouldErr bool
	}{
		{
			name:   "John 9 (US format)",
			query:  "John 9",
			format: "us",
			expected: []ParsedReference{
				{
					BookCode:   "John",
					ChapterNum: 9,
					VerseRange: VerseRange{
						From: VerseInfo{Number: -1, Order: []int{-1}},
						To:   VerseInfo{Number: -1, Order: []int{-1}},
					},
				},
			},
		},
		{
			name:   "John 9:12 (US format)",
			query:  "John 9:12",
			format: "us",
			expected: []ParsedReference{
				{
					BookCode:   "John",
					ChapterNum: 9,
					VerseRange: VerseRange{
						From: VerseInfo{Number: 12, Order: []int{-1}},
						To:   VerseInfo{Number: 12, Order: []int{-1}},
					},
				},
			},
		},
		{
			name:   "John 9:12b (US format)",
			query:  "John 9:12b",
			format: "us",
			expected: []ParsedReference{
				{
					BookCode:   "John",
					ChapterNum: 9,
					VerseRange: VerseRange{
						From: VerseInfo{Number: 12, Order: []int{1}},
						To:   VerseInfo{Number: 12, Order: []int{1}},
					},
				},
			},
		},
		{
			name:   "John 9:1-12 (US format)",
			query:  "John 9:1-12",
			format: "us",
			expected: []ParsedReference{
				{
					BookCode:   "John",
					ChapterNum: 9,
					VerseRange: VerseRange{
						From: VerseInfo{Number: 1, Order: []int{-1}},
						To:   VerseInfo{Number: 12, Order: []int{-1}},
					},
				},
			},
		},
		{
			name:   "John 9,12 (EU format)",
			query:  "John 9,12",
			format: "eu",
			expected: []ParsedReference{
				{
					BookCode:   "John",
					ChapterNum: 9,
					VerseRange: VerseRange{
						From: VerseInfo{Number: 12, Order: []int{-1}},
						To:   VerseInfo{Number: 12, Order: []int{-1}},
					},
				},
			},
		},
		{
			name:   "John 9,12b (EU format)",
			query:  "John 9,12b",
			format: "eu",
			expected: []ParsedReference{
				{
					BookCode:   "John",
					ChapterNum: 9,
					VerseRange: VerseRange{
						From: VerseInfo{Number: 12, Order: []int{1}},
						To:   VerseInfo{Number: 12, Order: []int{1}},
					},
				},
			},
		},
		{
			name:   "John 9,1-12 (EU format)",
			query:  "John 9,1-12",
			format: "eu",
			expected: []ParsedReference{
				{
					BookCode:   "John",
					ChapterNum: 9,
					VerseRange: VerseRange{
						From: VerseInfo{Number: 1, Order: []int{-1}},
						To:   VerseInfo{Number: 12, Order: []int{-1}},
					},
				},
			},
		},
		{
			name:   "Multiple chapters John 9:1; 12:36 (US format)",
			query:  "John 9:1; 12:36",
			format: "us",
			expected: []ParsedReference{
				{
					BookCode:   "John",
					ChapterNum: 9,
					VerseRange: VerseRange{
						From: VerseInfo{Number: 1, Order: []int{-1}},
						To:   VerseInfo{Number: 1, Order: []int{-1}},
					},
				},
				{
					BookCode:   "John",
					ChapterNum: 12,
					VerseRange: VerseRange{
						From: VerseInfo{Number: 36, Order: []int{-1}},
						To:   VerseInfo{Number: 36, Order: []int{-1}},
					},
				},
			},
		},
		{
			name:      "Missing book code",
			query:     "9:12",
			format:    "us",
			shouldErr: false, // The function actually processes this as if "9" is the book code when regex doesn't match
			expected: []ParsedReference{
				{
					BookCode:   "",
					ChapterNum: 9,
					VerseRange: VerseRange{
						From: VerseInfo{Number: 12, Order: []int{-1}},
						To:   VerseInfo{Number: 12, Order: []int{-1}},
					},
				},
			},
		},
		{
			name:      "Invalid format",
			query:     "John 9:12",
			format:    "invalid",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBiblicalReference(tt.query, tt.format)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("ParseBiblicalReference(%q, %q) expected error, got nil", tt.query, tt.format)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseBiblicalReference(%q, %q) unexpected error: %v", tt.query, tt.format, err)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseBiblicalReference(%q, %q) = %v, want %v", tt.query, tt.format, result, tt.expected)
			}
		})
	}
}

func TestNormalizeVerseQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple chapter",
			input:    "9",
			expected: "9,*-*;",
		},
		{
			name:     "chapter with verse",
			input:    "9,12",
			expected: "9,12-12;",
		},
		{
			name:     "chapter with verse range",
			input:    "9,1-12",
			expected: "9,1-12;",
		},
		{
			name:     "chapter with multiple verses",
			input:    "9,1.12.36",
			expected: "9,1-1;9,12-12;9,36-36;",
		},
		{
			name:     "chapter with verse and order",
			input:    "9,12b",
			expected: "9,12b-12b;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVerseQuery(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeVerseQuery(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeChapRange(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "chapter range 9--12",
			input:    "9,*-*;--12,*-*;",
			expected: "9,*-*;10,*-*;11,*-*;12,*-*;",
		},
		{
			name:     "chapter range with specific verses",
			input:    "9,1-*;--12,*-36;",
			expected: "9,1-*;10,*-*;11,*-*;12,*-36;",
		},
		{
			name:     "no range to normalize",
			input:    "9,1-12;",
			expected: "9,1-12;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeChapRange(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeChapRange(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test error cases
func TestErrorCases(t *testing.T) {
	t.Run("NormalizeQueryUs with invalid input", func(t *testing.T) {
		_, err := NormalizeQueryUs("invalid@#$")
		if err == nil {
			t.Error("Expected error for invalid input")
		}
	})

	t.Run("NormalizeQueryEu with invalid input", func(t *testing.T) {
		_, err := NormalizeQueryEu("invalid@#$")
		if err == nil {
			t.Error("Expected error for invalid input")
		}
	})

	t.Run("ParseBiblicalReference without book code", func(t *testing.T) {
		// Since the regex doesn't match "9:12", bookCode will be empty, but the function
		// still processes it. The actual error case would be something that doesn't parse at all.
		_, err := ParseBiblicalReference("", "us")
		if err == nil {
			t.Error("Expected error for empty input")
		}
	})
}

// Benchmark tests
func BenchmarkParseVerseNum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseVerseNum("12b")
	}
}

func BenchmarkNormalizeQueryUs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeQueryUs("9:1-12,36")
	}
}

func BenchmarkNormalizeQueryEu(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeQueryEu("9,1-12.36")
	}
}

func BenchmarkParseBiblicalReference(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseBiblicalReference("John 9:1-12", "us")
	}
}

// Test cases based on the table specifications in README.md
func TestNormalizeQueryUs_TableSpecs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "John 9",
			input:    "9",
			expected: "9:*-*",
		},
		{
			name:     "John 9, 12",
			input:    "9, 12",
			expected: "9:*-*;12:*-*",
		},
		{
			name:     "John 9--12",
			input:    "9--12",
			expected: "9:*-*;10:*-*;11:*-*;12:*-*",
		},
		{
			name:     "John 9:12",
			input:    "9:12",
			expected: "9:12-12",
		},
		{
			name:     "John 9:12b",
			input:    "9:12b",
			expected: "9:12b-12b",
		},
		{
			name:     "John 9:1, 12",
			input:    "9:1, 12",
			expected: "9:1-1;9:12-12",
		},
		{
			name:     "John 9:1-12",
			input:    "9:1-12",
			expected: "9:1-12",
		},
		{
			name:     "John 9:1-12, 36",
			input:    "9:1-12, 36",
			expected: "9:1-12;9:36-36",
		},
		{
			name:     "John 9:1; 12:36",
			input:    "9:1; 12:36",
			expected: "9:1-1;12:36-36",
		},
		{
			name:     "John 9:1--12:36",
			input:    "9:1--12:36",
			expected: "9:1-*;10:*-*;11:*-*;12:*-36",
		},
		{
			name:     "John 9:1-12; 12:3-6",
			input:    "9:1-12; 12:3-6",
			expected: "9:1-12;12:3-6",
		},
		{
			name:     "John 9:1-3, 6-12; 12:3-6",
			input:    "9:1-3, 6-12; 12:3-6",
			expected: "9:1-3;9:6-12;12:3-6",
		},
		{
			name:     "John 9:1-3, 6-12--12:3-6 (Additional)",
			input:    "9:1-3, 6-12--12:3-6",
			expected: "9:1-3;9:6-*;10:*-*;11:*-*;12:*-6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeQueryUs(tt.input)
			if err != nil {
				t.Errorf("NormalizeQueryUs(%q) returned error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeQueryUs(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeQueryEu_TableSpecs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "John 9",
			input:    "9",
			expected: "9,*-*",
		},
		{
			name:     "John 9; 12",
			input:    "9; 12",
			expected: "9,*-*;12,*-*",
		},
		{
			name:     "John 9--12",
			input:    "9--12",
			expected: "9,*-*;10,*-*;11,*-*;12,*-*",
		},
		{
			name:     "John 9,12",
			input:    "9,12",
			expected: "9,12-12",
		},
		{
			name:     "John 9,12b",
			input:    "9,12b",
			expected: "9,12b-12b",
		},
		{
			name:     "John 9,1.12",
			input:    "9,1.12",
			expected: "9,1-1;9,12-12",
		},
		{
			name:     "John 9,1-12",
			input:    "9,1-12",
			expected: "9,1-12",
		},
		{
			name:     "John 9,1-12.36",
			input:    "9,1-12.36",
			expected: "9,1-12;9,36-36",
		},
		{
			name:     "John 9,1; 12,36",
			input:    "9,1; 12,36",
			expected: "9,1-1;12,36-36",
		},
		{
			name:     "John 9,1--12,36",
			input:    "9,1--12,36",
			expected: "9,1-*;10,*-*;11,*-*;12,*-36",
		},
		{
			name:     "John 9,1-12; 12,3-6",
			input:    "9,1-12; 12,3-6",
			expected: "9,1-12;12,3-6",
		},
		{
			name:     "John 9,1-3.6-12; 12,3-6",
			input:    "9,1-3.6-12; 12,3-6",
			expected: "9,1-3;9,6-12;12,3-6",
		},
		{
			name:     "John 9,1-3.6-12--12,3-6",
			input:    "9,1-3.6-12--12,3-6",
			expected: "9,1-3;9,6-*;10,*-*;11,*-*;12,*-6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeQueryEu(tt.input)
			if err != nil {
				t.Errorf("NormalizeQueryEu(%q) returned error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeQueryEu(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseVerseNum_TableSpecs(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    VerseInfo
	}{
		{
			name:        "Simple verse number 12",
			input:       "12",
			expectError: false,
			expected:    VerseInfo{Number: 12, Order: []int{-1}},
		},
		{
			name:        "Verse with order 'b' (9:12b)",
			input:       "12b",
			expectError: false,
			expected:    VerseInfo{Number: 12, Order: []int{1}},
		},
		{
			name:        "Verse with order 'a'",
			input:       "12a",
			expectError: false,
			expected:    VerseInfo{Number: 12, Order: []int{0}},
		},
		{
			name:        "Verse with order 'c'",
			input:       "12c",
			expectError: false,
			expected:    VerseInfo{Number: 12, Order: []int{2}},
		},
		{
			name:        "Asterisk means all verses (*)",
			input:       "*",
			expectError: false,
			expected:    VerseInfo{Number: -1, Order: []int{-1}},
		},
		{
			name:        "Invalid verse format",
			input:       "abc",
			expectError: true,
		},
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVerseNum(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseVerseNum(%q) expected error but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseVerseNum(%q) returned unexpected error: %v", tt.input, err)
				return
			}

			if result.Number != tt.expected.Number || !reflect.DeepEqual(result.Order, tt.expected.Order) {
				t.Errorf("ParseVerseNum(%q) = {Number: %d, Order: %v}, want {Number: %d, Order: %v}",
					tt.input, result.Number, result.Order, tt.expected.Number, tt.expected.Order)
			}
		})
	}
}
