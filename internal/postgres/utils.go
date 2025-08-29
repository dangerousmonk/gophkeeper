package postgres

import (
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// isUniqueViolation checks if an error is a unique constraint error for given constraint name
func isUniqueViolation(err error, constraint string) bool {
	if pgError, ok := err.(*pgconn.PgError); ok {
		return pgError.Code == pgerrcode.UniqueViolation && pgError.ConstraintName == constraint
	}
	return false
}
