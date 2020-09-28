package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"test/MetaData"

	_ "github.com/go-sql-driver/mysql"
)

//mysql字段逻辑
//|-----table "tbl_user" -----|
//id PRI key
//user_name 用户名
//user_pwd 密码经过sha1加密后
//email 邮箱可以为空
//email_validated 邮箱可用标志
//phone_validated 电话号码可用标记
//signup_at 注册时间(datetime)
//last_active 最后活跃时间
//profile 个人介绍
//status 状态码

//|---------table "tbl_file"--------|
//id pri key
//file_sha1
//file_name
//file_addr

//因为windows平台的原因，所有mysql逻辑代码都在该文件中
var (
	db *sql.DB
)

func init() {
	db, _ = sql.Open("mysql", "root:Shu@123456@tcp(127.0.0.1:3306)/fileserver?charset=utf8")
	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil {
		fmt.Println("Failed to connect to mysql ,err:" + err.Error())
		os.Exit(1)
	}
	fmt.Println("db interface ok!")
}

func DBConn() *sql.DB {
	return db
}

//移植时可以用于创建数据表
func Establish() {
	sql := `CREATE TABLE tbl_file (
		id int NOT NULL AUTO_INCREMENT,
		file_sha1 char(40) NOT NULL DEFAULT '' COMMENT 文件hash,
		file_name char(255) NOT NULL DEFAULT '' COMMENT 文件名,
		file_size bigint DEFAULT '0' COMMENT '文件大小',
		file_addr varchar(1024) NOT NULL DEFAULT '' COMMENT 文件存储位置,
		create_at datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
		update_at  datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
		status int NOT NULL DEFAULT '0' COMMENT '状态(可用/禁用/已删除等状态)',
		ext1 int DEFAULT '0' COMMENT '备用字段1',
		ext2 text COMMENT '备用字段2',
		PRIMARY KEY (id),
		UNIQUE KEY idx_file_hash (file_sha1),
		KEY idx_status (status)
	  ) ENGINE=InnoDB AUTO_INCREMENT=20 DEFAULT CHARSET=utf8`
	_, err := db.Exec(sql)
	if err != nil {
		log.Println(err)
		return
	}
}

func Query(sql string) {
	rows, err := db.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Println(err)
		return
	}
	callback := []MetaData.UserTable{}
	for rows.Next() {
		x := MetaData.UserTable{}
		err = rows.Scan(&x.User_name, &x.User_pwd, &x.Email)
		if err != nil {
			log.Println(err)
			return
		}
		callback = append(callback, x)
	}
	fmt.Println(callback)
}
