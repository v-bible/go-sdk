package utils

import (
	"reflect"
	"testing"

	biblev1 "github.com/v-bible/protobuf/pkg/proto/bible/v1"
)

func TestResolveMarks(t *testing.T) {
	tests := []struct {
		name     string
		marks    []*biblev1.Mark
		options  *ResolveMarksOptions
		expected []*biblev1.Mark
	}{
		{
			name: "should handle non-overlapping marks",
			marks: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "The",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 0,
					EndOffset:   3,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "quick",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 4,
					EndOffset:   9,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "3",
					Content:     "brown",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 10,
					EndOffset:   15,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
			options: nil, // Use defaults
			expected: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "The",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 0,
					EndOffset:   3,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "quick",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 4,
					EndOffset:   9,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "3",
					Content:     "brown",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 10,
					EndOffset:   15,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
		},
		{
			name: "should handle overlapping marks with overlapKeepRight=true (default)",
			marks: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "The quick brown",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG1",
					StartOffset: 0,
					EndOffset:   15,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "quick brown fox",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG2",
					StartOffset: 4,
					EndOffset:   19,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
			options: nil, // Use defaults (overlapKeepRight=true)
			expected: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "The ",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG1",
					StartOffset: 0,
					EndOffset:   4,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "1",
					Content:     "quick brown",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG1",
					StartOffset: 4,
					EndOffset:   15,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "quick brown fox",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG2",
					StartOffset: 4,
					EndOffset:   19,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
		},
		{
			name: "should handle contained marks (one inside another)",
			marks: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "The quick brown fox",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG1",
					StartOffset: 0,
					EndOffset:   19,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "quick brown",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG2",
					StartOffset: 4,
					EndOffset:   15,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
			options: nil,
			expected: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "The quick brown fox",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG1",
					StartOffset: 0,
					EndOffset:   19,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "quick brown",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG2",
					StartOffset: 4,
					EndOffset:   15,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
		},
		{
			name: "should handle footnotes where startOffset equals endOffset",
			marks: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "Beginning of verse",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 0,
					EndOffset:   18,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "Footnote content",
					Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
					Label:       "a",
					StartOffset: 10,
					EndOffset:   10, // Zero-width footnote
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "3",
					Content:     "Reference content",
					Kind:        biblev1.MarkKind_MARK_KIND_REFERENCE,
					Label:       "1",
					StartOffset: 15,
					EndOffset:   15, // Zero-width reference
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
			options: nil,
			expected: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "Beginning of verse",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 0,
					EndOffset:   18,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "Footnote content",
					Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
					Label:       "a",
					StartOffset: 10,
					EndOffset:   10,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "3",
					Content:     "Reference content",
					Kind:        biblev1.MarkKind_MARK_KIND_REFERENCE,
					Label:       "1",
					StartOffset: 15,
					EndOffset:   15,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
		},
		{
			name:     "should return empty array when no marks are provided",
			marks:    []*biblev1.Mark{},
			options:  nil,
			expected: []*biblev1.Mark{},
		},
		{
			name: "should handle single mark",
			marks: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "Single mark",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 0,
					EndOffset:   11,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
			options: nil,
			expected: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "Single mark",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 0,
					EndOffset:   11,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
		},
		{
			name: "should handle multiple footnotes at same position",
			marks: []*biblev1.Mark{
				{
					Id:          "1",
					Content:     "footnote1",
					Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
					Label:       "a",
					StartOffset: 10,
					EndOffset:   10,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "footnote2",
					Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
					Label:       "b",
					StartOffset: 10,
					EndOffset:   10,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "3",
					Content:     "Base text here",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 0,
					EndOffset:   20,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
			options: nil,
			expected: []*biblev1.Mark{
				{
					Id:          "3",
					Content:     "Base text here",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "HIG",
					StartOffset: 0,
					EndOffset:   20,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "1",
					Content:     "footnote1",
					Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
					Label:       "a",
					StartOffset: 10,
					EndOffset:   10,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
				{
					Id:          "2",
					Content:     "footnote2",
					Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
					Label:       "b",
					StartOffset: 10,
					EndOffset:   10,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveMarks(tt.marks, tt.options)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ResolveMarks() failed for %s", tt.name)
				t.Errorf("Got:  %+v", result)
				t.Errorf("Want: %+v", tt.expected)

				// Additional detailed comparison for debugging
				if len(result) != len(tt.expected) {
					t.Errorf("Length mismatch: got %d marks, want %d marks", len(result), len(tt.expected))
				}

				for i := 0; i < min(len(result), len(tt.expected)); i++ {
					if !marksEqual(result[i], tt.expected[i]) {
						t.Errorf("Mark %d differs:\nGot:  %+v\nWant: %+v", i, result[i], tt.expected[i])
					}
				}
			}
		})
	}
}

// Test edge cases and error conditions
func TestResolveMarks_EdgeCases(t *testing.T) {
	t.Run("nil marks slice", func(t *testing.T) {
		result := ResolveMarks(nil, nil)
		if !reflect.DeepEqual(result, []*biblev1.Mark{}) {
			t.Errorf("Expected empty slice result for nil input, got %+v", result)
		}
	})

	t.Run("nil options should use defaults", func(t *testing.T) {
		marks := []*biblev1.Mark{
			{
				Id:          "1",
				Content:     "test",
				Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
				StartOffset: 0,
				EndOffset:   4,
			},
		}
		result := ResolveMarks(marks, nil)
		if len(result) != 1 {
			t.Errorf("Expected 1 mark, got %d", len(result))
		}
	})

	t.Run("marks with same start and end positions", func(t *testing.T) {
		marks := []*biblev1.Mark{
			{
				Id:          "1",
				Content:     "footnote1",
				Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
				StartOffset: 10,
				EndOffset:   10,
			},
			{
				Id:          "2",
				Content:     "footnote2",
				Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
				StartOffset: 10,
				EndOffset:   10,
			},
		}
		result := ResolveMarks(marks, nil)
		// Should preserve both zero-width marks at the same position
		if len(result) != 2 {
			t.Errorf("Expected 2 marks, got %d", len(result))
		}
	})

	t.Run("complex overlapping scenario with mixed mark types", func(t *testing.T) {
		marks := []*biblev1.Mark{
			{
				Id:          "1",
				Content:     "For God so loved the world",
				Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
				Label:       "HIG1",
				StartOffset: 0,
				EndOffset:   26,
				TargetId:    "verse1",
				TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
			},
			{
				Id:          "2",
				Content:     "God so loved",
				Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
				Label:       "HIG2",
				StartOffset: 4,
				EndOffset:   16,
				TargetId:    "verse1",
				TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
			},
			{
				Id:          "3",
				Content:     "Cross reference to Genesis",
				Kind:        biblev1.MarkKind_MARK_KIND_REFERENCE,
				Label:       "1",
				StartOffset: 8,
				EndOffset:   8, // Zero-width at "loved"
				TargetId:    "verse1",
				TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
			},
			{
				Id:          "4",
				Content:     "the world that he gave",
				Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
				Label:       "HIG3",
				StartOffset: 17,
				EndOffset:   39,
				TargetId:    "verse1",
				TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
			},
			{
				Id:          "5",
				Content:     "Footnote about giving",
				Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
				Label:       "a",
				StartOffset: 35,
				EndOffset:   35, // Zero-width at "gave"
				TargetId:    "verse1",
				TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
			},
		}

		result := ResolveMarks(marks, nil)

		// Should handle multiple overlapping scenarios
		if len(result) < 5 {
			t.Errorf("Expected at least 5 marks, got %d", len(result))
		}

		// Verify all marks are sorted by start position
		for i := 1; i < len(result); i++ {
			if result[i-1].StartOffset > result[i].StartOffset {
				t.Errorf("Marks not sorted by start position: mark %d (%d) > mark %d (%d)",
					i-1, result[i-1].StartOffset, i, result[i].StartOffset)
			}
		}
	})

	t.Run("overlapping highlights with note wrapping", func(t *testing.T) {
		marks := []*biblev1.Mark{
			{
				Id:          "1",
				Content:     "In the beginning God created",
				Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
				Label:       "HIG1",
				StartOffset: 0,
				EndOffset:   28,
				TargetId:    "verse1",
				TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
			},
			{
				Id:          "2",
				Content:     "beginning God",
				Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
				Label:       "HIG2",
				StartOffset: 7,
				EndOffset:   20,
				TargetId:    "verse1",
				TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
			},
			{
				Id:          "3",
				Content:     "God footnote",
				Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
				Label:       "a",
				StartOffset: 17,
				EndOffset:   17, // Zero-width footnote at "God"
				TargetId:    "verse1",
				TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
			},
			{
				Id:          "4",
				Content:     "created the heavens",
				Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
				Label:       "HIG3",
				StartOffset: 21,
				EndOffset:   40,
				TargetId:    "verse1",
				TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
			},
		}

		result := ResolveMarks(marks, nil)

		// Should handle overlapping highlights and note wrapping
		if len(result) < 4 {
			t.Errorf("Expected at least 4 marks, got %d", len(result))
		}

		// Verify all marks are sorted by start position
		for i := 1; i < len(result); i++ {
			if result[i-1].StartOffset > result[i].StartOffset {
				t.Errorf("Marks not sorted by start position: mark %d (%d) > mark %d (%d)",
					i-1, result[i-1].StartOffset, i, result[i].StartOffset)
			}
		}
	})
}

// Helper function to create a bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to compare marks (ignoring timestamp fields for testing)
func marksEqual(a, b *biblev1.Mark) bool {
	return a.Id == b.Id &&
		a.Content == b.Content &&
		a.Kind == b.Kind &&
		a.Label == b.Label &&
		a.SortOrder == b.SortOrder &&
		a.StartOffset == b.StartOffset &&
		a.EndOffset == b.EndOffset &&
		a.TargetId == b.TargetId &&
		a.TargetType == b.TargetType &&
		a.ChapterId == b.ChapterId
}

// Helper function for min (for Go versions < 1.21)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Benchmark tests
func BenchmarkResolveMarks_NoOverlap(b *testing.B) {
	marks := []*biblev1.Mark{
		{Id: "1", Content: "mark1", Kind: biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS, StartOffset: 0, EndOffset: 5},
		{Id: "2", Content: "mark2", Kind: biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS, StartOffset: 10, EndOffset: 15},
		{Id: "3", Content: "mark3", Kind: biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS, StartOffset: 20, EndOffset: 25},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResolveMarks(marks, nil)
	}
}

func BenchmarkResolveMarks_WithOverlaps(b *testing.B) {
	marks := []*biblev1.Mark{
		{Id: "1", Content: "overlapping mark 1", Kind: biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS, StartOffset: 0, EndOffset: 15},
		{Id: "2", Content: "overlapping mark 2", Kind: biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS, StartOffset: 10, EndOffset: 25},
		{Id: "3", Content: "overlapping mark 3", Kind: biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS, StartOffset: 20, EndOffset: 35},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResolveMarks(marks, nil)
	}
}
