package register

import (
	"database/sql"
	"fmt"
	"log"
	"test/MetaData"
)

//用于注册用户
const (
	salt = "asdjpqw12334askdn"
)

func SignUpOneUser(db *sql.DB, userName, userPwd string) bool {
	stmt, _ := db.Prepare("insert into tbl_user (`user_name`,`user_pwd`)  values (?,?)")
	defer stmt.Close()
	ret, err := stmt.Exec(userName, userPwd)
	if err != nil {
		log.Println(err)
		return false
	}
	fmt.Println(ret)
	return true
}

//验证密码是否正确
func CheckUserInfor(db *sql.DB, userName, pwd string) bool {
	stmt, err := db.Prepare("select user_pwd from tbl_user where user_name=? and user_pwd=?")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return false
	}
	rows, err := stmt.Query(userName, pwd)
	if err != nil {
		log.Println(err)
		return false
	}

	if rows.Next() {
		return true
	}
	return false
}

//验证用户名是否可用
func UserNameVaild(db *sql.DB, userName string) bool {
	stmt, err := db.Prepare("select user_pwd from tbl_user where user_name=?")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return false
	}
	rows, err := stmt.Query(userName)
	if err != nil {
		log.Println(err)
		return false
	}
	if rows.Next() {
		return true
	}
	return false
}

//删除一个用户信息
func DeleteOneUser(db *sql.DB, userName string) bool {
	stmt, err := db.Prepare("delete from tbl_user where user_name=?")
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

//获取用户信息
func GetUserInfor(db *sql.DB, userName string) *MetaData.User {
	stmt, err := db.Prepare("select `user_name`,`email`,`phone`,`signup_at` from tbl_user where user_name=?")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return nil
	}
	rows, err := stmt.Query(userName)
	if rows.Next() == false {
		return nil
	}
	userMeta := MetaData.User{}
	err = rows.Scan(&userMeta.Username, &userMeta.Email, &userMeta.Phone, &userMeta.SignupAt)
	if err != nil {
		log.Println("query failed")
		return nil
	}
	return &userMeta
}
