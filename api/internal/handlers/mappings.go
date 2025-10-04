// api/internal/handlers/mappings.go
package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
)

type ApplyRequest struct {
	Headers []string           `json:"headers"`
	Rows    [][]string         `json:"rows"`
	Rules   map[string]*string `json:"rules"` // dest -> source (null 可)
}

type ApplyResponse struct {
	NormalizedHeaders []string   `json:"normalizedHeaders"`
	NormalizedRows    [][]string `json:"normalizedRows"`
}

// POST /api/mappings/apply
func ApplyMapping(w http.ResponseWriter, r *http.Request) {
	var in ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// source header -> index
	idx := make(map[string]int, len(in.Headers))
	for i, h := range in.Headers {
		idx[h] = i
	}

	// スキーマは rules のキー順（安定化）
	schema := make([]string, 0, len(in.Rules))
	for k := range in.Rules {
		schema = append(schema, k)
	}
	sort.Strings(schema)

	out := ApplyResponse{
		NormalizedHeaders: schema,
		NormalizedRows:    make([][]string, 0, len(in.Rows)),
	}

	for _, row := range in.Rows {
		dest := make([]string, len(schema))
		for j, col := range schema {
			src := in.Rules[col]
			if src == nil {
				dest[j] = ""
				continue
			}
			if i, ok := idx[*src]; ok && i < len(row) {
				dest[j] = row[i]
			} else {
				dest[j] = ""
			}
		}
		out.NormalizedRows = append(out.NormalizedRows, dest)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
