package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ParsedReference struct {
	BookCode   string
	ChapterNum int
	VerseRange
}

type VerseRange struct {
	From VerseInfo
	To   VerseInfo
}

type VerseInfo struct {
	Number int
	Order  []int
}

var (
	ErrFailedToParseVerseNum       = errors.New("failed to parse verse number")
	ErrFailedToNormalizeVerseQuery = errors.New("failed to normalize verse query")
	ErrMissingBookCode             = errors.New("missing book code")
)

const InvalidStringFormatError = "%error%"

var (
	// NOTE: "*" will be used as a wildcard for verse number, and parsed as -1.
	ReVerse      = regexp.MustCompile(`^(?<verseNum>\d+)(?<verseOrder>[a-z]*)$|^(?<verseWildcard>\*)$`)
	ReVerseRange = regexp.MustCompile(`(?<fromVerse>[a-zA-Z0-9*]+)(-(?<toVerse>[a-zA-Z0-9*]+))?`)
	// NOTE: Only allow "*" in the first verse range right after chapNum
	// NOTE: It's also match numbers only queries, so MAKE SURE it is
	// chapter-split first!
	ReVerseQuery = regexp.MustCompile(`(?<chapNum>\d+)(?<verseQuery>(,|:)[a-zA-Z0-9*]+(-[a-zA-Z0-9*]+)?(\s?(\.|,)\s?[a-zA-Z0-9]+(-[a-zA-Z0-9]+)?)*)?;?`)
	// NOTE: In case "John 9--12", first is it parsed as "John 9,*-*;--12,*-*;",
	// so we have to check for "*-" and "-*", the case like "9,*-5b" won't
	// likely to happen in "casual" queries, if it does happen we only pick the
	// last like "*-5b" is "5b" for "fromVerse", as well as in "5a-*" is "5a"
	// for "toVerse".
	ReChapRangeEu          = regexp.MustCompile(`(?<fromChap>\d+),(?<fromVerse>[a-zA-Z0-9*]+)(-[a-zA-Z0-9*]+)?;--(?<toChap>\d+),([a-zA-Z0-9*]+-)?(?<toVerse>[a-zA-Z0-9*]+);?`)
	ReNormalizedVerseQuery = regexp.MustCompile(`(?<chapNum>\d+)(:|,)(?<fromVerse>[a-zA-Z0-9*]+)-(?<toVerse>[a-zA-Z0-9*]+);?`)
	ReNormalizedQueryEu    = regexp.MustCompile(`^(\d+,[a-zA-Z0-9*]+-[a-zA-Z0-9*]+;?)+$`)
	ReNormalizedQueryUs    = regexp.MustCompile(`^(\d+:[a-zA-Z0-9*]+-[a-zA-Z0-9*]+;?)+$`)
	// NOTE: For multiple chapters query, like: "John 1,12", the "," is reused.
	ReMultipleChapUs = regexp.MustCompile(`^\d+(,\d+)*$`)
	ReBookCode       = regexp.MustCompile(`^(\d+\s)?[a-zA-Z0-9]+\s`)
)

func ParseStringToOrder(str string) []int {
	order := []int{}

	for i := 0; i < len(str); i++ {
		order = append(order, int(str[i]-'a'))
	}

	return order
}

func ParseVerseNum(verse string) (VerseInfo, error) {
	matches := ReVerse.FindStringSubmatch(strings.ToLower(verse))
	if matches == nil {
		return VerseInfo{}, fmt.Errorf("%w", ErrFailedToParseVerseNum)
	}

	verseNum := matches[ReVerse.SubexpIndex("verseNum")]
	verseOrder := matches[ReVerse.SubexpIndex("verseOrder")]
	verseWildcard := matches[ReVerse.SubexpIndex("verseWildcard")]

	var verseNumInt int

	if verseWildcard == "*" {
		verseNumInt = -1
	} else {
		parsedVerseNum, err := strconv.Atoi(verseNum)
		if err != nil {
			return VerseInfo{}, fmt.Errorf("%w", ErrFailedToParseVerseNum)
		}

		verseNumInt = parsedVerseNum
	}

	if verseOrder == "" {
		return VerseInfo{Number: verseNumInt, Order: []int{-1}}, nil
	}

	order := ParseStringToOrder(verseOrder)

	return VerseInfo{Number: verseNumInt, Order: order}, nil
}

func NormalizeVerseQuery(query string) string {
	return ReVerseQuery.ReplaceAllStringFunc(query, func(match string) string {
		matches := ReVerseQuery.FindStringSubmatch(match)

		chapNum := matches[ReVerseQuery.SubexpIndex("chapNum")]
		verseQuery, _ := strings.CutPrefix(matches[ReVerseQuery.SubexpIndex("verseQuery")], ",")

		_, err := strconv.Atoi(chapNum)
		if err != nil {
			return InvalidStringFormatError
		}

		verses := strings.Split(verseQuery, ".")

		newQuery := ""

		for _, verse := range verses {
			if verse == "" {
				continue
			}

			vMatches := ReVerseRange.FindStringSubmatch(verse)

			if vMatches == nil {
				return InvalidStringFormatError
			}

			fromVerse := vMatches[ReVerseRange.SubexpIndex("fromVerse")]
			toVerse := vMatches[ReVerseRange.SubexpIndex("toVerse")]

			if toVerse == "" {
				toVerse = fromVerse
			}

			newQuery += fmt.Sprintf("%s,%s-%s;", chapNum, fromVerse, toVerse)
		}

		if newQuery == "" {
			return fmt.Sprintf("%s,*-*;", chapNum)
		}

		return newQuery
	})
}

