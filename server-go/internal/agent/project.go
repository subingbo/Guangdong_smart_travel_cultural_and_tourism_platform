package agent

import (
	"encoding/json"
	"fmt"
)

// Projector keeps tool output from blowing up the model context. Search
// results are small by construction, but planner/polyline payloads are not.
type Projector struct {
	MaxBytes  int
	MaxString int
	MaxArray  int
}

func NewProjector(maxBytes int) *Projector {
	if maxBytes <= 0 {
		maxBytes = 16 * 1024
	}
	return &Projector{MaxBytes: maxBytes, MaxString: 300, MaxArray: 50}
}

func (p *Projector) Project(tool, content string) string {
	if len(content) <= p.MaxBytes {
		return content
	}
	var value any
	if err := json.Unmarshal([]byte(content), &value); err != nil {
		return truncateRunes(content, p.MaxBytes) + "\n...[truncated]"
	}
	shrunk := p.shrink(value, true)
	encoded, err := json.Marshal(shrunk)
	if err != nil || len(encoded) <= p.MaxBytes {
		if err != nil {
			return truncateRunes(content, p.MaxBytes) + "\n...[truncated]"
		}
		return string(encoded)
	}
	return truncateRunes(string(encoded), p.MaxBytes) + "\n...[truncated]"
}

func (p *Projector) shrink(value any, top bool) any {
	switch v := value.(type) {
	case string:
		if len(v) > p.MaxString {
			return truncateRunes(v, p.MaxString) + "…"
		}
		return v
	case []any:
		limit := len(v)
		if limit > p.MaxArray {
			limit = p.MaxArray
		}
		out := make([]any, 0, limit)
		for i := 0; i < limit; i++ {
			out = append(out, p.shrink(v[i], false))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			switch key {
			case "polyline", "geometry", "geom", "gemo":
				out[key] = "[omitted]"
			default:
				out[key] = p.shrink(item, false)
			}
		}
		if top {
			out["_projected"] = fmt.Sprintf("tool output truncated from %d bytes", len(fmt.Sprint(value)))
		}
		return out
	default:
		return value
	}
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
