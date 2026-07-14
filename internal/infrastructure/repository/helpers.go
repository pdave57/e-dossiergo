package repository

import (
	"database/sql"
	"errors"

	"github.com/edossier/api/pkg/apperror"
	"github.com/lib/pq"
)

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

// checkRowsAffected returns NotFound if no rows were updated/deleted.
func checkRowsAffected(res sql.Result, err error, entity, id string) error {
	if err != nil {
		return apperror.Internal(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperror.Internal(err)
	}
	if n == 0 {
		return apperror.NotFound(entity, id)
	}
	return nil
}

// isUniqueViolation detects PostgreSQL unique constraint errors (code 23505).
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

// isFKViolation detects foreign-key constraint errors (code 23503).
func isFKViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23503"
	}
	return false
}

// nullableString returns sql.NullString for optional fields.
func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