func NormalizeChapRange(query string) string {
	return ReChapRangeEu.ReplaceAllStringFunc(query, func(match string) string {
		matches := ReChapRangeEu.FindStringSubmatch(match)

		fromChap := matches[ReChapRangeEu.SubexpIndex("fromChap")]
		fromVerse := matches[ReChapRangeEu.SubexpIndex("fromVerse")]
		toChap := matches[ReChapRangeEu.SubexpIndex("toChap")]
		toVerse := matches[ReChapRangeEu.SubexpIndex("toVerse")]

		parsedFromChap, err := ParseVerseNum(fromChap)
		if err != nil {
			return InvalidStringFormatError
		}

		parsedToChap, err := ParseVerseNum(toChap)
		if err != nil {
			return InvalidStringFormatError
		}

		if parsedFromChap.Number > parsedToChap.Number {
			return InvalidStringFormatError
		}

		newQuery := fmt.Sprintf("%d,%s-*;", parsedFromChap.Number, fromVerse)

		for i := parsedFromChap.Number + 1; i < parsedToChap.Number; i++ {
			newQuery += fmt.Sprintf("%d,*-*;", i)
		}

		newQuery += fmt.Sprintf("%d,*-%s;", parsedToChap.Number, toVerse)

		return newQuery
	})
}

// Ref: https://catholic-resources.org/Bible/Biblical_References.htm
// NOTE: Make sure no bookCode here.
func NormalizeQueryEu(query string) (string, error) {
	verseQuery := strings.ReplaceAll(query, " ", "")

	// NOTE: Support using "+" to concatenate multiple verses. E.g.: "John
	// 9:1+12"
	verseQuery = strings.ReplaceAll(verseQuery, "+", ".")

	// NOTE: MUST split by chapters before parse verse query
	splitChapters := strings.Split(verseQuery, ";")

	newQuery := ""

	for _, vQuery := range splitChapters {
		newQuery += NormalizeVerseQuery(vQuery)
	}

	verseQuery = NormalizeChapRange(newQuery)

	verseQuery = strings.Trim(verseQuery, ";")

	if !ReNormalizedQueryEu.MatchString(verseQuery) {
		return "", ErrFailedToNormalizeVerseQuery
	}

	return verseQuery, nil
}

func NormalizeQueryUs(query string) (string, error) {
	verseQuery := strings.ReplaceAll(query, " ", "")

	// NOTE: Only multiple chapters pattern like: "John 9,12" rEuses ",", if
	// they want complex patterns, they have to use ";" instead.
	if ReMultipleChapUs.MatchString(verseQuery) {
		verseQuery = strings.ReplaceAll(verseQuery, ",", ";")
	}

	// NOTE: Support using "+" to concatenate multiple verses. E.g.: "John
	// 9:1+12"
	verseQuery = strings.ReplaceAll(verseQuery, "+", ",")

	verseQuery = strings.ReplaceAll(verseQuery, ",", ".")

	verseQuery = strings.ReplaceAll(verseQuery, ":", ",")

	normalizedQuery, err := NormalizeQueryEu(verseQuery)
	if err != nil {
		return "", err
	}

	// NOTE: We have to convert the chapter separator back to ":" as Us query
	verseQuery = strings.ReplaceAll(normalizedQuery, ",", ":")

	if !ReNormalizedQueryUs.MatchString(verseQuery) {
		return "", ErrFailedToNormalizeVerseQuery
	}

	return verseQuery, nil
}

func ParseBiblicalReference(query string, format string) ([]ParsedReference, error) {
	bookCode := ReBookCode.FindString(query)

	verseQuery, isBookCodeExist := strings.CutPrefix(query, bookCode)
	if !isBookCodeExist {
		return []ParsedReference{}, ErrMissingBookCode
	}

	bookCode = strings.TrimSpace(bookCode)

	verseQuery = strings.ReplaceAll(verseQuery, " ", "")

	var normalizeFunc func(string) (string, error)

	switch strings.ToLower(format) {
	case "us":
		normalizeFunc = NormalizeQueryUs
	case "eu":
		normalizeFunc = NormalizeQueryEu
	default:
		return []ParsedReference{}, ErrFailedToNormalizeVerseQuery
	}

	normalizedQuery, err := normalizeFunc(verseQuery)
	if err != nil {
		return []ParsedReference{}, err
	}

	splitQueries := strings.Split(normalizedQuery, ";")

	parsedList := []ParsedReference{}

	for _, vQuery := range splitQueries {
		matches := ReNormalizedVerseQuery.FindStringSubmatch(vQuery)
		if matches == nil {
			return []ParsedReference{}, ErrFailedToNormalizeVerseQuery
		}

		chapNum := matches[ReNormalizedVerseQuery.SubexpIndex("chapNum")]
		fromVerse := matches[ReNormalizedVerseQuery.SubexpIndex("fromVerse")]
		toVerse := matches[ReNormalizedVerseQuery.SubexpIndex("toVerse")]

		if toVerse == "" {
			toVerse = fromVerse
		}

		chapNumInt, err := strconv.Atoi(chapNum)
		if err != nil {
			return []ParsedReference{}, err
		}

		parsedFromVerse, err := ParseVerseNum(fromVerse)
		if err != nil {
			return []ParsedReference{}, err
		}

		parsedToVerse, err := ParseVerseNum(toVerse)
		if err != nil {
			return []ParsedReference{}, err
		}

		parsedList = append(parsedList, ParsedReference{
			BookCode:   bookCode,
			ChapterNum: chapNumInt,
			VerseRange: VerseRange{
				From: parsedFromVerse,
				To:   parsedToVerse,
			},
		})
	}

	return parsedList, nil
}
