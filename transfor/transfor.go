package transfor

import (
	"database/sql"
	"fmt"
	"log"
	"test/MetaData"
)

func UpdateFileTbl(db *sql.DB, p *MetaData.FileMeta) bool {
	stmt, err := db.Prepare("insert ignore into tbl_file(file_sha1,file_name,file_size,file_addr) values (?,?,?,?)")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return false
	}
	_, err = stmt.Exec(p.FileSha1, p.FileName, p.FileSize, p.Location)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func QueryFileTbl(db *sql.DB, filehash string) (MetaData.FileMeta, bool) {
	ans := MetaData.FileMeta{}
	stmt, err := db.Prepare("select file_sha1,file_name,file_size,file_addr from tbl_file where `file_sha1`=?")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return ans, false
	}
	rows, err := stmt.Query(filehash)
	if err != nil {
		log.Println(err)
		return ans, false
	}
	if rows.Next() == false {
		return ans, false
	}
	err = rows.Scan(&ans.FileSha1, &ans.FileName, &ans.FileSize, &ans.Location)
	if err != nil {
		log.Println(err)
		return ans, false
	}
	return ans, true
}

//因为tbl_file和tbl_oss_file字段雷同较多，所以可以用相同的metadata进行操作
//仅需要将fileMeta中的Location定义为objname即可

func InsertSingleOssRecord(db *sql.DB, p MetaData.FileMeta) bool {
	stmt, err := db.Prepare("insert ignore into tbl_oss_file (file_sha1,file_name,file_size,objname)values(?,?,?,?)")
	if err != nil {
		log.Println(err)
		return false
	}
	_, err = stmt.Exec(p.FileSha1, p.FileName, p.FileSize, p.Location)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

//返回的是objname
func GetObjNameFromDb(db *sql.DB, filehash string) string {
	rows, err := db.Query(fmt.Sprintf(`select objname from tbl_oss_file where file_sha1="%s"`, filehash))
	if err != nil {
		log.Println(err)
		return ""
	}

	if rows.Next() == false {
		return ""
	}

	objname := ""

	rows.Scan(&objname)
	return objname
}
