package model

import (
	"testing"
	"time"
)

func TestGroupSessionsBySlug_Empty(t *testing.T) {
	result := GroupSessionsBySlug(nil)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestGroupSessionsBySlug_NoSlugs(t *testing.T) {
	sessions := []*Session{
		{ID: "a", ModTime: time.Unix(300, 0)},
		{ID: "b", ModTime: time.Unix(200, 0)},
		{ID: "c", ModTime: time.Unix(100, 0)},
	}
	result := GroupSessionsBySlug(sessions)
	if len(result) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(result))
	}
	// Should be sorted by ModTime descending
	if result[0].ID != "a" || result[1].ID != "b" || result[2].ID != "c" {
		t.Errorf("expected a,b,c order; got %s,%s,%s", result[0].ID, result[1].ID, result[2].ID)
	}
	// No group representatives
	for _, s := range result {
		if s.IsGroupRepresentative() {
			t.Errorf("session %s should not be a group representative", s.ID)
		}
	}
}

func TestGroupSessionsBySlug_CollapsedToSingleRow(t *testing.T) {
	sessions := []*Session{
		{ID: "s1", Slug: "fizzy-stallman", ModTime: time.Unix(100, 0), NumTurns: 5, AgentCount: 1, FileSize: 1000},
		{ID: "s2", Slug: "fizzy-stallman", ModTime: time.Unix(300, 0), NumTurns: 3, AgentCount: 2, FileSize: 2000},
		{ID: "s3", ModTime: time.Unix(250, 0), NumTurns: 1},
		{ID: "s4", Slug: "cool-turing", ModTime: time.Unix(400, 0), NumTurns: 7, AgentCount: 1, FileSize: 500},
	}
	result := GroupSessionsBySlug(sessions)

	// cool-turing (1 session, solo), fizzy-stallman (collapsed to 1), s3 (ungrouped) = 3 rows
	if len(result) != 3 {
		t.Fatalf("expected 3 rows (collapsed), got %d", len(result))
	}

	// cool-turing first (maxMod=400), solo — no GroupSessions
	if result[0].ID != "s4" {
		t.Errorf("expected s4 first (cool-turing, mod=400), got %s", result[0].ID)
	}
	if result[0].IsGroupRepresentative() {
		t.Error("solo slug s4 should not be a group representative")
	}

	// fizzy-stallman representative = s2 (newest)
	if result[1].ID != "s2" {
		t.Errorf("expected s2 as fizzy-stallman representative, got %s", result[1].ID)
	}
	if !result[1].IsGroupRepresentative() {
		t.Error("s2 should be a group representative")
	}

	// s3 last (ungrouped)
	if result[2].ID != "s3" {
		t.Errorf("expected s3 last, got %s", result[2].ID)
	}
}

func TestGroupSessionsBySlug_Aggregation(t *testing.T) {
	sessions := []*Session{
		{
			ID: "s1", Slug: "grp", ModTime: time.Unix(100, 0),
			NumTurns: 5, AgentCount: 1, FileSize: 1000,
			TokensByModel: map[string]TokenCount{
				"opus": {InputTokens: 100, OutputTokens: 200},
			},
		},
		{
			ID: "s2", Slug: "grp", ModTime: time.Unix(200, 0),
			NumTurns: 3, AgentCount: 2, FileSize: 2000,
			TokensByModel: map[string]TokenCount{
				"opus":   {InputTokens: 50, OutputTokens: 50},
				"sonnet": {InputTokens: 300, OutputTokens: 400},
			},
		},
	}
	result := GroupSessionsBySlug(sessions)
	if len(result) != 1 {
		t.Fatalf("expected 1 collapsed row, got %d", len(result))
	}

	rep := result[0]
	if rep.ID != "s2" {
		t.Errorf("expected s2 as representative, got %s", rep.ID)
	}
	if rep.NumTurns != 8 {
		t.Errorf("expected NumTurns=8 (5+3), got %d", rep.NumTurns)
	}
	if rep.AgentCount != 3 {
		t.Errorf("expected AgentCount=3 (1+2), got %d", rep.AgentCount)
	}
	if rep.FileSize != 3000 {
		t.Errorf("expected FileSize=3000 (1000+2000), got %d", rep.FileSize)
	}
	// TokensByModel: opus(150+250) + sonnet(300+400)
	opus := rep.TokensByModel["opus"]
	if opus.InputTokens != 150 || opus.OutputTokens != 250 {
		t.Errorf("expected opus tokens 150/250, got %d/%d", opus.InputTokens, opus.OutputTokens)
	}
	sonnet := rep.TokensByModel["sonnet"]
	if sonnet.InputTokens != 300 || sonnet.OutputTokens != 400 {
		t.Errorf("expected sonnet tokens 300/400, got %d/%d", sonnet.InputTokens, sonnet.OutputTokens)
	}
}

