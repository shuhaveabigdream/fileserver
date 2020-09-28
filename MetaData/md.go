package MetaData

//将该结构体数据存入到redis
type FileChunks struct {
	Size      int64
	Batch     int64
	BatchSize int64
	CurBatch  int
	Sha1Code  string
}

type UserTable struct {
	User_name string
	User_pwd  string
	Email     string
}

type FileMeta struct {
	FileSha1    string `json:"FileHash"`
	FileName    string
	FileSize    int64
	Location    string
	UploadAt    string
	LastUpdated string
}

type UserFile struct {
	UserName string
	FileSha1 string
	FileName string
	FileSize int64
}

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

type MultipartUploadinfo struct {
	FileHash  string
	FileSize  int
	UploadID  string
	ChunkSize int
	//向上取整
	ChunkCount int
}
