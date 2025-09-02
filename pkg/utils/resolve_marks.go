package utils

import (
	"cmp"
	"slices"

	biblev1 "github.com/v-bible/protobuf/pkg/proto/bible/v1"
	"golang.org/x/exp/utf8string"
)

type ResolveMarksOptions struct {
	overlapKeepRight *bool
}

func ResolveMarks(marks []*biblev1.Mark, options *ResolveMarksOptions) []*biblev1.Mark {
	defaultOverlapKeepRight := true

	if options == nil {
		options = &ResolveMarksOptions{
			overlapKeepRight: &defaultOverlapKeepRight,
		}
	}

	overlapKeepRight := options.overlapKeepRight

	if len(marks) == 0 {
		return make([]*biblev1.Mark, 0)
	}

	if len(marks) == 1 {
		return marks
	}

	// Sort annotations by start position (descending) for processing
	sortedMarks := slices.SortedFunc(slices.Values(marks), func(a, b *biblev1.Mark) int {
		return cmp.Compare(b.StartOffset, a.StartOffset)
	})

	additionalMarks := make([]*biblev1.Mark, 0)

	for i := 1; i < len(sortedMarks); i += 1 {
		currMark := sortedMarks[i]
		prevMark := sortedMarks[i-1]

		if prevMark.StartOffset >= currMark.EndOffset {
			// No overlap, continue to next
		} else if prevMark.StartOffset >= currMark.StartOffset && prevMark.EndOffset <= currMark.EndOffset && prevMark.StartOffset < currMark.EndOffset {
			// Contained case, continue to next
		} else if prevMark.StartOffset < currMark.EndOffset && prevMark.EndOffset > currMark.StartOffset {
			// NOTE: Overlapping case
			// NOTE: If overlapKeepRight is true then we keep "prev" (as right reversed)
			// annotation, update end of the "current" annotation to the start of
			// the "prev" annotation, and push the new inner annotation of "current"
			// annotation to the array

			if *overlapKeepRight {
				newContent := utf8string.NewString(currMark.Content)

				// NOTE: Add the new inner annotation of "current" annotation
				// NOTE: Add before we modify the "current" annotation
				additionalMarks = append(additionalMarks, &biblev1.Mark{
					Id:          currMark.Id,
					Content:     newContent.Slice(int(prevMark.StartOffset)-int(currMark.StartOffset), newContent.RuneCount()),
					Kind:        currMark.Kind,
					Label:       currMark.Label,
					SortOrder:   currMark.SortOrder,
					StartOffset: prevMark.StartOffset,
					EndOffset:   currMark.EndOffset,
					TargetId:    currMark.TargetId,
					TargetType:  currMark.TargetType,
					ChapterId:   currMark.ChapterId,
				})

				// Update current annotation to end before overlap
				sortedMarks[i] = currMark
				sortedMarks[i].EndOffset = prevMark.StartOffset
				sortedMarks[i].Content = newContent.Slice(0, int(prevMark.StartOffset)-int(currMark.StartOffset))
			} else {
				// NOTE: If overlapKeepRight is false then we keep "current" annotation,
				// update the start of the "prev" annotation to the end of
				// the "current" annotation, and push the new inner annotation of
				// "prev" annotation to the array

				// NOTE: Add the new inner annotation of "prev" annotation
				// NOTE: Add before we modify the "prev" annotation
				newContent := utf8string.NewString(prevMark.Content)

				// NOTE: Add the new inner annotation of "prev" annotation
				additionalMarks = append(additionalMarks, &biblev1.Mark{
					Id:          prevMark.Id,
					Content:     newContent.Slice(0, int(currMark.EndOffset)-int(prevMark.StartOffset)),
					Kind:        prevMark.Kind,
					Label:       prevMark.Label,
					SortOrder:   prevMark.SortOrder,
					StartOffset: prevMark.StartOffset,
					EndOffset:   currMark.EndOffset,
					TargetId:    prevMark.TargetId,
					TargetType:  prevMark.TargetType,
					ChapterId:   prevMark.ChapterId,
				})

				// NOTE: Update prev annotation to start after overlap
				sortedMarks[i-1] = prevMark
				sortedMarks[i-1].StartOffset = currMark.EndOffset
				sortedMarks[i-1].Content = newContent.Slice(int(currMark.EndOffset)-int(prevMark.StartOffset), newContent.RuneCount())
			}
		}
	}

	// NOTE: Return the resolved annotations with additional annotations
	sortedMarks = append(sortedMarks, additionalMarks...)

	return slices.SortedFunc(slices.Values(sortedMarks), func(a, b *biblev1.Mark) int {
		return cmp.Or(cmp.Compare(a.StartOffset, b.StartOffset), cmp.Compare(a.EndOffset, b.EndOffset))
	})
}
