package redis

import (
	"log"
	"test/MetaData"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	redisHost = "127.0.0.1:6379"
	passWord  = "Shu@123456"
)

var pool *redis.Pool

func conn() {
	pool = &redis.Pool{
		MaxIdle:     50,
		MaxActive:   30,
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			//获取链接
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				log.Println("redis err:", err)
				return nil, err
			}
			//访问认证
			_, err = c.Do("AUTH", passWord)
			if err != nil {
				log.Println("redis err:", err)
				return nil, err
			}
			return c, nil
		},
		//定期检查链接状况
		TestOnBorrow: func(connect redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := connect.Do("Ping")
			return err
		},
	}
}

func init() {
	conn()
}

//解析get后的数据成为map
func Release(t []interface{}) map[string]string {
	buf := make(map[string]string)
	for i := 0; i < len(t); i += 2 {
		key := string(t[i].([]uint8))
		val := string(t[i+1].([]uint8))
		buf[key] = val
	}
	return buf
}

func GetConn() *redis.Pool {
	return pool
}

// type FileChunks struct {
// 	size      int64
// 	batch     int64
// 	batchSize int64
// 	curBatch  int
//  sha1Code string
// }

//逻辑实现分块读取文件
func UpdateChuckMem(t *MetaData.FileChunks) bool {
	rConn := pool.Get()
	defer rConn.Close()
	_, err := rConn.Do("HMSET", t.Sha1Code, "size", t.Size, "batch", t.Batch, "BatchSize", t.BatchSize, "CurBatch", t.CurBatch)
	if err != nil {
		log.Println("InitChuck err:", err)
		return false
	}
	return true
}

//获取分块读取数据
func GetChuckMem(fileSha1 string) *map[string]string {
	rConn := pool.Get()
	defer rConn.Close()
	data, err := rConn.Do("HGETALL", fileSha1)
	if err != nil {
		log.Println("GetChuckMem err", err)
		return nil
	}

	mp := Release(data.([]interface{}))
	return &mp
}

func GetPool() *redis.Pool {
	return pool
}

func WriteMap(mp map[string]string, id string) {
	rConn := pool.Get()

	for k, v := range mp {
		rConn.Do("HSET", id, k, v)
	}
}
