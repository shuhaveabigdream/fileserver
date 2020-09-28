package fileOperation

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"test/MetaData"
	"test/redis"
	"time"
)

var DownloadUrl string

func init() {
	DownloadUrl, _ = filepath.Abs("../downloadfile/")
}

//sha1码生成
func Sha1Generate(content []byte) []byte {
	encoder := sha1.New()
	encoder.Write(content)
	return encoder.Sum(nil)
}

func Sha1String(content string) []byte {
	encoder := sha1.New()
	io.WriteString(encoder, content)
	return encoder.Sum(nil)
}

func FileSha1(file *os.File) string {
	encoder := sha1.New()
	io.Copy(encoder, file)
	return fmt.Sprintf("%x", encoder.Sum(nil))
}

//生成md5码
func Md5String(content string) [16]byte {
	encoder := md5.New()
	io.WriteString(encoder, content)
	return md5.Sum(nil)
}

func NowTime() string {
	return time.Now().Format("2006-01-02")
}

//一次性读取文件所有数据
func ReadsigleFile(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return nil

	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		return nil
	}
	return content
}

func TransFrom(mp map[string]string) *MetaData.FileChunks {
	if len(mp) == 0 {
		return nil
	}
	CurBatch, _ := strconv.Atoi(mp["CurBatch"])
	BatchSize, _ := strconv.Atoi(mp["BatchSize"])
	Batch, _ := strconv.Atoi(mp["batch"])
	Size, _ := strconv.Atoi(mp["size"])
	ans := MetaData.FileChunks{
		CurBatch:  CurBatch,
		BatchSize: int64(BatchSize),
		Batch:     int64(Batch),
		Size:      int64(Size),
		Sha1Code:  mp["Sha1Code"],
	}
	return &ans
}

//分块读取
//这里还是需要使用到redis,缓存读取的相关数据
func ReadChuckFile(path string, filesha1 string, size int64) []byte {
	data := TransFrom(*redis.GetChuckMem(filesha1))

	f, err := os.Open(path)
	if err != nil {
		log.Println("file can't open", err)
		return nil
	}

	if data == nil {
		f.Seek(0, 0)
		infor, _ := os.Stat(path)
		fileSize := float64(infor.Size())
		floatSize := float64(size)
		ans := fileSize / floatSize
		fmt.Println(infor)
		data = &MetaData.FileChunks{
			Sha1Code: filesha1,
			Size:     infor.Size(),
			//to do：获取批次数，需要知道文件大小
			Batch:     int64(math.Ceil(ans)),
			BatchSize: size,
			CurBatch:  0,
		}
	} else {
		data.Sha1Code = filesha1
		//完成读取，已经没有剩余数据
		if int64(data.CurBatch) == data.Batch {
			return nil
		}
		f.Seek(int64(data.CurBatch)*data.BatchSize, 0)
	}
	if err != nil {
		log.Println(err)
		return nil
	}
	scaner := bufio.NewReader(f)
	buf := make([]byte, size)
	n, err := scaner.Read(buf)
	if err != nil {
		log.Println(err)
		return nil
	}
	//刷新redis数据
	data.CurBatch++
	redis.UpdateChuckMem(data)
	return buf[:n]
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

//默认追加模式
func CopyChunks(path string, content []byte) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	f.Write(content)
}

//分块上传情况下记录文件
//params:
//sha1Code:be used to FileName
//idx:index of chunkcs
//content:the file data
func WriteMpFile(sha1Code string, idx int, content []byte) bool {
	path := DownloadUrl + "/" + sha1Code
	if flag, _ := PathExists(path); !flag {
		os.Mkdir(path, os.ModePerm)
	}
	f, err := os.OpenFile(path+"/"+strconv.Itoa(idx), os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
		return false
	}
	defer f.Close()
	f.Write(content)
	return true
}

// func main() {
// 	filePath := "C:/Users/root/Desktop/算法图解.pdf"
// 	sha1Code := "81f7eb62bd1686b16e1733f45951f4cd7daafacd"
// 	//tmpPath := "./test.pdf"
// 	total := 0
// 	i := 0
// 	for {
// 		if res := ReadChuckFile(filePath, sha1Code, 512*1024); res != nil {
// 			fmt.Println("read bytes:", len(res))
// 			//CopyChunks(tmpPath, res)
// 			WriteMpFile(sha1Code, i, res)
// 			i++
// 			total += len(res)
// 		} else {
// 			break
// 		}
// 	}

// 	fmt.Println("totally recived:", total)
// }
