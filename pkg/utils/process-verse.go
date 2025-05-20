package utils

import (
	"bytes"
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/samber/lo"
	biblev1 "github.com/v-bible/protobuf/pkg/proto/bible/v1"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"golang.org/x/exp/utf8string"
)

const MaxHeading = 6

type MapNote struct {
	ChapterId string
	Position  int32
	Order     int32
	Type      string
}

func processFootnoteAndRef(str string, footnotes []*biblev1.BookFootnote, refs []*biblev1.BookReference, fnLabel func(order int32, chapterId string) string, refLabel func(order int32, chapterId string) string) string {
	mappedNote := make([]*MapNote, 0)

	for _, footnote := range footnotes {
		mappedNote = append(mappedNote, &MapNote{
			ChapterId: footnote.ChapterId,
			Position:  footnote.Position,
			Order:     footnote.Order,
			Type:      "footnote",
		})
	}

	for _, ref := range refs {
		if ref.Position != nil {
			mappedNote = append(mappedNote, &MapNote{
				ChapterId: ref.ChapterId,
				Position:  *ref.Position,
				Order:     ref.Order,
				Type:      "reference",
			})
		}
	}

	// NOTE: Sort the footnotes and refs in descending order so when we add
	// footnote content, the position of the next footnote will not be affected
	slices.SortFunc(mappedNote, func(a, b *MapNote) int {
		return cmp.Compare(b.Position, a.Position)
	})

	for _, note := range mappedNote {
		newString := utf8string.NewString(str)
		// newRefLabel := fmt.Sprintf("[^%d-%s]", note.Order+1, note.ChapterId)
		newRefLabel := fnLabel(note.Order+1, note.ChapterId)

		if note.Type == "reference" {
			// NOTE: Must match with corresponding footnote label
			newRefLabel = refLabel(note.Order+1, note.ChapterId)
		}

		if int(note.Position) > newString.RuneCount() {
			str = newString.String() + newRefLabel
		} else {
			str = newString.Slice(0, int(note.Position)) + newRefLabel + newString.Slice(int(note.Position), newString.RuneCount())
		}
	}

	return str
}

var fnMdLabel = func(order int32, chapterId string) string {
	return fmt.Sprintf("[^%d-%s]", order, chapterId)
}

var refMdLabel = func(order int32, chapterId string) string {
	return fmt.Sprintf("[^%d@-%s]", order, chapterId)
}

var fnHtmlLabel = func(order int32, chapterId string) string {
	return fmt.Sprintf(`<sup><a href="#fn-%d-%s" id="fnref-%d-%s">%d</a></sup>`, order, chapterId, order, chapterId, order)
}

var refHtmlLabel = func(order int32, chapterId string) string {
	return fmt.Sprintf(`<sup><a href="#fn-%d@-%s" id="fnref-%d@-%s">%d@</a></sup>`, order, chapterId, order, chapterId, order)
}

