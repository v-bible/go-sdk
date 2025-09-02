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

func InjectMarkLabel(str string, marks []*biblev1.Mark, labelMap map[biblev1.MarkKind]func(mark *biblev1.Mark, chapterId string) string) string {
	resolvedMarks := ResolveMarks(marks, nil)

	slices.Reverse(resolvedMarks)

	for _, mark := range resolvedMarks {
		if labelFunc, ok := labelMap[mark.Kind]; ok {
			newString := utf8string.NewString(str)

			newMarkLabel := labelFunc(mark, mark.ChapterId)

			if int(mark.StartOffset) > newString.RuneCount() {
				str = newString.String() + newMarkLabel
			} else {
				str = newString.Slice(0, int(mark.StartOffset)) + newMarkLabel + newString.Slice(int(mark.EndOffset), newString.RuneCount())
			}
		}
	}

	return str
}

var unspecifiedMdLabel = func(mark *biblev1.Mark, chapterId string) string {
	return ""
}

var fnMdLabel = func(mark *biblev1.Mark, chapterId string) string {
	return fmt.Sprintf("[^%d-%s]", mark.SortOrder+1, chapterId)
}

var refMdLabel = func(mark *biblev1.Mark, chapterId string) string {
	return fmt.Sprintf("[^%d@-%s]", mark.SortOrder+1, chapterId)
}

var wojMdLabel = func(mark *biblev1.Mark, chapterId string) string {
	return fmt.Sprintf("<b>%s</b>", mark.Content)
}

var unspecifiedHtmlLabel = func(mark *biblev1.Mark, chapterId string) string {
	return ""
}

var fnHtmlLabel = func(mark *biblev1.Mark, chapterId string) string {
	return fmt.Sprintf(`<sup><a href="#fn-%d-%s" id="fnref-%d-%s">%d</a></sup>`, mark.SortOrder, chapterId, mark.SortOrder+1, chapterId, mark.SortOrder)
}

var refHtmlLabel = func(mark *biblev1.Mark, chapterId string) string {
	return fmt.Sprintf(`<sup><a href="#fn-%d@-%s" id="fnref-%d@-%s">%d@</a></sup>`, mark.SortOrder, chapterId, mark.SortOrder+1, chapterId, mark.SortOrder)
}

var wojHtmlLabel = func(mark *biblev1.Mark, chapterId string) string {
	return fmt.Sprintf("<b>%s</b>", mark.Content)
}

