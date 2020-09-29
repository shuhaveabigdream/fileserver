package handlers

/*************该文件包含了一个简单的文件传输系统****************/
//实现功能
//1.用户的注册，登录，删除
//2.文件的上传，下载，删除，改名
//3.文件信息查询
//4.用户信息查询

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"test/MetaData"
	"test/config"
	fop "test/fileOperation"
	"test/mpupload"
	"test/msgQue"
	"test/mysql"
	"test/oss"
	"test/redis"
	"test/register"
	"test/transfor"
	"test/userToken"
	"test/userfile"
	"time"

	redislib "github.com/garyburd/redigo/redis"
)

//为了和前端匹配，修改相关结构
type CallBack struct {
	Status int         `json:"code"`
	Msg    string      `json:"msg"`
	Body   interface{} `json:"data"`
}

//状态码0为ok -1为异常

func LoadJson(p *CallBack) []byte {
	data, err := json.Marshal(p)
	if err != nil {
		log.Println(err)
		return nil
	}
	return data
}

const (
	salt       = "zhoukkkk1234sq"
	token_salt = "poipozxcsjlkqweuoasdj"
)

func GenToken(userName string) string {
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := fmt.Sprintf("%x", fop.Md5String(token_salt+ts+userName))
	return tokenPrefix + ts[:8]
}

//注意返回结构体的属性名必须大写首字母否则访问不到
func AccountSignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			fmt.Println("page file not found!", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write(data)
		return
	}
	r.ParseForm()
	userName := r.Form.Get("username")
	passWd := r.Form.Get("password")

	//获取数据库句柄
	db := mysql.DBConn()
	if register.UserNameVaild(db, userName) {
		w.Write(LoadJson(&CallBack{
			-1,
			"username invalid",
			nil,
		}))
		return
	}

	pwdHash := fmt.Sprintf("%x", fop.Sha1String((passWd + salt + passWd + userName)))
	suc := register.SignUpOneUser(db, userName, pwdHash)
	if !suc {
		w.Write(LoadJson(&CallBack{
			-1,
			"register failed",
			nil,
		}))
		return
	}
	//TO DO:加密生成token
	tokenprefix := GenToken(userName)
	//将token插入
	userToken.UpdateOneToken(db, userName, tokenprefix)
	w.Write([]byte("SUCCESS"))
}

//原接口中signin居然是明文
func AccountSignIn(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	return
	// }
	r.ParseForm()
	userName := r.Form.Get("username")
	userPwd := r.Form.Get("password")

	//结构修改后，需要先获取db，再进行数据库相关操作
	db := mysql.DBConn()
	pwdHash := fmt.Sprintf("%x", fop.Sha1String((userPwd + salt + userPwd + userName)))

	if register.CheckUserInfor(db, userName, pwdHash) {
		//登录成功
		//1.替换mysql中token
		tokenPerfix := GenToken(userName)
		userToken.UpdateOneToken(db, userName, tokenPerfix)
		//2.返回token
		w.Write(LoadJson(&CallBack{
			0,
			"OK",
			struct {
				Location string
				Username string
				Token    string
			}{

				"http://" + r.Host + "/static/view/home.html",
				userName,
				tokenPerfix,
			},
		}))
	} else {
		//登录失败
		w.Write([]byte("Failed"))
		return
	}
}

func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	r.ParseForm()
	userName := r.PostForm.Get("username")
	db := mysql.DBConn()
	scc := register.DeleteOneUser(db, userName)
	suc := userToken.DeleteOneToken(db, userName)
	if suc && scc {
		//删除成功
		w.Write(LoadJson(&CallBack{
			0,
			"OK",
			nil,
		}))
	} else {
		w.Write(LoadJson(&CallBack{
			-1,
			"Failed",
			nil,
		}))
	}
}

