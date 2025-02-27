package utils

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	biblev1 "github.com/v-bible/protobuf/pkg/proto/bible/v1"
	"golang.org/x/exp/utf8string"
)

func FilterByVerse[T biblev1.BookFootnote | biblev1.BookHeading | biblev1.BookReference](verseId string, items []*T) []*T {
	var itemList []*T
	itemList = append(itemList, items...)

	itemList = slices.DeleteFunc(itemList, func(item *T) bool {
		var check bool

		switch i := any(item).(type) {
		case *biblev1.BookFootnote:
			if i.VerseId == nil {
				return true
			}

			check = *i.VerseId != verseId
		case *biblev1.BookHeading:
			check = i.VerseId != verseId
		case *biblev1.BookReference:
			if i.VerseId == nil {
				return true
			}

			check = *i.VerseId != verseId
		}

		return check
	})

	return itemList
}

func FilterByHeading[T biblev1.BookFootnote | biblev1.BookReference](headingId string, items []*T) []*T {
	var itemList []*T
	itemList = append(itemList, items...)

	itemList = slices.DeleteFunc(itemList, func(item *T) bool {
		var check bool

		switch i := any(item).(type) {
		case *biblev1.BookFootnote:
			if i.HeadingId == nil {
				return true
			}

			check = *i.HeadingId != headingId
		case *biblev1.BookReference:
			if i.HeadingId == nil {
				return true
			}

			check = *i.HeadingId != headingId
		}

		return check
	})

	return itemList
}

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
	for _, verse := range verses {
		// NOTE: Order is 1st Heading -> Ref -> 2nd Heading -> Content ->
		// Footnote
		verseFootnotes := FilterByVerse(verse.Id, footnotes)
		verseHeadings := FilterByVerse(verse.Id, headings)
		verseRefs := FilterByVerse(verse.Id, refs)

		verse.Content = processFootnoteAndRef(verse.Content, verseFootnotes, verseRefs, fnMdLabel, refMdLabel)

		// NOTE: Add verse number only to the first verse
		if verse.Order == 0 {
			verse.Content = fmt.Sprintf("<sup><b>%d</b></sup>", verse.Number) + " " + verse.Content
		}

		if verse.IsPoetry {
			verse.Content = "\n> " + verse.Content + "\n>"
		}

		// NOTE: Add the Psalm title to the first verse
		if verse.Order == 0 && verse.ParNumber == 0 && verse.ParIndex == 0 && psalms != nil {
			for _, psalm := range psalms {
				if psalm.ChapterId == verse.ChapterId {
					verse.Content = fmt.Sprintf("*%s*", psalm.Title) + "\n" + verse.Content
				}
			}
		}

		for i := range verseHeadings {
			revIdx := len(verseHeadings) - 1 - i

			var headingFootnotes []*biblev1.BookFootnote = make([]*biblev1.BookFootnote, 0)

			var headingRefs []*biblev1.BookReference = make([]*biblev1.BookReference, 0)

			headingFootnotes = append(headingFootnotes, FilterByHeading(verseHeadings[revIdx].Id, footnotes)...)

			headingRefs = append(headingRefs, FilterByHeading(verseHeadings[revIdx].Id, refs)...)

			newHeadingContent := processFootnoteAndRef(verseHeadings[revIdx].Content, headingFootnotes, headingRefs, fnMdLabel, refMdLabel)

			// NOTE: Only first heading is h1
			if revIdx == 0 {
				verse.Content = "\n# " + newHeadingContent + "\n" + verse.Content
			} else {
				verse.Content = "\n## " + newHeadingContent + "\n" + verse.Content
			}
		}
	}

	mdString := ""
	currPar := 0
	// NOTE: Store to add newlines between chapters
	currentChapterId := ""

	for _, verse := range verses {
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
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

func ProcessVerseHtml(verses []*biblev1.BookVerse, footnotes []*biblev1.BookFootnote, headings []*biblev1.BookHeading, refs []*biblev1.BookReference, psalms []*biblev1.PsalmMetadata) (string, error) {
	for _, verse := range verses {
		// NOTE: Order is 1st Heading -> Ref -> 2nd Heading -> Content ->
		// Footnote
		verseFootnotes := FilterByVerse(verse.Id, footnotes)
		verseHeadings := FilterByVerse(verse.Id, headings)
		verseRefs := FilterByVerse(verse.Id, refs)

		verse.Content = processFootnoteAndRef(verse.Content, verseFootnotes, verseRefs, fnHtmlLabel, refHtmlLabel)

		// NOTE: Add verse number only to the first verse
		if verse.Order == 0 {
			verse.Content = fmt.Sprintf("<sup><b>%d</b></sup>", verse.Number) + " " + verse.Content
		}

		if verse.IsPoetry {
			verse.Content = "\n<blockquote>" + verse.Content + "</blockquote>\n"
		}

		// NOTE: Add the Psalm title to the first verse
		if verse.Order == 0 && verse.ParNumber == 0 && verse.ParIndex == 0 && psalms != nil {
			for _, psalm := range psalms {
				if psalm.ChapterId == verse.ChapterId {
					verse.Content = fmt.Sprintf("<i>%s</i>", psalm.Title) + "\n" + verse.Content
				}
			}
		}

		for i := range verseHeadings {
			revIdx := len(verseHeadings) - 1 - i

			var headingFootnotes []*biblev1.BookFootnote = make([]*biblev1.BookFootnote, 0)

			var headingRefs []*biblev1.BookReference = make([]*biblev1.BookReference, 0)

			headingFootnotes = append(headingFootnotes, FilterByHeading(verseHeadings[revIdx].Id, footnotes)...)

			headingRefs = append(headingRefs, FilterByHeading(verseHeadings[revIdx].Id, refs)...)

			newHeadingContent := processFootnoteAndRef(verseHeadings[revIdx].Content, headingFootnotes, headingRefs, fnHtmlLabel, refHtmlLabel)

			// NOTE: Only first heading is h1
			if revIdx == 0 {
				verse.Content = "\n<h1>" + newHeadingContent + "</h1>\n" + verse.Content
			} else {
				verse.Content = "\n<h2>" + newHeadingContent + "</h2>\n" + verse.Content
			}
		}
	}

	htmlString := ""
	currPar := 0
	// NOTE: Store to add newlines between chapters
	currentChapterId := ""

	for _, verse := range verses {
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
		htmlString += fmt.Sprintf(`<li id="fn-%d-%s"><p>%s [<a href="#fnref-%d-%s">%d</a>]</p></li>`, footnote.Order+1, footnote.ChapterId, footnote.Content, footnote.Order+1, footnote.ChapterId, footnote.Order+1) + "\n"
	}

	for _, ref := range refs {
		if ref.Position != nil {
			htmlString += fmt.Sprintf(`<li id="fn-%d@-%s"><p>%s [<a href="#fnref-%d@-%s">%d@</a>]</p></li>`, ref.Order+1, ref.ChapterId, ref.Content, ref.Order+1, ref.ChapterId, ref.Order+1) + "\n"
		}
	}

	htmlString += "</ol>"

	htmlString = strings.TrimSpace(htmlString)

	htmlString = mdToHTML(htmlString)

	return htmlString, nil
}
