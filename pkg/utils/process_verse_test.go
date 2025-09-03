package utils

import (
	"fmt"
	"testing"

	biblev1 "github.com/v-bible/protobuf/pkg/proto/bible/v1"
)

func TestInjectMarkLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		marks    []*biblev1.Mark
		expected string
	}{
		{
			name:     "no marks",
			input:    "The quick brown fox",
			marks:    []*biblev1.Mark{},
			expected: "The quick brown fox",
		},
		{
			name:  "single footnote mark",
			input: "In the beginning God created the heavens and the earth.",
			marks: []*biblev1.Mark{
				{
					Id:          "fn1",
					Content:     "God",
					Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
					Label:       "a",
					StartOffset: 17,
					EndOffset:   20,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
			expected: "In the beginning <sup>a</sup> created the heavens and the earth.",
		},
		{
			name:  "single reference mark",
			input: "For God so loved the world",
			marks: []*biblev1.Mark{
				{
					Id:          "ref1",
					Content:     "God",
					Kind:        biblev1.MarkKind_MARK_KIND_REFERENCE,
					Label:       "1",
					StartOffset: 4,
					EndOffset:   7,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
			expected: "For <sup>1</sup> so loved the world",
		},
		{
			name:  "words of Jesus mark",
			input: "Jesus said, I am the way, the truth, and the life.",
			marks: []*biblev1.Mark{
				{
					Id:          "woj1",
					Content:     "I am the way, the truth, and the life.",
					Kind:        biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS,
					Label:       "",
					StartOffset: 12,
					EndOffset:   50,
					TargetId:    "verse1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
				},
			},
			expected: "Jesus said, <span class=\"words-of-jesus\">I am the way, the truth, and the life.</span>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the labelMap as used in the actual functions
			labelMap := map[biblev1.MarkKind]func(mark *biblev1.Mark, chapterId string) string{
				biblev1.MarkKind_MARK_KIND_FOOTNOTE: func(mark *biblev1.Mark, chapterId string) string {
					return fmt.Sprintf("<sup>%s</sup>", mark.Label)
				},
				biblev1.MarkKind_MARK_KIND_REFERENCE: func(mark *biblev1.Mark, chapterId string) string {
					return fmt.Sprintf("<sup>%s</sup>", mark.Label)
				},
				biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS: func(mark *biblev1.Mark, chapterId string) string {
					return fmt.Sprintf("<span class=\"words-of-jesus\">%s</span>", mark.Content)
				},
			}

			result := InjectMarkLabel(tt.input, tt.marks, labelMap)
			if result != tt.expected {
				t.Errorf("InjectMarkLabel() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProcessVerseMd(t *testing.T) {
	tests := []struct {
		name     string
		verses   []*biblev1.Verse
		marks    []*biblev1.Mark
		headings []*biblev1.Heading
		psalms   []*biblev1.PsalmMetadata
		expected string
	}{
		{
			name: "simple verse without marks",
			verses: []*biblev1.Verse{
				{
					Id:     "GEN.1.1",
					Number: 1,
					Text:   "In the beginning God created the heavens and the earth.",
					Label:  "1",
				},
			},
			marks:    []*biblev1.Mark{},
			headings: []*biblev1.Heading{},
			psalms:   []*biblev1.PsalmMetadata{},
			expected: "<sup><b>1</b></sup> In the beginning God created the heavens and the earth.",
		},
		{
			name: "verse with footnote",
			verses: []*biblev1.Verse{
				{
					Id:     "GEN.1.1",
					Number: 1,
					Text:   "In the beginning God created the heavens and the earth.",
					Label:  "1",
				},
			},
			marks: []*biblev1.Mark{
				{
					Id:          "fn1",
					Content:     "God",
					Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
					Label:       "a",
					StartOffset: 17,
					EndOffset:   20,
					TargetId:    "GEN.1.1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
					SortOrder:   0,
					ChapterId:   "",
				},
			},
			headings: []*biblev1.Heading{},
			psalms:   []*biblev1.PsalmMetadata{},
			expected: "<sup><b>1</b></sup> In the beginning [^1-] created the heavens and the earth.\n\n[^1-]: God",
		},
		{
			name: "verse with heading",
			verses: []*biblev1.Verse{
				{
					Id:     "GEN.1.1",
					Number: 1,
					Text:   "In the beginning God created the heavens and the earth.",
					Label:  "1",
				},
			},
			marks: []*biblev1.Mark{},
			headings: []*biblev1.Heading{
				{
					Id:      "heading1",
					Text:    "The Creation of the World",
					Level:   1,
					VerseId: "GEN.1.1",
				},
			},
			psalms:   []*biblev1.PsalmMetadata{},
			expected: "# The Creation of the World\n<sup><b>1</b></sup> In the beginning God created the heavens and the earth.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessVerseMd(tt.verses, tt.marks, tt.headings, tt.psalms)
			if err != nil {
				t.Errorf("ProcessVerseMd() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ProcessVerseMd() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProcessVerseHtml(t *testing.T) {
	tests := []struct {
		name     string
		verses   []*biblev1.Verse
		marks    []*biblev1.Mark
		headings []*biblev1.Heading
		psalms   []*biblev1.PsalmMetadata
		expected string
	}{
		{
			name: "simple verse without marks",
			verses: []*biblev1.Verse{
				{
					Id:     "GEN.1.1",
					Number: 1,
					Text:   "In the beginning God created the heavens and the earth.",
					Label:  "1",
				},
			},
			marks:    []*biblev1.Mark{},
			headings: []*biblev1.Heading{},
			psalms:   []*biblev1.PsalmMetadata{},
			expected: "<sup><b>1</b></sup> In the beginning God created the heavens and the earth.<hr>\n\n<ol></ol>",
		},
		{
			name: "verse with footnote",
			verses: []*biblev1.Verse{
				{
					Id:     "GEN.1.1",
					Number: 1,
					Text:   "In the beginning God created the heavens and the earth.",
					Label:  "1",
				},
			},
			marks: []*biblev1.Mark{
				{
					Id:          "fn1",
					Content:     "God",
					Kind:        biblev1.MarkKind_MARK_KIND_FOOTNOTE,
					Label:       "a",
					StartOffset: 17,
					EndOffset:   20,
					TargetId:    "GEN.1.1",
					TargetType:  biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE,
					SortOrder:   0,
					ChapterId:   "",
				},
			},
			headings: []*biblev1.Heading{},
			psalms:   []*biblev1.PsalmMetadata{},
			expected: "<sup><b>1</b></sup> In the beginning <sup><a href=\"#fn-1-\" id=\"fnref-1-\">1</a></sup> created the heavens and the earth.<hr>\n\n<ol><li id=\"fn-1-\"><p>God [<a href=\"#fnref-1-\">1</a>]</p></li>\n\n</ol>",
		},
		{
			name: "verse with heading",
			verses: []*biblev1.Verse{
				{
					Id:     "GEN.1.1",
					Number: 1,
					Text:   "In the beginning God created the heavens and the earth.",
					Label:  "1",
				},
			},
			marks: []*biblev1.Mark{},
			headings: []*biblev1.Heading{
				{
					Id:      "heading1",
					Text:    "The Creation of the World",
					Level:   1,
					VerseId: "GEN.1.1",
				},
			},
			psalms:   []*biblev1.PsalmMetadata{},
			expected: "<h1>The Creation of the World</h1>\n<sup><b>1</b></sup> In the beginning God created the heavens and the earth.<hr>\n\n<ol></ol>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessVerseHtml(tt.verses, tt.marks, tt.headings, tt.psalms)
			if err != nil {
				t.Errorf("ProcessVerseHtml() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ProcessVerseHtml() = %q, want %q", result, tt.expected)
			}
		})
	}
}