func TestGroupSessionsBySlug_GroupSessionsOrder(t *testing.T) {
	sessions := []*Session{
		{ID: "s3", Slug: "grp", ModTime: time.Unix(300, 0)},
		{ID: "s1", Slug: "grp", ModTime: time.Unix(100, 0)},
		{ID: "s2", Slug: "grp", ModTime: time.Unix(200, 0)},
	}
	result := GroupSessionsBySlug(sessions)
	if len(result) != 1 {
		t.Fatalf("expected 1 collapsed row, got %d", len(result))
	}
	rep := result[0]
	if len(rep.GroupSessions) != 3 {
		t.Fatalf("expected 3 GroupSessions, got %d", len(rep.GroupSessions))
	}
	// Should be sorted oldest-first
	if rep.GroupSessions[0].ID != "s1" || rep.GroupSessions[1].ID != "s2" || rep.GroupSessions[2].ID != "s3" {
		t.Errorf("expected GroupSessions sorted oldest-first: s1,s2,s3; got %s,%s,%s",
			rep.GroupSessions[0].ID, rep.GroupSessions[1].ID, rep.GroupSessions[2].ID)
	}
}

func TestGroupSessionsBySlug_TopicFromOldestSession(t *testing.T) {
	sessions := []*Session{
		{ID: "s1", Slug: "grp", ModTime: time.Unix(100, 0), Topic: "in history view, collapse rows"},
		{ID: "s2", Slug: "grp", ModTime: time.Unix(200, 0), Topic: "Implement the following plan:"},
	}
	result := GroupSessionsBySlug(sessions)
	if len(result) != 1 {
		t.Fatalf("expected 1 collapsed row, got %d", len(result))
	}
	if result[0].Topic != "in history view, collapse rows" {
		t.Errorf("expected topic from oldest session, got %q", result[0].Topic)
	}
}

func TestGroupSessionsBySlug_TopicSkipsEmptyOldest(t *testing.T) {
	sessions := []*Session{
		{ID: "s1", Slug: "grp", ModTime: time.Unix(100, 0), Topic: ""},
		{ID: "s2", Slug: "grp", ModTime: time.Unix(200, 0), Topic: "real topic"},
	}
	result := GroupSessionsBySlug(sessions)
	if len(result) != 1 {
		t.Fatalf("expected 1 collapsed row, got %d", len(result))
	}
	if result[0].Topic != "real topic" {
		t.Errorf("expected first non-empty topic, got %q", result[0].Topic)
	}
}

func TestGroupNameCell_Solo(t *testing.T) {
	s := &Session{ID: "d2559feb-1234-5678-9abc-def012345678"}
	if got := s.GroupNameCell(); got != "d2559feb" {
		t.Errorf("expected 'd2559feb', got %q", got)
	}
}

func TestGroupNameCell_Group(t *testing.T) {
	s := &Session{
		ID: "360eb907-1234-5678-9abc-def012345678",
		GroupSessions: []*Session{
			{ID: "d2559feb-1234-5678-9abc-def012345678"},
			{ID: "360eb907-1234-5678-9abc-def012345678"},
		},
	}
	if got := s.GroupNameCell(); got != "d2559feb..360eb907" {
		t.Errorf("expected 'd2559feb..360eb907', got %q", got)
	}
}

func TestGroupSessionsBySlug_NoCacheMutation(t *testing.T) {
	sessions := []*Session{
		{ID: "s1", Slug: "grp", ModTime: time.Unix(100, 0), NumTurns: 5, FileSize: 1000},
		{ID: "s2", Slug: "grp", ModTime: time.Unix(200, 0), NumTurns: 3, FileSize: 2000},
	}
	r1 := GroupSessionsBySlug(sessions)
	r2 := GroupSessionsBySlug(sessions)

	if r1[0].NumTurns != 8 {
		t.Errorf("first call: expected NumTurns=8, got %d", r1[0].NumTurns)
	}
	if r2[0].NumTurns != 8 {
		t.Errorf("second call: expected NumTurns=8, got %d (double-counted)", r2[0].NumTurns)
	}
	// Original sessions must not be mutated
	if sessions[1].NumTurns != 3 {
		t.Errorf("original s2 mutated: expected NumTurns=3, got %d", sessions[1].NumTurns)
	}
}

func TestIsGroupRepresentative(t *testing.T) {
	solo := &Session{ID: "solo"}
	if solo.IsGroupRepresentative() {
		t.Error("solo session should not be representative")
	}

	rep := &Session{
		ID:            "rep",
		GroupSessions: []*Session{{ID: "a"}, {ID: "rep"}},
	}
	if !rep.IsGroupRepresentative() {
		t.Error("session with GroupSessions > 1 should be representative")
	}
}
