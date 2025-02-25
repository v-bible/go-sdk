package utils

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"golang.org/x/exp/utf8string"
)

type Book struct {
	Id        string         `json:"id"`
	Code      string         `json:"code"`
	Title     string         `json:"title"`
	Canon     string         `json:"canon"`
	Order     int            `json:"order"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Chapters  []*BookChapter `json:"chapters"`
	VersionId string         `json:"versionId"`
}

type BookChapter struct {
	Id        string    `json:"id"`
	Number    int       `json:"number"`
	Ref       string    `json:"ref"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	BookId    string    `json:"bookId"`
}

type BookVerse struct {
	Id        string    `json:"id"`
	Number    int       `json:"number"`
	Content   string    `json:"content"`
	Order     int       `json:"order"`
	ParNumber int       `json:"parNumber"`
	ParIndex  int       `json:"parIndex"`
	IsPoetry  bool      `json:"isPoetry"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ChapterId string    `json:"chapterId"`
}

type BookFootnote struct {
	Id        string    `json:"id"`
	Content   string    `json:"content"`
	Position  int       `json:"position"`
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	VerseId   *string   `json:"verseId"`
	HeadingId *string   `json:"headingId"`
	ChapterId string    `json:"chapterId"`
}

type BookHeading struct {
	Id        string    `json:"id"`
	Content   string    `json:"content"`
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	VerseId   string    `json:"verseId"`
	ChapterId string    `json:"chapterId"`
}

type BookReference struct {
	Id        string    `json:"id"`
	Content   string    `json:"content"`
	Position  *int      `json:"position"`
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	VerseId   *string   `json:"verseId"`
	HeadingId *string   `json:"headingId"`
	ChapterId string    `json:"chapterId"`
}

type PsalmMetadata struct {
	Id        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ChapterId string    `json:"chapterId"`
}

type GetAllBookParams struct {
	VersionCode *string `json:"versionCode"`
	LangCode    *string `json:"langCode"`
	WebOrigin   *string `json:"webOrigin"`
}

type GetOneBookParams struct {
	BookCode    string  `json:"bookCode"`
	VersionCode *string `json:"versionCode"`
	LangCode    *string `json:"langCode"`
	WebOrigin   *string `json:"webOrigin"`
}

type GetOneChapterParams struct {
	BookCode    string  `json:"bookCode"`
	ChapterNum  string  `json:"chapterNum"`
	VersionCode *string `json:"versionCode"`
	LangCode    *string `json:"langCode"`
	WebOrigin   *string `json:"webOrigin"`
}

func FilterByVerse[T BookFootnote | BookHeading | BookReference](verseId string, items []*T) []*T {
	var itemList []*T
	itemList = append(itemList, items...)

	itemList = slices.DeleteFunc(itemList, func(item *T) bool {
		var check bool

		switch i := any(item).(type) {
		case *BookFootnote:
			if i.VerseId == nil {
				return true
			}

			check = *i.VerseId != verseId
		case *BookHeading:
			check = i.VerseId != verseId
		case *BookReference:
			if i.VerseId == nil {
				return true
			}

			check = *i.VerseId != verseId
		}

		return check
	})

	return itemList
}

func FilterByHeading[T BookFootnote | BookReference](headingId string, items []*T) []*T {
	var itemList []*T
	itemList = append(itemList, items...)

	itemList = slices.DeleteFunc(itemList, func(item *T) bool {
		var check bool

		switch i := any(item).(type) {
		case *BookFootnote:
			if i.HeadingId == nil {
				return true
			}

			check = *i.HeadingId != headingId
		case *BookReference:
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
	Position  int
	Order     int
	Type      string
}

func processFootnoteAndRef(str string, footnotes []*BookFootnote, refs []*BookReference, fnLabel func(order int, chapterId string) string, refLabel func(order int, chapterId string) string) string {
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

		if note.Position > newString.RuneCount() {
			str = newString.String() + newRefLabel
		} else {
			str = newString.Slice(0, note.Position) + newRefLabel + newString.Slice(note.Position, newString.RuneCount())
		}
	}

	return str
}

var fnMdLabel = func(order int, chapterId string) string {
	return fmt.Sprintf("[^%d-%s]", order, chapterId)
}

var refMdLabel = func(order int, chapterId string) string {
	return fmt.Sprintf("[^%d@-%s]", order, chapterId)
}

var fnHtmlLabel = func(order int, chapterId string) string {
	return fmt.Sprintf(`<sup><a href="#fn-%d-%s" id="fnref-%d-%s">%d</a></sup>`, order, chapterId, order, chapterId, order)
}

var refHtmlLabel = func(order int, chapterId string) string {
	return fmt.Sprintf(`<sup><a href="#fn-%d@-%s" id="fnref-%d@-%s">%d@</a></sup>`, order, chapterId, order, chapterId, order)
}

func ProcessVerseMd(verses []*BookVerse, footnotes []*BookFootnote, headings []*BookHeading, refs []*BookReference, psalms []*PsalmMetadata) (string, error) {
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

			var headingFootnotes []*BookFootnote = make([]*BookFootnote, 0)

			var headingRefs []*BookReference = make([]*BookReference, 0)

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

		if verse.ParNumber > currPar {
			mdString += "\n\n" + verse.Content
		} else {
			mdString += " " + verse.Content
		}

		currPar = verse.ParNumber
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

func ProcessVerseHtml(verses []*BookVerse, footnotes []*BookFootnote, headings []*BookHeading, refs []*BookReference, psalms []*PsalmMetadata) (string, error) {
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

			var headingFootnotes []*BookFootnote = make([]*BookFootnote, 0)

			var headingRefs []*BookReference = make([]*BookReference, 0)

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

		if verse.ParNumber > currPar {
			htmlString += "\n\n" + verse.Content
		} else {
			htmlString += " " + verse.Content
		}

		currPar = verse.ParNumber
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
