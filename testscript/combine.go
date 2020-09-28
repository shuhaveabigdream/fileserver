package main

import (
	"fmt"
	"log"
	"test/oss"
)

func upload(filepath string, filehash string) {
	suc := oss.OssUploadFileStream("oss/"+filehash, filepath)
	if suc != true {
		log.Println("上传失败")
		return
	}
}

func download(filehash string) {
	fmt.Println("url", oss.OssDownloadUrl("oss/"+filehash))
}

func main() {
	download("81f7eb62bd1686b16e1733f45951f4cd7daafacd")
}
