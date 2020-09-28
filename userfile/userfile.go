package userfile

import (
	"database/sql"
	"fmt"
	"log"
	"test/MetaData"
)

//插入一条记录
func InsertSigleRecored(db *sql.DB, p *MetaData.UserFile) bool {
	stmt, err := db.Prepare("insert ignore into tbl_user_file(`user_name`,`file_sha1`,`file_name`,`file_size`) values(?,?,?,?)")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return false
	}
	_, err = stmt.Exec(p.UserName, p.FileSha1, p.FileName, p.FileSize)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

//删除一条记录
func DeleteSingleRecord(db *sql.DB, filehash string) bool {
	stmt, err := db.Prepare("delete from tbl_user_file where file_sha1=?")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return false
	}
	_, err = stmt.Exec(filehash)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

//修改文件名
func ChangeFileName(db *sql.DB, filehash, newName string) bool {
	stmt, err := db.Prepare("update tbl_user_file set `file_name`=? where file_sha1=?")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return false
	}
	_, err = stmt.Exec(filehash)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

//鉴权
func FindSingleFile(db *sql.DB, filehash, username string) bool {
	stmt, err := db.Prepare("select * from tbl_user_file where file_sha1=? and user_name=?")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return false
	}
	rows, err := stmt.Query(filehash, username)
	if err != nil {
		log.Println(err)
		return false
	}
	if rows.Next() {
		return true
	}
	log.Println("no perminate")
	return false
}

//获取文件列表
func QueryUserFileMetas(db *sql.DB, username string, limit int) ([]MetaData.FileMeta, error) {
	stmt, err := db.Prepare("select file_sha1,file_name,file_size,upload_at,last_update from tbl_user_file where user_name=? limit?")
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		return nil, err
	}

	var Userfiles []MetaData.FileMeta
	for rows.Next() {
		ufile := MetaData.FileMeta{}
		err := rows.Scan(&ufile.FileSha1, &ufile.FileName, &ufile.FileSize, &ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			fmt.Println(err)
			break
		}
		Userfiles = append(Userfiles, ufile)
	}
	return Userfiles, nil
}
