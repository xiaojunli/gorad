package db

import (
	_ "github.com/go-sql-driver/mysql" //mysql driver
	"github.com/ilibs/gosql"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

func ConnectDB() {
	configs := make(map[string]*gosql.Config)

	configs["default"] = &gosql.Config{
		Enable:  true,
		Driver:  "mysql",
		Dsn:     "root:gorad@tcp(127.0.0.1:3306)/test?charset=utf8",
		ShowSql: true,
	}

	//connection database
	gosql.Connect(configs)
}

/**
 exec
 */
func Execx(query string, args ...interface{}) (sql.Result, error) {
	return gosql.Exec(query, args...)
}
/**
 query
 */
func Queryx(query string, args ...interface{})  (*sqlx.Rows, error) {
	return gosql.Queryx(query, args...)
}

/**
 get rows
 */
func QueryRowx(query string, args ...interface{}) (*sqlx.Row){
	return gosql.QueryRowx(query, args...)
}

/**
 get row for struct
 */
func Getx(dest interface{}, query string, args ...interface{}) error {
	return gosql.Get(dest, query, args...)
}

/**
 select
 */
func Selectx(dest interface{}, query string, args ...interface{}) error {
	return gosql.Select(dest, query, args...)
}

/**
  charge database
 */
func Use(db string)  *gosql.Wrapper {
	return gosql.Use(db)
}