//鉴权接口
func InterAction(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			username := ""
			token := ""

			if username = r.Form.Get("username"); len(username) == 0 {
				if username = r.PostForm.Get("username"); len(username) == 0 {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			if token = r.Form.Get("token"); len(token) == 0 {
				if username = r.PostForm.Get("token"); len(username) == 0 {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}

			db := mysql.DBConn()
			if name, ok := userToken.GetUserName(db, token); ok && name == username {
				f(w, r)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return
		})
}

//上传接口
//TODO 上传接口需要大改
//对于基本上传接口，在本地存储后同步的上传到阿里云
//对于分块上传接口，在本地存储后，在阿里云追加写入
//对于下载，云端而言，只需要下拉一个url即可
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//get则读取html文件返回
		//这里返回静态文件即可
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internet sever error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		//检验用户数据
		db := mysql.DBConn()
		token := r.Form.Get("token")
		userName, ok := userToken.GetUserName(db, token)
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		file, head, err := r.FormFile("file")
		//获取关键字file&head
		if err != nil {
			fmt.Printf("Failed to get data,err:%s\n", err.Error())
			return
		}
		defer file.Close()
		lc, _ := filepath.Abs("../downloadfile/" + head.Filename)

		//插入文件表所需的数据
		fileMeta := MetaData.FileMeta{
			FileName: head.Filename,
			Location: lc,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		//创建文件
		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			log.Println(err)
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)

		if err != nil {
			log.Println(err)
			return
		}

		newFile.Seek(0, 0) //恢复位置

		fileMeta.FileSha1 = fop.FileSha1(newFile)

		//插入用户文件表所需要的数据
		userFileMeta := MetaData.UserFile{
			UserName: userName,
			FileName: head.Filename,
			FileSize: fileMeta.FileSize,
			FileSha1: fileMeta.FileSha1,
		}

		userfile.InsertSigleRecored(db, &userFileMeta)

		suc := transfor.UpdateFileTbl(db, &fileMeta)
		if suc {
			//不再等待云操作，而是发出消息
			data := msgQue.TransferData{
				FileHash:     fileMeta.FileSha1,
				DestLocation: config.ObjName + fileMeta.FileSha1,
				CurLocation:  fileMeta.Location,
			}
			//oss.OssUploadFileStream(config.ObjName+fileMeta.FileSha1, fileMeta.Location)
			//fileMeta.Location = config.ObjName + fileMeta.FileSha1
			//数据插入
			//transfor.InsertSingleOssRecord(db, fileMeta)

			//将data数据发送到消息队列
			pubData, _ := json.Marshal(data)
			suc := msgQue.Publish(
				config.TransExchangeName,
				config.TransOSSRoutingKey,
				pubData)
			if !suc {
				log.Println("消息未能成功发送")
				//TO DO重新发送
			}

			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

//获取用户信息
func UserInforHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	db := mysql.DBConn()
	username := r.Form.Get("username")
	user := register.GetUserInfor(db, username)
	if user == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Write(LoadJson(&CallBack{
		0,
		"ok",
		user,
	}))
}

//获取文件列表(重要接口)
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	db := mysql.DBConn()
	UserFiles, err := userfile.QueryUserFileMetas(db, username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//是一个字典数组
	data, err := json.Marshal(UserFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

//普通文件下载方案
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	db := mysql.DBConn()
	fileMeta, ok := transfor.QueryFileTbl(db, filehash)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	buf := fop.ReadsigleFile(fileMeta.Location)
	if buf == nil {

		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.Header().Set("content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment;filename=\""+fileMeta.FileName+"\"")
	w.Write(buf)
}

func FastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	filesha1 := r.Form.Get("filehash")

	db := mysql.DBConn()
	suc := mpupload.FastUpload(db, username, filesha1)
	if suc {
		w.Write(LoadJson(&CallBack{
			0,
			"success",
			nil,
		}))
	} else {
		w.Write(LoadJson(&CallBack{
			-1,
			"failed",
			nil,
		}))
	}
}

//分块上传相关接口
//1.初始化分块上传
func UploadChunksInit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusForbidden)
	}
	//2.获取redis链接
	rConn := redis.GetPool().Get()
	//3.生成分块初始化信息
	upinfo := MetaData.MultipartUploadinfo{
		FileHash:  filehash,
		FileSize:  filesize,
		UploadID:  username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize: 5 * 1024 * 1024,
		//向上取整
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}
	//4.将初始化信息写入redis缓存
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "chunkcount", upinfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filehash", upinfo.FileHash)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filesize", upinfo.FileSize)

	//5.初始化信息返回到客户端
	w.Write(LoadJson(&CallBack{
		0,
		"success",
		upinfo.UploadID,
	}))
}

