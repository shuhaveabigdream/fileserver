package oss

import (
	"fmt"
	"log"
	"os"
	"strconv"
	cfg "test/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

//该文件作用于对阿里云进行远程传输
//在不改变原数据库的结构的情况下需要设置一个oss表
//filename string
//filehash string
//filesize int
//objname string

//其中关键数据在于objname，需要通过filehash来获取objname
var ossCli *oss.Client

//获取服务器
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}
	ossCli, err := oss.New(cfg.OSSEndpoint, cfg.OSSAccessKeyID, cfg.OSSAccessKeySecret)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return ossCli
}

//获取桶子
func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(cfg.OSSBucket)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return bucket
	}
	fmt.Println("Init first")
	return nil
}

//获取下载链接
func OssDownloadUrl(objName string) string {
	signedUrl, err := Bucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err.Error)
		return ""
	}
	return signedUrl
}

//上传文件流
func OssUploadFileStream(objname, filepath string) bool {
	err := Bucket().PutObjectFromFile(objname, filepath)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//分块上传逻辑封装
func OssUploadByChunks(ObjName, path string) bool {
	bucket := Bucket()
	//采用追加模式
	idx := 0
	var cur int64 = 0
	//分块上传
	for {
		curPath := path + "/" + strconv.Itoa(idx)
		flag, _ := PathExists(curPath)
		if flag {
			f, _ := os.Open(curPath)
			// finfor, _ := os.Stat(curPath)
			// size := finfor.Size()
			ep, err := bucket.AppendObject(ObjName, f, cur)
			cur = ep
			if err != nil {
				log.Println(err)
				return false
			}
			idx++
			f.Close()
		} else {
			break
		}
	}
	return true
}