func ProcessVerseMd(verses []*biblev1.BookVerse, footnotes []*biblev1.BookFootnote, headings []*biblev1.BookHeading, refs []*biblev1.BookReference, psalms []*biblev1.PsalmMetadata) (string, error) {
	newVerses := make([]*biblev1.BookVerse, len(verses))

	for i := range verses {
		newVerses[i] = &biblev1.BookVerse{
			Id:        verses[i].Id,
			Number:    verses[i].Number,
			Content:   verses[i].Content,
			Order:     verses[i].Order,
			ParNumber: verses[i].ParNumber,
			ParIndex:  verses[i].ParIndex,
			IsPoetry:  verses[i].IsPoetry,
			CreatedAt: verses[i].CreatedAt,
			UpdatedAt: verses[i].UpdatedAt,
			ChapterId: verses[i].ChapterId,
		}
	}

	for _, verse := range newVerses {
		// NOTE: Order is 1st Heading -> Ref -> 2nd Heading -> Content ->
		// Footnote
		verseFootnotes := lo.Filter(footnotes, func(fn *biblev1.BookFootnote, index int) bool {
			return fn.VerseId != nil && *fn.VerseId == verse.Id
		})
		verseHeadings := lo.Filter(headings, func(h *biblev1.BookHeading, index int) bool {
			return h.VerseId == verse.Id
		})
		verseRefs := lo.Filter(refs, func(r *biblev1.BookReference, index int) bool {
			return r.VerseId != nil && *r.VerseId == verse.Id
		})

		newContent := strings.Clone(verse.Content)

		newContent = processFootnoteAndRef(newContent, verseFootnotes, verseRefs, fnMdLabel, refMdLabel)

		// NOTE: Add verse number only to the first verse
		if verse.Order == 0 {
			newContent = fmt.Sprintf("<sup><b>%d</b></sup>", verse.Number) + " " + newContent
		}

		if verse.IsPoetry {
			newContent = "\n> " + newContent + "\n>"
		}

		// NOTE: Add the Psalm title to the first verse
		if verse.Order == 0 && verse.ParNumber == 0 && verse.ParIndex == 0 {
			for _, psalm := range psalms {
				if psalm.ChapterId == verse.ChapterId {
					newContent = fmt.Sprintf("*%s*", psalm.Title) + "\n" + newContent
				}
			}
		}

		for i := range verseHeadings {
			revIdx := len(verseHeadings) - 1 - i

			headingFootnotes := lo.Filter(footnotes, func(fn *biblev1.BookFootnote, index int) bool {
				return fn.HeadingId != nil && *fn.HeadingId == verseHeadings[revIdx].Id
			})

			headingRefs := lo.Filter(refs, func(r *biblev1.BookReference, index int) bool {
				return r.HeadingId != nil && *r.HeadingId == verseHeadings[revIdx].Id
			})

			newHeadingContent := processFootnoteAndRef(verseHeadings[revIdx].Content, headingFootnotes, headingRefs, fnMdLabel, refMdLabel)

			// NOTE: Heading level starts from 1
			newContent = fmt.Sprintf("\n%s ", strings.Repeat("#", int(verseHeadings[revIdx].Level)%MaxHeading)) + newHeadingContent + "\n" + newContent
		}

		verse.Content = newContent
	}

	mdString := ""
	currPar := 0
	// NOTE: Store to add newlines between chapters
	currentChapterId := ""

	for _, verse := range newVerses {
		// NOTE: Add line break between chapters
		if currentChapterId != "" && currentChapterId != verse.ChapterId {
			mdString += "\n\n---\n\n"
		}

		currentChapterId = verse.ChapterId

		if int(verse.ParNumber) > currPar {
			mdString += "\n\n" + verse.Content
		} else {
			mdString += " " + verse.Content
		}

		currPar = int(verse.ParNumber)
	}

	mdString += "\n\n"

	for _, footnote := range footnotes {
		mdString += fmt.Sprintf("[^%d-%s]: %s", footnote.Order+1, footnote.ChapterId, footnote.Content) + "\n"
	}

	for _, ref := range refs {
		if ref.Position != nil {
			mdString += fmt.Sprintf("[^%d@-%s]: %s", ref.Order+1, ref.ChapterId, ref.Content) + "\n"
		}
	}

	// NOTE: Clean up the blockquote redundant characters
	mdString = regexp.MustCompile(`>\n+>`).ReplaceAllString(mdString, ">\n>")
	mdString = strings.ReplaceAll(mdString, ">\n\n", "\n")
	// NOTE: Clean up the redundant newlines
	mdString = regexp.MustCompile(`\n{3,}`).ReplaceAllString(mdString, "\n\n")
	mdString = strings.TrimSpace(mdString)

	return mdString, nil
}

func mdToHTML(md string) string {
	converter := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var res bytes.Buffer
	if err := converter.Convert([]byte(md), &res); err != nil {
		return md
	}

	return res.String()
}

