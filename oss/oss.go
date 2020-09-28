package oss

import (
	"fmt"
	"log"
	cfg "test/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

//该文件作用于对阿里云进行远程传输
//
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
