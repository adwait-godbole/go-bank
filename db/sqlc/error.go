package db

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// The below error numbers come from PostgreSQL itself. lib/pq had there mappings with strings like "unique_violation",
// "foreign_key_violation" but pgx doesn't have such kind of mappings. So we are adding here manually.
const (
	ForeignKeyViolation = "23503"
	UniqueViolation     = "23505"
)

var ErrRecordNotFound = pgx.ErrNoRows

var ErrUniqueViolation = &pgconn.PgError{
	Code: UniqueViolation,
}

func ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// fmt.Println(">> ", pgErr.ConstraintName) // "users_email_key", "users_pkey"
		return pgErr.Code
	}
	return ""
}