func ProcessVerseMd(verses []*biblev1.Verse, marks []*biblev1.Mark, headings []*biblev1.Heading, psalms []*biblev1.PsalmMetadata) (string, error) {
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
		// NOTE: Order is Woj -> Footnote labels -> Verse number -> Poetry ->
		// Psalms -> Headings -> Heading Footnotes -> Chapter separator ->
		// Footnote text
		verseFootnotes := lo.Filter(marks, func(m *biblev1.Mark, index int) bool {
			return m.Kind == biblev1.MarkKind_MARK_KIND_FOOTNOTE && m.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE && m.TargetId == verse.Id
		})
		verseReferences := lo.Filter(marks, func(m *biblev1.Mark, index int) bool {
			return m.Kind == biblev1.MarkKind_MARK_KIND_REFERENCE && m.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE && m.TargetId == verse.Id
		})
		verseHeadings := lo.Filter(headings, func(h *biblev1.Heading, index int) bool {
			return h.VerseId == verse.Id
		})
		verseWoj := lo.Filter(marks, func(m *biblev1.Mark, index int) bool {
			return m.Kind == biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS && m.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE && m.TargetId == verse.Id
		})

		newContent := strings.Clone(verse.Text)

		verseMarks := append([]*biblev1.Mark{}, verseFootnotes...)
		verseMarks = append(verseMarks, verseReferences...)
		verseMarks = append(verseMarks, verseWoj...)

		labelMap := map[biblev1.MarkKind]func(mark *biblev1.Mark, chapterId string) string{
			biblev1.MarkKind_MARK_KIND_UNSPECIFIED:    unspecifiedMdLabel,
			biblev1.MarkKind_MARK_KIND_FOOTNOTE:       fnMdLabel,
			biblev1.MarkKind_MARK_KIND_REFERENCE:      refMdLabel,
			biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS: wojMdLabel,
		}

		newContent = InjectMarkLabel(newContent, verseMarks, labelMap)

		// NOTE: Add verse number label only to the first verse or the first
		// verse in the paragraph
		if verse.SubVerseIndex == 0 || verse.ParagraphIndex == 0 {
			newContent = fmt.Sprintf("<sup><b>%s</b></sup>", verse.Label) + " " + newContent
		}

		if verse.IsPoetry {
			newContent = "\n> " + newContent + "\n>"
		}

		slices.Reverse(psalms)

		// NOTE: Add the Psalm title to the first verse
		if verse.SubVerseIndex == 0 && verse.ParagraphNumber == 0 {
			for _, psalm := range psalms {
				if psalm.ChapterId == verse.ChapterId {
					newContent = fmt.Sprintf("*%s*", psalm.Text) + "\n" + newContent
				}
			}
		}

		for i := range verseHeadings {
			revIdx := len(verseHeadings) - 1 - i

			headingFootnotes := lo.Filter(marks, func(fn *biblev1.Mark, index int) bool {
				return fn.Kind == biblev1.MarkKind_MARK_KIND_FOOTNOTE && fn.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_HEADING && fn.TargetId == verseHeadings[revIdx].Id
			})
			headingReferences := lo.Filter(marks, func(m *biblev1.Mark, index int) bool {
				return m.Kind == biblev1.MarkKind_MARK_KIND_REFERENCE && m.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_HEADING && m.TargetId == verseHeadings[revIdx].Id
			})

			headingMarks := append([]*biblev1.Mark{}, headingFootnotes...)
			headingMarks = append(headingMarks, headingReferences...)

			newHeadingContent := InjectMarkLabel(verseHeadings[revIdx].Text, headingMarks, labelMap)

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

	slices.SortFunc(marks, func(a, b *biblev1.Mark) int {
		return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.SortOrder, b.SortOrder))
	})

	for _, footnote := range marks {
		if footnote.Kind == biblev1.MarkKind_MARK_KIND_FOOTNOTE {
			fnSection += fmt.Sprintf("[^%d-%s]: %s", footnote.SortOrder+1, footnote.ChapterId, footnote.Content) + "\n\n"
		} else if footnote.Kind == biblev1.MarkKind_MARK_KIND_REFERENCE {
			fnSection += fmt.Sprintf("[^%d@-%s]: %s", footnote.SortOrder+1, footnote.ChapterId, footnote.Content) + "\n\n"
		}
	}

	fnLines := strings.Split(fnSection, "\n\n")

	// NOTE: Remove duplicate footnotes and references
	uniqueFnLines := lo.Uniq(fnLines)

	mdString += strings.Join(uniqueFnLines, "\n\n")

	// NOTE: Clean up the blockquote redundant characters. Note to cleanup the
	// blockquote characters, we need to replace the `>` characters at the
	// beginning of the line
	mdString = regexp.MustCompile(`(?m)^>\n+>`).ReplaceAllString(mdString, ">\n>")
	mdString = regexp.MustCompile(`(?m)^>\n\n`).ReplaceAllString(mdString, ">\n>")
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

