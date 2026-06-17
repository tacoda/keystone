package primitive

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SearchHit is one match returned by Search.
type SearchHit struct {
	Primitive Primitive `json:"primitive"`
	Score     int       `json:"score"`
	Excerpt   string    `json:"excerpt"`
	Where     string    `json:"where"` // "id" | "description" | "body" | "globs" | "traces"
}

// Search runs a simple substring-frequency search across every
// primitive: id, description, body, globs, traces. Ranks by a small
// weighted score; ties broken alphabetically.
//
// Scoring (deliberately naive — readable + good-enough for a
// few-hundred-primitive harness):
//
//   id match            +10
//   description match    +5
//   globs / traces       +3
//   body substring       +1
//
// Returns at most `limit` results (0 = all). Query is lowercased
// and matched case-insensitively.
func Search(projectDir string, primitives []Primitive, query string, limit int) []SearchHit {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	var hits []SearchHit
	for _, p := range primitives {
		score := 0
		where := ""
		if strings.Contains(strings.ToLower(p.ID), q) {
			score += 10
			where = "id"
		}
		if strings.Contains(strings.ToLower(p.Description), q) {
			score += 5
			if where == "" {
				where = "description"
			}
		}
		for _, g := range p.Globs {
			if strings.Contains(strings.ToLower(g), q) {
				score += 3
				if where == "" {
					where = "globs"
				}
				break
			}
		}
		for _, t := range p.Traces {
			if strings.Contains(strings.ToLower(t), q) {
				score += 3
				if where == "" {
					where = "traces"
				}
				break
			}
		}
		excerpt := ""
		if body, err := os.ReadFile(filepath.Join(projectDir, p.Path)); err == nil {
			low := strings.ToLower(string(body))
			if idx := strings.Index(low, q); idx >= 0 {
				score += 1
				if where == "" {
					where = "body"
				}
				excerpt = excerptAround(string(body), idx, len(q), 80)
			}
		}
		if score == 0 {
			continue
		}
		hits = append(hits, SearchHit{
			Primitive: p,
			Score:     score,
			Where:     where,
			Excerpt:   excerpt,
		})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		if hits[i].Primitive.Kind != hits[j].Primitive.Kind {
			return hits[i].Primitive.Kind < hits[j].Primitive.Kind
		}
		return hits[i].Primitive.ID < hits[j].Primitive.ID
	})
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	return hits
}

// excerptAround returns ~window chars on either side of pos, with
// ellipses when truncated. Preserves byte offsets — fine for ASCII;
// good-enough for unicode bodies the dashboard only displays.
func excerptAround(body string, pos, matchLen, window int) string {
	start := pos - window
	if start < 0 {
		start = 0
	}
	end := pos + matchLen + window
	if end > len(body) {
		end = len(body)
	}
	out := strings.TrimSpace(body[start:end])
	if start > 0 {
		out = "…" + out
	}
	if end < len(body) {
		out = out + "…"
	}
	return strings.ReplaceAll(out, "\n", " ")
}