//2.开始分块上传
func UploadSingleChunk(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")
	filehash := r.Form.Get("filehash")
	index, _ := strconv.Atoi(chunkIndex)

	fp, _, _ := r.FormFile("file")
	buffer := make([]byte, 1024*1024*5)

	for {
		n, _ := fp.Read(buffer)
		if n == 0 {
			break
		}
		fop.WriteMpFile(filehash, index, buffer[:n])
	}
	rConn := redis.GetPool().Get()
	defer rConn.Close()
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
	w.Write(LoadJson(&CallBack{
		0,
		"success",
		nil,
	}))
}

//3.通知上传完成
func UploadComplete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	upid := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	rConn := redis.GetPool().Get()
	defer rConn.Close()

	val, err := redislib.Values(rConn.Do("HGETALL", "MP_"+upid))
	if err != nil {
		w.Write(LoadJson(&CallBack{
			-1,
			"failed",
			nil,
		}))
		return
	}
	totalCount := 0
	chunkCount := 0

	for i := 0; i < len(val); i += 2 {
		k := string(val[i].([]byte))
		v := string(val[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}

	if totalCount != chunkCount {
		w.Write(LoadJson(&CallBack{
			-1,
			"failed",
			nil,
		}))
		return
	}
	//4.TODO:合并分块
	//5.更新文件表和用户文件表
	db := mysql.DBConn()
	fsize, _ := strconv.Atoi(filesize)
	fm := MetaData.FileMeta{
		FileSha1: filehash,
		FileName: filename,
		FileSize: int64(fsize),
		Location: fop.DownloadUrl + "\\" + filehash,
	}

	uf := MetaData.UserFile{
		UserName: username,
		FileSha1: filehash,
		FileName: filename,
		FileSize: int64(fsize),
	}
	transfor.UpdateFileTbl(db, &fm)
	userfile.InsertSigleRecored(db, &uf)
	//6.相应结果
	w.Write(LoadJson(&CallBack{
		1,
		"upload ok",
		nil,
	}))
}

func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return fi.IsDir()
}

//将下载信息传入到redis
func InitChucksDownload(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form.Get("filehash")
	username := r.Form.Get("username")

	db := mysql.DBConn()
	suc := userfile.FindSingleFile(db, fileHash, username)
	if !suc {
		w.Write(LoadJson(&CallBack{
			-1,
			"无权访问",
			nil,
		}))
		return
	}

	fMeta, _ := transfor.QueryFileTbl(db, fileHash)
	fileSize := strconv.FormatInt(fMeta.FileSize, 10)

	if IsFile(fMeta.Location) == false {
		w.Write(LoadJson(&CallBack{
			-1,
			"文件不存在，或不支持分块",
			nil,
		}))
		return
	}
	//TO DO:redis有效时限问题
	infor := map[string]string{
		"filehash": fileHash,
		"filename": fMeta.FileName,
		"filesize": fileSize,
		"location": fMeta.Location,
	}

	uploadId := fop.Md5String(username + time.Now().String())

	redis.WriteMap(infor, fmt.Sprintf("%x", uploadId[:]))
	w.Write(LoadJson(&CallBack{
		0,
		"success",
		fmt.Sprintf("%x", uploadId[:]),
	}))
	return
}

//分块下载
func DownloadChunksHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	uploadId := r.Form.Get("dwloadid")
	idx := r.Form.Get("index")
	rConn := redis.GetPool().Get()
	body, err := rConn.Do("HGETALL", uploadId)
	if err != nil {
		w.Write(LoadJson(&CallBack{
			-1,
			"未找到授权信息",
			nil,
		}))
		return
	}

	data := redis.Release(body.([]interface{}))
	fdata := fop.ReadsigleFile(data["location"] + "/" + idx)
	w.Write(fdata)
	return
}

//分块下载完成,回执以删除redis中相关内容
func ChunkDownloadCmp(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	uploadId := r.Form.Get("uploadid")
	rConn := redis.GetPool().Get()
	rConn.Do("DEL", uploadId)
	w.Write(LoadJson(&CallBack{
		0,
		"success",
		nil,
	}))
}

func DownloadUrlHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	db := mysql.DBConn()
	suc := userfile.FindSingleFile(db, fileHash, username)
	if !suc {
		w.Write(LoadJson(&CallBack{
			-1,
			"无权访问",
			nil,
		}))
		return
	}
	objname := transfor.GetObjNameFromDb(db, fileHash)
	if len(objname) == 0 {
		w.Write(LoadJson(&CallBack{
			-1,
			"未在云上发现",
			nil,
		}))
		return
	}
	url := oss.OssDownloadUrl(objname)
	w.Write([]byte(url))
}
