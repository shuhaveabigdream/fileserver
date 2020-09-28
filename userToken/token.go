package userToken

import (
	"database/sql"
	"log"
)

//更新一条token数据
func UpdateOneToken(db *sql.DB, userName, token string) bool {
	//使用replace  每次更新上一次的token数据作废
	stmt, err := db.Prepare("replace into tbl_user_token (user_name,user_token) values(?,?)")
	if err != nil {
		log.Println(err)
		return false
	}
	_, err = stmt.Exec(userName, token)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

//通过token获取用户名信息
func GetUserName(db *sql.DB, userToken string) (string, bool) {
	stmt, err := db.Prepare("select user_name from tbl_user_token where user_token=?")
	if err != nil {
		log.Println(err)
		return "", false
	}
	rows, err := stmt.Query(userToken)
	if err != nil {
		log.Println(err)
		return "", false
	}
	userName := ""
	rows.Next()
	rows.Scan(&userName)
	if userName == "" {
		return "", false
	}
	return userName, true
}

//删除一条token数据
func DeleteOneToken(db *sql.DB, userName string) bool {
	stmt, err := db.Prepare("delete from tbl_user_token where user_name=?")
	if err != nil {
		log.Println(err)
		return false
	}
	_, err = stmt.Exec(userName)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
