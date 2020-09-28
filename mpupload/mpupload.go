package mpupload

import (
	"database/sql"
	"test/MetaData"
	"test/transfor"
	"test/userfile"
)

/**********该文件在base服务器的基础上添加秒传和分块上传与下载的服务*************/

//快传
//返回秒传成功与否
func FastUpload(db *sql.DB, username, filesha1 string) bool {
	fileMeta, ok := transfor.QueryFileTbl(db, filesha1)
	if !ok {
		return false
	}
	tmp := MetaData.UserFile{
		UserName: username,
		FileSha1: fileMeta.FileSha1,
		FileSize: fileMeta.FileSize,
		FileName: fileMeta.FileName,
	}

	suc := userfile.InsertSigleRecored(db, &tmp)

	if suc {
		return true
	}
	return false
}

