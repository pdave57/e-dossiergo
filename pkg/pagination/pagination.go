// Package pagination provides query-page helpers for list endpoints.
package pagination

import (
	"net/http"
	"strconv"
)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

// Params holds parsed page/per_page values from a request.
type Params struct {
	Page    int
	PerPage int
	Offset  int
}

// Meta is returned in list responses.
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Parse extracts pagination params from an HTTP request.
func Parse(r *http.Request) Params {
	page := intQuery(r, "page", defaultPage)
	perPage := intQuery(r, "per_page", defaultPerPage)

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return Params{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
	}
}

// BuildMeta constructs a Meta for a response.
func BuildMeta(p Params, total int) Meta {
	totalPages := total / p.PerPage
	if total%p.PerPage != 0 {
		totalPages++
	}
	return Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

func intQuery(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
