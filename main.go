package main

import (
	"log"
	"net/http"
	"test/handlers"
)

//当前该main文件用于测试各个模块的运行状态
func main() {
	//开服务器
	http.HandleFunc("/user/signup", handlers.AccountSignUp)
	http.HandleFunc("/user/signin", handlers.AccountSignIn)
	http.HandleFunc("/user/delete", handlers.InterAction(handlers.DeleteAccount))
	http.HandleFunc("/file/upload", handlers.InterAction(handlers.UploadHandler))
	http.HandleFunc("/file/query", handlers.InterAction(handlers.GetFileMetaHandler))
	http.HandleFunc("/user/info", handlers.InterAction(handlers.UserInforHandler))
	http.HandleFunc("/file/download", handlers.InterAction(handlers.DownloadHandler))

	//分块上传服务
	http.HandleFunc("/file/chunkinit", handlers.InterAction(handlers.UploadChunksInit))
	http.HandleFunc("/file/chunkupload", handlers.InterAction(handlers.UploadSingleChunk))
	http.HandleFunc("/file/chunkcmp", handlers.InterAction(handlers.UploadComplete))

	//分块下载服务
	http.HandleFunc("/file/mdownloadinit", handlers.InterAction(handlers.InitChucksDownload))
	http.HandleFunc("/file/mdownload", handlers.InterAction(handlers.DownloadChunksHandler))
	http.HandleFunc("/file/mdownloadcmp", handlers.InterAction(handlers.ChunkDownloadCmp))

	//TO DO:
	//秒传
	http.HandleFunc("file/fastupload", handlers.InterAction(handlers.FastUploadHandler))
	//上云

	//消息队列处理由本地到云

	//处理静态文件
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("启动失败")
	}
}
