package mysqldb

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func NewDb(host, port, scheme, user, password string) (*sqlx.DB, error) {
	return NewDbParams(host, port, scheme, user, password, map[string]string{})
}

func NewDbWithMutiStatements(host, port, scheme, user, password string) (*sqlx.DB, error) {
	return NewDbParams(host, port, scheme, user, password, map[string]string{
		"multiStatements": "true",
	})
}

func NewDbParams(host, port, scheme, user, password string, parameters map[string]string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&columnsWithAlias=false&&loc=Local", user, password, host, port, scheme)
	for k, v := range parameters {
		dsn += fmt.Sprintf("&%s=%s", k, v)
	}

	tries := 0
	for {
		tries++

		conn, err := sqlx.Connect("mysql", dsn)
		if err != nil {
			if tries >= 5 {
				return nil, err
			}

			fmt.Println("Can't connect to the DB. Error: ", err, ". Try #", tries, "...")

			// Wait a bit and try again
			time.Sleep(time.Second * time.Duration(2^tries))

			continue
		}
		return conn, nil
	}
}
