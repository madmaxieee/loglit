package renderer

import "sort"

type MatchLayer []Match

func (matches MatchLayer) Len() int {
	return len(matches)
}

func (matches *MatchLayer) Sort() *MatchLayer {
	sort.SliceStable(*matches, func(i, j int) bool {
		a, b := (*matches)[i], (*matches)[j]
		if a.Start == b.Start {
			return a.End > b.End
		}
		return a.Start < b.Start
	})
	return matches
}

func (matches *MatchLayer) removeOverlaps() *MatchLayer {
	var validMatches MatchLayer
	// NOTE: later matches have higher priority
	for i := len(*matches) - 1; i >= 0; i-- {
		match := (*matches)[i]
		collision := false
		for _, existingMatch := range validMatches {
			if !(match.End <= existingMatch.Start || match.Start >= existingMatch.End) {
				collision = true
				break
			}
		}
		if !collision {
			validMatches = append(validMatches, match)
		}
	}
	*matches = validMatches
	return matches
}

func Stack(top MatchLayer, bottom MatchLayer) MatchLayer {
	out := MatchLayer{}
	iTop, iBot := 0, 0

	for iTop < len(top) && iBot < len(bottom) {
		top := top[iTop]
		bot := bottom[iBot]

		// no overlap: top before bottom
		if top.End <= bot.Start {
			out = append(out, top)
			iTop++
			continue
		}
		// no overlap: bottom before top
		if bot.End <= top.Start {
			out = append(out, bot)
			iBot++
			continue
		}

		// top covers bottom entirely: skip bottom
		if top.Start <= bot.Start && top.End >= bot.End {
			iBot++
			continue
		}

		// bottom covers top entirely: add left remainder, top, and adjust bottom
		if bot.Start <= top.Start && bot.End >= top.End {
			if bot.Start != top.Start {
				leftRemainder := Match{
					Start:     bot.Start,
					End:       top.Start,
					AnsiStart: bot.AnsiStart,
					AnsiEnd:   bot.AnsiEnd,
				}
				out = append(out, leftRemainder)
			}

			out = append(out, top)
			iTop++

			// right remainder, will be compared against next top
			if bot.End > top.End {
				bottom[iBot].Start = top.End
			} else {
				// if theres is no right remainder, drop the bottom
				iBot++
			}

			continue
		}

		// partial overlap: truncate bottom
		if bot.Start < top.Start && bot.End > top.Start {
			bottom[iBot].End = top.Start
			out = append(out, bottom[iBot])
			iBot++
			continue
		}
		if bot.Start < top.End && bot.End > top.End {
			bottom[iBot].Start = top.End
			out = append(out, top)
			iTop++
			continue
		}
	}

	for iTop < len(top) {
		out = append(out, top[iTop])
		iTop++
	}
	for iBot < len(bottom) {
		out = append(out, bottom[iBot])
		iBot++
	}

	return out
}
