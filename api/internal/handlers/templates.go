// api/internal/handlers/templates.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"csv-import-kit/api/internal/store"

	"github.com/jackc/pgx/v5"
)

type Template struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	SchemaKey   string                 `json:"schema_key"`
	Rules       map[string]interface{} `json:"rules"`
	Description *string                `json:"description,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type TemplateCreateReq struct {
	Name        string                 `json:"name"`
	SchemaKey   string                 `json:"schema_key"`
	Rules       map[string]interface{} `json:"rules"`
	Description *string                `json:"description,omitempty"`
}

type TemplateCreateResp struct {
	ID string `json:"id"`
}

type TemplateHandler struct {
	Store *store.Store
}

func NewTemplateHandler(s *store.Store) *TemplateHandler {
	return &TemplateHandler{Store: s}
}

// POST /api/templates
func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var in TemplateCreateReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if in.Name == "" || in.SchemaKey == "" || in.Rules == nil {
		http.Error(w, "name, schema_key, rules are required", http.StatusBadRequest)
		return
	}

	b, err := json.Marshal(in.Rules)
	if err != nil {
		http.Error(w, "invalid rules", http.StatusBadRequest)
		return
	}

	const q = `
insert into public.mapping_templates (name, schema_key, rules, description)
values ($1, $2, $3::jsonb, $4)
returning id;
`
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var id string
	if err := h.Store.Pool.QueryRow(ctx, q, in.Name, in.SchemaKey, string(b), in.Description).Scan(&id); err != nil {
		http.Error(w, "db insert error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(TemplateCreateResp{ID: id})
}

// GET /api/templates  （最新20件）
func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	const q = `
select id, name, schema_key, rules, description, created_at, updated_at
from public.mapping_templates
order by created_at desc
limit 20;
`
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.Store.Pool.Query(ctx, q)
	if err != nil {
		http.Error(w, "db query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := make([]Template, 0, 20)
	for rows.Next() {
		var t Template
		var rawRules []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.SchemaKey, &rawRules, &t.Description, &t.CreatedAt, &t.UpdatedAt); err != nil {
			http.Error(w, "db scan error", http.StatusInternalServerError)
			return
		}
		if len(rawRules) > 0 {
			if err := json.Unmarshal(rawRules, &t.Rules); err != nil {
				http.Error(w, "rules unmarshal error", http.StatusInternalServerError)
				return
			}
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "db rows error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
