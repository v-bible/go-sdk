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

func processFootnoteAndRef(str string, footnotes []*biblev1.Footnote, fnLabel func(order int32, chapterId string) string, refLabel func(order int32, chapterId string) string) string {
	// NOTE: Sort the footnotes and refs in descending order so when we add
	// footnote content, the position of the next footnote will not be affected
	slices.SortFunc(footnotes, func(a, b *biblev1.Footnote) int {
		return cmp.Compare(b.Position, a.Position)
	})

	for _, note := range footnotes {
		newString := utf8string.NewString(str)
		newRefLabel := fnLabel(note.SortOrder+1, note.ChapterId)

		if note.Type == "reference" {
			// NOTE: Must match with corresponding footnote label
			newRefLabel = refLabel(note.SortOrder+1, note.ChapterId)
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

func ProcessVerseMd(verses []*biblev1.Verse, footnotes []*biblev1.Footnote, headings []*biblev1.Heading, psalms []*biblev1.PsalmMetadata) (string, error) {
	newVerses := make([]*biblev1.Verse, len(verses))

	for i := range verses {
		newVerses[i] = &biblev1.Verse{
			Id:              verses[i].Id,
			Text:            verses[i].Text,
			Label:           verses[i].Label,
			Number:          verses[i].Number,
			SubVerseIndex:   verses[i].SubVerseIndex,
			ParagraphNumber: verses[i].ParagraphNumber,
			IsPoetry:        verses[i].IsPoetry,
			AudioUrl:        verses[i].AudioUrl,
			CreatedAt:       verses[i].CreatedAt,
			UpdatedAt:       verses[i].UpdatedAt,
			ChapterId:       verses[i].ChapterId,
		}
	}

	for _, verse := range newVerses {
		// NOTE: Order is 1st Heading -> Ref -> 2nd Heading -> Content ->
		// Footnote
		verseFootnotes := lo.Filter(footnotes, func(fn *biblev1.Footnote, index int) bool {
			return fn.VerseId != nil && *fn.VerseId == verse.Id
		})
		verseHeadings := lo.Filter(headings, func(h *biblev1.Heading, index int) bool {
			return h.VerseId == verse.Id
		})

		newContent := strings.Clone(verse.Text)

		newContent = processFootnoteAndRef(newContent, verseFootnotes, fnMdLabel, refMdLabel)

		// NOTE: Add verse number label only to the first verse or the first
		// verse in the paragraph
		if verse.SubVerseIndex == 0 || verse.ParagraphIndex == 0 {
			newContent = fmt.Sprintf("<sup><b>%s</b></sup>", verse.Label) + " " + newContent
		}

		if verse.IsPoetry {
			newContent = "\n> " + newContent + "\n>"
		}

		// NOTE: Add the Psalm title to the first verse
		if verse.SubVerseIndex == 0 && verse.ParagraphNumber == 0 {
			for _, psalm := range psalms {
				if psalm.ChapterId == verse.ChapterId {
					newContent = fmt.Sprintf("*%s*", psalm.Title) + "\n" + newContent
				}
			}
		}

		for i := range verseHeadings {
			revIdx := len(verseHeadings) - 1 - i

			headingFootnotes := lo.Filter(footnotes, func(fn *biblev1.Footnote, index int) bool {
				return fn.HeadingId != nil && *fn.HeadingId == verseHeadings[revIdx].Id
			})

			newHeadingContent := processFootnoteAndRef(verseHeadings[revIdx].Text, headingFootnotes, fnMdLabel, refMdLabel)

			// NOTE: Heading level starts from 1
			newContent = fmt.Sprintf("\n%s ", strings.Repeat("#", int(verseHeadings[revIdx].Level)%MaxHeading)) + newHeadingContent + "\n" + newContent
		}

		verse.Text = newContent
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

		if int(verse.ParagraphNumber) > currPar {
			mdString += "\n\n" + verse.Text
		} else {
			mdString += " " + verse.Text
		}

		currPar = int(verse.ParagraphNumber)
	}

	mdString += "\n\n"

	fnSection := ""

	slices.SortFunc(footnotes, func(a, b *biblev1.Footnote) int {
		return cmp.Or(cmp.Compare(a.Type, b.Type), cmp.Compare(a.SortOrder, b.SortOrder))
	})

	for _, footnote := range footnotes {
		if footnote.Type == "footnote" {
			fnSection += fmt.Sprintf("[^%d-%s]: %s", footnote.SortOrder+1, footnote.ChapterId, footnote.Text) + "\n\n"
		} else {
			fnSection += fmt.Sprintf("[^%d@-%s]: %s", footnote.SortOrder+1, footnote.ChapterId, footnote.Text) + "\n\n"
		}
	}

	fnLines := strings.Split(fnSection, "\n\n")

	// NOTE: Remove duplicate footnotes and references
	uniqueFnLines := lo.Uniq(fnLines)

	mdString += strings.Join(uniqueFnLines, "\n\n")

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

func ProcessVerseHtml(verses []*biblev1.Verse, footnotes []*biblev1.Footnote, headings []*biblev1.Heading, psalms []*biblev1.PsalmMetadata) (string, error) {
	newVerses := make([]*biblev1.Verse, len(verses))

	for i := range verses {
		newVerses[i] = &biblev1.Verse{
			Id:              verses[i].Id,
			Text:            verses[i].Text,
			Label:           verses[i].Label,
			Number:          verses[i].Number,
			SubVerseIndex:   verses[i].SubVerseIndex,
			ParagraphNumber: verses[i].ParagraphNumber,
			IsPoetry:        verses[i].IsPoetry,
			AudioUrl:        verses[i].AudioUrl,
			CreatedAt:       verses[i].CreatedAt,
			UpdatedAt:       verses[i].UpdatedAt,
			ChapterId:       verses[i].ChapterId,
		}
	}

	for _, verse := range newVerses {
		// NOTE: Order is 1st Heading -> Ref -> 2nd Heading -> Content ->
		// Footnote
		verseFootnotes := lo.Filter(footnotes, func(fn *biblev1.Footnote, index int) bool {
			return fn.VerseId != nil && *fn.VerseId == verse.Id
		})
		verseHeadings := lo.Filter(headings, func(h *biblev1.Heading, index int) bool {
			return h.VerseId == verse.Id
		})

		newContent := strings.Clone(verse.Text)

		// NOTE: Clean up p element wrapped because it will create a new line
		newContent = processFootnoteAndRef(regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(newContent), ""), verseFootnotes, fnHtmlLabel, refHtmlLabel)

		// NOTE: Add verse number label only to the first verse
		if verse.SubVerseIndex == 0 {
			newContent = fmt.Sprintf("<sup><b>%s</b></sup>", verse.Label) + " " + newContent
		}

		if verse.IsPoetry {
			newContent = "\n<blockquote>" + newContent + "</blockquote>\n"
		}

		// NOTE: Add the Psalm title to the first verse
		if verse.SubVerseIndex == 0 && verse.ParagraphNumber == 0 {
			for _, psalm := range psalms {
				if psalm.ChapterId == verse.ChapterId {
					newContent = fmt.Sprintf("<i>%s</i>", regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(psalm.Title), "")) + "\n" + newContent
				}
			}
		}

		for i := range verseHeadings {
			revIdx := len(verseHeadings) - 1 - i

			headingFootnotes := lo.Filter(footnotes, func(fn *biblev1.Footnote, index int) bool {
				return fn.HeadingId != nil && *fn.HeadingId == verseHeadings[revIdx].Id
			})

			newHeadingContent := processFootnoteAndRef(regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(verseHeadings[revIdx].Text), ""), headingFootnotes, fnHtmlLabel, refHtmlLabel)

			// NOTE: Heading level starts from 1
			newContent = fmt.Sprintf("\n<h%d>", verseHeadings[revIdx].Level%MaxHeading) + newHeadingContent + fmt.Sprintf("</h%d>\n", verseHeadings[revIdx].Level%MaxHeading) + newContent
		}

		verse.Text = newContent
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

		if int(verse.ParagraphNumber) > currPar {
			htmlString += "\n\n" + verse.Text
		} else {
			htmlString += " " + verse.Text
		}

		currPar = int(verse.ParagraphNumber)
	}

	htmlString += "<hr>\n\n<ol>"

	fnSection := ""

	slices.SortFunc(footnotes, func(a, b *biblev1.Footnote) int {
		return cmp.Or(cmp.Compare(a.Type, b.Type), cmp.Compare(a.SortOrder, b.SortOrder))
	})

	for _, footnote := range footnotes {
		if footnote.Type == "footnote" {
			fnSection += fmt.Sprintf(`<li id="fn-%d-%s"><p>%s [<a href="#fnref-%d-%s">%d</a>]</p></li>`, footnote.SortOrder+1, footnote.ChapterId, regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(footnote.Text), ""), footnote.SortOrder+1, footnote.ChapterId, footnote.SortOrder+1) + "\n\n"
		} else {
			fnSection += fmt.Sprintf(`<li id="fn-%d@-%s"><p>%s [<a href="#fnref-%d@-%s">%d@</a>]</p></li>`, footnote.SortOrder+1, footnote.ChapterId, regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(footnote.Text), ""), footnote.SortOrder+1, footnote.ChapterId, footnote.SortOrder+1) + "\n\n"
		}
	}

	fnLines := strings.Split(fnSection, "\n\n")

	// NOTE: Remove duplicate footnotes and references
	uniqueFnLines := lo.Uniq(fnLines)

	htmlString += strings.Join(uniqueFnLines, "\n\n")

	htmlString += "</ol>"

	// NOTE: Should I clean up all "\n"?
	htmlString = strings.ReplaceAll(htmlString, "\n</blockquote>", "</blockquote>")
	// NOTE: Clean up the redundant newlines
	htmlString = regexp.MustCompile(`\n{3,}`).ReplaceAllString(htmlString, "\n\n")
	htmlString = strings.TrimSpace(htmlString)

	return htmlString, nil
}