func ProcessVerseHtml(verses []*biblev1.BookVerse, footnotes []*biblev1.BookFootnote, headings []*biblev1.BookHeading, refs []*biblev1.BookReference, psalms []*biblev1.PsalmMetadata) (string, error) {
	newVerses := make([]*biblev1.BookVerse, len(verses))

	for i := range verses {
		newVerses[i] = &biblev1.BookVerse{
			Id:        verses[i].Id,
			Number:    verses[i].Number,
			Content:   verses[i].Content,
			Order:     verses[i].Order,
			ParNumber: verses[i].ParNumber,
			ParIndex:  verses[i].ParIndex,
			IsPoetry:  verses[i].IsPoetry,
			CreatedAt: verses[i].CreatedAt,
			UpdatedAt: verses[i].UpdatedAt,
			ChapterId: verses[i].ChapterId,
		}
	}

	for _, verse := range newVerses {
		// NOTE: Order is 1st Heading -> Ref -> 2nd Heading -> Content ->
		// Footnote
		verseFootnotes := lo.Filter(footnotes, func(fn *biblev1.BookFootnote, index int) bool {
			return fn.VerseId != nil && *fn.VerseId == verse.Id
		})
		verseHeadings := lo.Filter(headings, func(h *biblev1.BookHeading, index int) bool {
			return h.VerseId == verse.Id
		})
		verseRefs := lo.Filter(refs, func(r *biblev1.BookReference, index int) bool {
			return r.VerseId != nil && *r.VerseId == verse.Id
		})

		newContent := strings.Clone(verse.Content)

		// NOTE: Clean up p element wrapped because it will create a new line
		newContent = processFootnoteAndRef(regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(newContent), ""), verseFootnotes, verseRefs, fnHtmlLabel, refHtmlLabel)

		// NOTE: Add verse number only to the first verse
		if verse.Order == 0 {
			newContent = fmt.Sprintf("<sup><b>%d</b></sup>", verse.Number) + " " + newContent
		}

		if verse.IsPoetry {
			newContent = "\n<blockquote>" + newContent + "</blockquote>\n"
		}

		// NOTE: Add the Psalm title to the first verse
		if verse.Order == 0 && verse.ParNumber == 0 && verse.ParIndex == 0 {
			for _, psalm := range psalms {
				if psalm.ChapterId == verse.ChapterId {
					newContent = fmt.Sprintf("<i>%s</i>", regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(psalm.Title), "")) + "\n" + newContent
				}
			}
		}

		for i := range verseHeadings {
			revIdx := len(verseHeadings) - 1 - i

			headingFootnotes := lo.Filter(footnotes, func(fn *biblev1.BookFootnote, index int) bool {
				return fn.HeadingId != nil && *fn.HeadingId == verseHeadings[revIdx].Id
			})

			headingRefs := lo.Filter(refs, func(r *biblev1.BookReference, index int) bool {
				return r.HeadingId != nil && *r.HeadingId == verseHeadings[revIdx].Id
			})

			newHeadingContent := processFootnoteAndRef(regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(verseHeadings[revIdx].Content), ""), headingFootnotes, headingRefs, fnHtmlLabel, refHtmlLabel)

			// NOTE: Heading level starts from 1
			newContent = fmt.Sprintf("\n<h%d>", verseHeadings[revIdx].Level%MaxHeading) + newHeadingContent + fmt.Sprintf("</h%d>\n", verseHeadings[revIdx].Level%MaxHeading) + newContent
		}

		verse.Content = newContent
	}

	htmlString := ""
	currPar := 0
	// NOTE: Store to add newlines between chapters
	currentChapterId := ""

	for _, verse := range newVerses {
		if currentChapterId != "" && currentChapterId != verse.ChapterId {
			htmlString += "\n\n<hr>\n\n"
		}

		currentChapterId = verse.ChapterId

		if int(verse.ParNumber) > currPar {
			htmlString += "\n\n" + verse.Content
		} else {
			htmlString += " " + verse.Content
		}

		currPar = int(verse.ParNumber)
	}

	htmlString += "<hr>\n\n<ol>"

	for _, footnote := range footnotes {
		htmlString += fmt.Sprintf(`<li id="fn-%d-%s"><p>%s [<a href="#fnref-%d-%s">%d</a>]</p></li>`, footnote.Order+1, footnote.ChapterId, regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(footnote.Content), ""), footnote.Order+1, footnote.ChapterId, footnote.Order+1) + "\n"
	}

	for _, ref := range refs {
		if ref.Position != nil {
			htmlString += fmt.Sprintf(`<li id="fn-%d@-%s"><p>%s [<a href="#fnref-%d@-%s">%d@</a>]</p></li>`, ref.Order+1, ref.ChapterId, regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(ref.Content), ""), ref.Order+1, ref.ChapterId, ref.Order+1) + "\n"
		}
	}

	htmlString += "</ol>"

	// NOTE: Should I clean up all "\n"?
	htmlString = strings.ReplaceAll(htmlString, "\n</blockquote>", "</blockquote>")
	// NOTE: Clean up the redundant newlines
	htmlString = regexp.MustCompile(`\n{3,}`).ReplaceAllString(htmlString, "\n\n")
	htmlString = strings.TrimSpace(htmlString)

	return htmlString, nil
}
