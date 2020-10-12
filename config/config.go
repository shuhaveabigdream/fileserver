package config

const (
	//buketName
	OSSBucket = "shuqi-store"
	//EndPoint
	OSSEndpoint = "xxxxxx"
	//AccessKey
	OSSAccessKeyID = "xxxxxxxxxxxxxxx"
	//AccessScrete
	OSSAccessKeySecret = "xxxxxxxxxxxxxx"
	//obj Name
	ObjName = "oss/"
	//rabbitmq Params

	//是否开启文件异步传输(默认情况是同步)
	AsyncTransferEnable = true
	RabbitmqURL         = "amqp://admin:admin@192.168.198.128:5672"
	//交换机名
	TransExchangeName = "uploadserver.trans"
	//队列名
	TransOSSQueueName = "uploadserver.trans.oss"
	//失败后转移
	TransOSSErrQueueName = "uploadserver.trans.oss.err"
	//routin key
	TransOSSRoutingKey = "oss"
)
