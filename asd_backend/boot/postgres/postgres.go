package postgres

import (
	"asd/conf"
	"asd/conf/config"
	"fmt"
	"time"

	_ "asd/app/models"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/lib/pq"
)

// 注册PostgreSql
func init() {
	// 注册数据库驱动
	err := orm.RegisterDriver("postgres", orm.DRPostgres)
	if err != nil {
		beego.Error("poostgres register driver error:", err)
	}

	// 注册默认数据库
	registerDatabase("default", conf.CONFIG.Postgres.Default)
}

// 注册数据库的通用方法
func registerDatabase(aliasName string, dbConfig config.DBConfig) {
	dataSource := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		//dataSource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Asia%%2FShanghai&sql_mode=''",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database,
	)
	fmt.Println(dataSource)
	// 注册数据库
	err := orm.RegisterDataBase(aliasName, "postgres", dataSource)
	if err != nil {
		beego.Error(fmt.Sprintf("%s database register error:", aliasName), err)
	}

	db, err := orm.GetDB(aliasName)
	if err != nil {
		beego.Error(fmt.Sprintf("%s get database error:", aliasName), err)
	}

	// 设置客户端超时时长小于服务端的 8 小时
	db.SetConnMaxLifetime(2 * time.Hour)

	// Debug 模式下开启 ORM 调试，打印生成的SQL语句
	if beego.BConfig.RunMode == "dev" {
		fmt.Printf("%s database debug mode\n", aliasName)
		orm.Debug = true
	}

	err = orm.RunSyncdb(aliasName, false, orm.Debug)
	if err != nil {
		beego.Error(fmt.Sprintf("%s sync database error:", aliasName), err)
	}
}