func ProcessVerseHtml(verses []*biblev1.Verse, marks []*biblev1.Mark, headings []*biblev1.Heading, psalms []*biblev1.PsalmMetadata) (string, error) {
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
		// NOTE: Order is Woj -> Footnote labels -> Verse number -> Poetry ->
		// Psalms -> Headings -> Heading Footnotes -> Chapter separator ->
		// Footnote text
		verseFootnotes := lo.Filter(marks, func(fn *biblev1.Mark, index int) bool {
			return fn.Kind == biblev1.MarkKind_MARK_KIND_FOOTNOTE && fn.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE && fn.TargetId == verse.Id
		})
		verseReferences := lo.Filter(marks, func(m *biblev1.Mark, index int) bool {
			return m.Kind == biblev1.MarkKind_MARK_KIND_REFERENCE && m.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE && m.TargetId == verse.Id
		})
		verseHeadings := lo.Filter(headings, func(h *biblev1.Heading, index int) bool {
			return h.VerseId == verse.Id
		})
		verseWoj := lo.Filter(marks, func(w *biblev1.Mark, index int) bool {
			return w.Kind == biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS && w.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_VERSE && w.TargetId == verse.Id
		})

		newContent := strings.Clone(verse.Text)

		verseMarks := append([]*biblev1.Mark{}, verseFootnotes...)
		verseMarks = append(verseMarks, verseReferences...)
		verseMarks = append(verseMarks, verseWoj...)

		labelMap := map[biblev1.MarkKind]func(mark *biblev1.Mark, chapterId string) string{
			biblev1.MarkKind_MARK_KIND_UNSPECIFIED:    unspecifiedHtmlLabel,
			biblev1.MarkKind_MARK_KIND_FOOTNOTE:       fnHtmlLabel,
			biblev1.MarkKind_MARK_KIND_REFERENCE:      refHtmlLabel,
			biblev1.MarkKind_MARK_KIND_WORDS_OF_JESUS: wojHtmlLabel,
		}

		// NOTE: Clean up p element wrapped because it will create a new line
		newContent = InjectMarkLabel(regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(newContent), ""), verseMarks, labelMap)

		// NOTE: Add verse number label only to the first verse or the first
		// verse in the paragraph
		if verse.SubVerseIndex == 0 || verse.ParagraphIndex == 0 {
			newContent = fmt.Sprintf("<sup><b>%s</b></sup>", verse.Label) + " " + newContent
		}

		if verse.IsPoetry {
			newContent = "\n<blockquote>" + newContent + "</blockquote>\n"
		}

		slices.Reverse(psalms)

		// NOTE: Add the Psalm title to the first verse
		if verse.SubVerseIndex == 0 && verse.ParagraphNumber == 0 {
			for _, psalm := range psalms {
				if psalm.ChapterId == verse.ChapterId {
					newContent = fmt.Sprintf("<i>%s</i>", regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(psalm.Text), "")) + "\n" + newContent
				}
			}
		}

		for i := range verseHeadings {
			revIdx := len(verseHeadings) - 1 - i

			headingFootnotes := lo.Filter(marks, func(fn *biblev1.Mark, index int) bool {
				return fn.Kind == biblev1.MarkKind_MARK_KIND_FOOTNOTE && fn.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_HEADING && fn.TargetId == verseHeadings[revIdx].Id
			})
			headingReferences := lo.Filter(marks, func(m *biblev1.Mark, index int) bool {
				return m.Kind == biblev1.MarkKind_MARK_KIND_REFERENCE && m.TargetType == biblev1.MarkTargetType_MARK_TARGET_TYPE_HEADING && m.TargetId == verseHeadings[revIdx].Id
			})

			headingMarks := append([]*biblev1.Mark{}, headingFootnotes...)
			headingMarks = append(headingMarks, headingReferences...)

			newHeadingContent := InjectMarkLabel(regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(verseHeadings[revIdx].Text), ""), headingMarks, labelMap)

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

	slices.SortFunc(marks, func(a, b *biblev1.Mark) int {
		return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.SortOrder, b.SortOrder))
	})

	for _, footnote := range marks {
		if footnote.Kind == biblev1.MarkKind_MARK_KIND_FOOTNOTE {
			fnSection += fmt.Sprintf(`<li id="fn-%d-%s"><p>%s [<a href="#fnref-%d-%s">%d</a>]</p></li>`, footnote.SortOrder+1, footnote.ChapterId, regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(footnote.Content), ""), footnote.SortOrder+1, footnote.ChapterId, footnote.SortOrder+1) + "\n\n"
		} else if footnote.Kind == biblev1.MarkKind_MARK_KIND_REFERENCE {
			fnSection += fmt.Sprintf(`<li id="fn-%d@-%s"><p>%s [<a href="#fnref-%d@-%s">%d@</a>]</p></li>`, footnote.SortOrder+1, footnote.ChapterId, regexp.MustCompile(`<p>|<\/p>\n?`).ReplaceAllString(mdToHTML(footnote.Content), ""), footnote.SortOrder+1, footnote.ChapterId, footnote.SortOrder+1) + "\n\n"
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
