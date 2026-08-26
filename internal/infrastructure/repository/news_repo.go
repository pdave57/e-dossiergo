package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/pagination"
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// NEWS & ANNOUNCEMENT REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type newsRepo struct{ db *sql.DB }

func NewNewsAnnouncementRepository(db *sql.DB) domain.NewsAnnouncementRepository {
	return &newsRepo{db: db}
}

func (r *newsRepo) Create(ctx context.Context, n *domain.NewsAnnouncement) error {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	now := time.Now()
	n.CreatedAt, n.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO news_announcements
		 (id,state_id,type,headline,sub_headline,news_body,news_date,posted_by,created_at,updated_at,created_by,updated_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		n.ID, n.StateID, n.Type, n.Headline, nullableString(n.SubHeadline),
		n.NewsBody, n.NewsDate, n.PostedBy, n.CreatedAt, n.UpdatedAt, n.CreatedBy, n.UpdatedBy)
	if err != nil {
		if isFKViolation(err) {
			return apperror.BadRequest("invalid state reference")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *newsRepo) GetByID(ctx context.Context, id string) (*domain.NewsAnnouncement, error) {
	return scanNews(r.db.QueryRowContext(ctx,
		newsSelectSQL+" WHERE n.id=$1 AND n.deleted_at IS NULL", id))
}

func (r *newsRepo) Update(ctx context.Context, n *domain.NewsAnnouncement) error {
	n.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE news_announcements SET type=$1,headline=$2,sub_headline=$3,news_body=$4,news_date=$5,
		  posted_by=$6,updated_at=$7,updated_by=$8
		 WHERE id=$9 AND deleted_at IS NULL`,
		n.Type, n.Headline, nullableString(n.SubHeadline), n.NewsBody,
		n.NewsDate, n.PostedBy, n.UpdatedAt, n.UpdatedBy, n.ID)
	return checkRowsAffected(res, err, "news", n.ID)
}

func (r *newsRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE news_announcements SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "news", id)
}

func (r *newsRepo) List(ctx context.Context, f domain.NewsFilter, p pagination.Params) ([]*domain.NewsAnnouncement, int, error) {
	where, args := []string{"n.deleted_at IS NULL"}, []any{}
	idx := 1
	if f.StateID != "" {
		where = append(where, fmt.Sprintf("n.state_id=$%d", idx))
		args = append(args, f.StateID)
		idx++
	}
	if f.Type != "" {
		where = append(where, fmt.Sprintf("n.type=$%d", idx))
		args = append(args, f.Type)
		idx++
	}
	if f.From != nil {
		where = append(where, fmt.Sprintf("n.news_date>=$%d", idx))
		args = append(args, *f.From)
		idx++
	}
	if f.To != nil {
		where = append(where, fmt.Sprintf("n.news_date<=$%d", idx))
		args = append(args, *f.To)
		idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(n.headline ILIKE $%d OR n.sub_headline ILIKE $%d OR n.news_body ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+f.Search+"%")
		idx++
	}

	clause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM news_announcements n "+clause, args...).Scan(&total); err != nil {
		return nil, 0, apperror.Internal(err)
	}

	args = append(args, p.PerPage, p.Offset)
	q := fmt.Sprintf("%s %s ORDER BY n.news_date DESC, n.created_at DESC LIMIT $%d OFFSET $%d",
		newsSelectSQL, clause, idx, idx+1)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	defer rows.Close()

	var out []*domain.NewsAnnouncement
	for rows.Next() {
		n, err := scanNews(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, n)
	}
	return out, total, rows.Err()
}

const newsSelectSQL = `
	SELECT n.id,n.state_id,n.type,n.headline,COALESCE(n.sub_headline,''),n.news_body,COALESCE(n.news_body,''),n.news_date,n.posted_by,
	       n.created_at,n.updated_at,
	       COALESCE(n.created_by,''),COALESCE(n.updated_by,'')
	FROM news_announcements n`

func scanNews(s scanner) (*domain.NewsAnnouncement, error) {
	n := &domain.NewsAnnouncement{}
	err := s.Scan(&n.ID, &n.StateID, &n.Type, &n.Headline, &n.SubHeadline, &n.NewsBody,
		&n.NewsDate, &n.PostedBy,
		&n.CreatedAt, &n.UpdatedAt, &n.CreatedBy, &n.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("news", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return n, nil
}
