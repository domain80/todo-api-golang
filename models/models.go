package models

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

type Models struct {
	db    *sql.DB
	User  userModel
	Todos todoModel
}

var ErrDuplicateUsername = errors.New("models err: duplicate username ")

// connect to the database using the provided dsn
func Init(dsn string) (Models, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return Models{}, err
	}
	// todo: see if there is a better way to handle the db connectin pool without passing it around
	return Models{
		db:    db,
		User:  userModel{db: db},
		Todos: todoModel{db: db},
	}, err
}
