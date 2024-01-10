package gorm

import (
	"database/sql"
	"fmt"
	config "nas_server/conf"
	"nas_server/logs"

	_ "github.com/go-sql-driver/mysql"
)

var client *sql.DB

func MysqlInit(mysqlConfig *config.MySQLConfig) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s", mysqlConfig.Name, mysqlConfig.PassWord, mysqlConfig.Addr, mysqlConfig.DB)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		logs.GetInstance().Logger.Errorf("init mysql error %s", err)
	}
	client = db
}

func GetClient() *sql.DB {
	return client
}