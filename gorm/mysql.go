package gorm

import (
	"database/sql"
	"fmt"
	"nas_server/logs"

	_ "github.com/go-sql-driver/mysql"
)

var client *sql.DB

func MysqlInit(sqlName, sqlPassword, addr, dbName string) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s", sqlName, sqlPassword, addr, dbName)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		logs.GetInstance().Logger.Errorf("init mysql error %s", err)
	}
	client = db
}

func GetClient() *sql.DB {
	return client
}