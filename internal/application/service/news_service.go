package service

import (
	"context"
	"strings"
	"time"

	"github.com/edossier/api/internal/application/dto"
	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/pagination"
	"github.com/edossier/api/pkg/validator"
)

// NewsService handles news items and announcements.
//
// Reads are served publicly (no authentication); create, update and delete are
// gated at the router by the news:create / news:update / news:delete permissions.
type NewsService struct {
	news domain.NewsAnnouncementRepository
}

func NewNewsService(news domain.NewsAnnouncementRepository) *NewsService {
	return &NewsService{news: news}
}

var newsTypes = []string{
	string(domain.NewsTypeNews),
	string(domain.NewsTypeAnnouncement),
}

// Create publishes a news item or announcement.
// stateID is the owning state, resolved by the handler from the body or the JWT.
func (uc *NewsService) Create(ctx context.Context, stateID string, req dto.CreateNewsRequest, createdBy string) (*domain.NewsAnnouncement, error) {
	v := validator.New().
		Required(stateID, "state_id").
		Required(req.Headline, "headline").
		MaxLen(req.Headline, 300, "headline").
		MaxLen(req.SubHeadline, 500, "sub_headline").
		Required(req.PostedBy, "posted_by").
		MaxLen(req.PostedBy, 200, "posted_by").
		Check(!req.NewsDate.IsZero(), "news_date", "is required")
	if req.Type != "" {
		v.OneOf(strings.ToUpper(req.Type), newsTypes, "type")
	}
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	n := &domain.NewsAnnouncement{
		StateID:     stateID,
		Type:        newsType(req.Type),
		Headline:    strings.TrimSpace(req.Headline),
		SubHeadline: strings.TrimSpace(req.SubHeadline),
		NewsBody:    strings.TrimSpace(req.NewsBody),
		NewsDate:    truncateToDay(req.NewsDate),
		PostedBy:    strings.TrimSpace(req.PostedBy),
		AuditFields: domain.AuditFields{CreatedBy: createdBy, UpdatedBy: createdBy},
	}
	if err := uc.news.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

// GetByID returns a single news item. Public — no authentication required.
func (uc *NewsService) GetByID(ctx context.Context, id string) (*domain.NewsAnnouncement, error) {
	if id == "" {
		return nil, apperror.BadRequest("id is required")
	}
	return uc.news.GetByID(ctx, id)
}

// List returns a paginated, newest-first feed. Public — no authentication required.
func (uc *NewsService) List(ctx context.Context, f domain.NewsFilter, p pagination.Params) ([]*domain.NewsAnnouncement, int, error) {
	if f.Type != "" {
		f.Type = strings.ToUpper(f.Type)
	}
	return uc.news.List(ctx, f, p)
}

// Update edits an existing news item. The owning state is immutable.
func (uc *NewsService) Update(ctx context.Context, id string, req dto.UpdateNewsRequest, updatedBy string) (*domain.NewsAnnouncement, error) {
	if id == "" {
		return nil, apperror.BadRequest("id is required")
	}

	v := validator.New().
		Required(req.Headline, "headline").
		MaxLen(req.Headline, 300, "headline").
		MaxLen(req.SubHeadline, 500, "sub_headline").
		Required(req.PostedBy, "posted_by").
		MaxLen(req.PostedBy, 200, "posted_by").
		Check(!req.NewsDate.IsZero(), "news_date", "is required")
	if req.Type != "" {
		v.OneOf(strings.ToUpper(req.Type), newsTypes, "type")
	}
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	n, err := uc.news.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Type != "" {
		n.Type = newsType(req.Type)
	}
	n.Headline = strings.TrimSpace(req.Headline)
	n.SubHeadline = strings.TrimSpace(req.SubHeadline)
	n.NewsBody = strings.TrimSpace(req.NewsBody)
	n.NewsDate = truncateToDay(req.NewsDate)
	n.PostedBy = strings.TrimSpace(req.PostedBy)
	n.UpdatedBy = updatedBy

	if err := uc.news.Update(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

// Delete soft-deletes a news item.
func (uc *NewsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return apperror.BadRequest("id is required")
	}
	return uc.news.Delete(ctx, id)
}

// newsType normalises the wire value, defaulting to NEWS when omitted.
func newsType(s string) domain.NewsType {
	if s == "" {
		return domain.NewsTypeNews
	}
	return domain.NewsType(strings.ToUpper(s))
}

// truncateToDay drops the time component — news_date is a calendar date.
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
