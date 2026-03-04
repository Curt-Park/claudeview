package model

import "sort"

// GroupSessionsBySlug collapses sessions sharing a slug into a single representative row.
// The representative is the newest session (last in asc-sorted group).
// Aggregated fields: NumTurns, AgentCount, FileSize, TokensByModel.
// Representative.GroupSessions = all sessions in the group (oldest-first).
// Solo-slug sessions (group of 1) keep GroupSessions = nil.
// The returned list is sorted by latest ModTime descending.
func GroupSessionsBySlug(sessions []*Session) []*Session {
	if len(sessions) == 0 {
		return sessions
	}

	// Partition into slug groups and ungrouped
	groups := make(map[string][]*Session)
	var ungrouped []*Session
	var slugOrder []string
	seen := make(map[string]bool)

	for _, s := range sessions {
		if s.Slug == "" {
			ungrouped = append(ungrouped, s)
			continue
		}
		if !seen[s.Slug] {
			seen[s.Slug] = true
			slugOrder = append(slugOrder, s.Slug)
		}
		groups[s.Slug] = append(groups[s.Slug], s)
	}

	// Sort within each group by ModTime ascending (oldest first)
	for _, slug := range slugOrder {
		sort.Slice(groups[slug], func(i, j int) bool {
			return groups[slug][i].ModTime.Before(groups[slug][j].ModTime)
		})
	}

	// Build sortable items: each slug group collapsed into one representative
	type sortItem struct {
		maxMod int64
		rep    *Session
	}

	var sortItems []sortItem
	for _, slug := range slugOrder {
		g := groups[slug]
		rep := g[len(g)-1] // newest = representative

		if len(g) > 1 {
			// Aggregate stats into representative using separate accumulators
			// (rep is g[last], so we can't zero-and-sum in place).
			var totalTurns, totalAgents int
			var totalSize int64
			merged := make(map[string]TokenCount)
			for _, s := range g {
				totalTurns += s.NumTurns
				totalAgents += s.AgentCount
				totalSize += s.FileSize
				for m, tc := range s.TokensByModel {
					cur := merged[m]
					cur.InputTokens += tc.InputTokens
					cur.OutputTokens += tc.OutputTokens
					merged[m] = cur
				}
			}
			rep.NumTurns = totalTurns
			rep.AgentCount = totalAgents
			rep.FileSize = totalSize
			rep.TokensByModel = merged
			rep.GroupSessions = g
			// Use the first non-empty topic from the oldest session forward,
			// since the merged history view shows sessions oldest-first.
			for _, s := range g {
				if s.Topic != "" {
					rep.Topic = s.Topic
					break
				}
			}
		}
		// Solo groups (len==1): GroupSessions stays nil

		sortItems = append(sortItems, sortItem{maxMod: rep.ModTime.Unix(), rep: rep})
	}
	for _, s := range ungrouped {
		sortItems = append(sortItems, sortItem{maxMod: s.ModTime.Unix(), rep: s})
	}

	// Sort all items by maxMod descending (most recent first)
	sort.SliceStable(sortItems, func(i, j int) bool {
		return sortItems[i].maxMod > sortItems[j].maxMod
	})

	// Flatten to single representative per group
	result := make([]*Session, 0, len(sortItems))
	for _, si := range sortItems {
		result = append(result, si.rep)
	}
	return result
}
