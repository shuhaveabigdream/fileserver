# fileserver
Similar to Baidu network disk design
# quick start
## 1. DataBase establish
* DataBase Name:fileserver
* tbl_user:

``` 
tbl_user | CREATE TABLE `tbl_user` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_name` varchar(64) NOT NULL DEFAULT '' COMMENT '用户名',
  `user_pwd` varchar(256) NOT NULL DEFAULT '' COMMENT '用户encoded密码',
  `email` varchar(64) DEFAULT '' COMMENT '邮箱',
  `phone` varchar(128) DEFAULT '' COMMENT '手机号',
  `email_validated` tinyint(1) DEFAULT '0' COMMENT '邮箱是否已验证',
  `phone_validated` tinyint(1) DEFAULT '0' COMMENT '手机号是否已经注册过',
  `signup_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '注册日期',
  `last_active` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后活跃时间',
  `profile` text COMMENT '用户属性',
  `status` int NOT NULL DEFAULT '0' COMMENT '账户状态(启用/禁用/锁定/标记删除等)',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=30 DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;

```
* tbl_user_file
```
tbl_user_file | CREATE TABLE `tbl_user_file` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_name` varchar(64) NOT NULL,
  `file_sha1` varchar(64) NOT NULL DEFAULT '' COMMENT '文件hash',
  `file_name` varchar(256) NOT NULL DEFAULT '' COMMENT '文件名',
  `upload_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  `last_update` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `status` int NOT NULL DEFAULT '0' COMMENT '文件状态(0正常1已删除2禁用)',
  `file_size` bigint NOT NULL DEFAULT '0' COMMENT '文件大小',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_file` (`user_name`,`file_sha1`),
  KEY `idx_status` (`status`),
  KEY `idx_user_id` (`user_name`)
) ENGINE=InnoDB AUTO_INCREMENT=38 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
```

* tbl_user_token
```
tbl_user_token | CREATE TABLE `tbl_user_token` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_name` varchar(64) NOT NULL DEFAULT '' COMMENT '用户名',
  `user_token` char(40) NOT NULL DEFAULT '' COMMENT '用户登录token',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`user_name`)
) ENGINE=InnoDB AUTO_INCREMENT=57 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
```

* tbl_file
```
tbl_file | CREATE TABLE `tbl_file` (
  `id` int NOT NULL AUTO_INCREMENT,
  `file_sha1` char(40) NOT NULL DEFAULT '' COMMENT '文件hash',
  `file_name` char(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `file_size` bigint DEFAULT '0' COMMENT '文件大小',
  `file_addr` varchar(1024) NOT NULL DEFAULT '' COMMENT '文件存储位置',
  `create_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
  `update_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态(可用/禁用/已删除等状态)',
  `ext1` int DEFAULT '0' COMMENT '备用字段1',
  `ext2` text COMMENT '备用字段2',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_file_hash` (`file_sha1`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=44 DEFAULT CHARSET=utf8
```
* tbl_oss_file
```
tbl_oss_file | CREATE TABLE `tbl_oss_file` (
  `id` int NOT NULL AUTO_INCREMENT,
  `file_sha1` char(40) NOT NULL DEFAULT '' COMMENT '文件hash',
  `file_name` char(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `file_size` bigint DEFAULT '0' COMMENT '文件大小',
  `objname` varchar(1024) NOT NULL DEFAULT '' COMMENT '文件存储位置',
  `create_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
  `update_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态(可用/禁用/已删除等状态)',
  `ext1` int DEFAULT '0' COMMENT '备用字段1',
  `ext2` text COMMENT '备用字段2',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_file_hash` (`file_sha1`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=34 DEFAULT CHARSET=utf8
```
## 2.Basic Config
```
const (
	//buketName
	OSSBucket = ""
	//EndPoint
	OSSEndpoint = ""
	//AccessKey
	OSSAccessKeyID = ""
	//AccessScrete
	OSSAccessKeySecret = ""
	//obj Name
	ObjName = "oss/"
	//rabbitmq Params

	//是否开启文件异步传输(默认情况是同步)
	AsyncTransferEnable = true
	RabbitmqURL         = ""
	//交换机名
	TransExchangeName = ""
	//队列名
	TransOSSQueueName = ""
	//失败后转移
	TransOSSErrQueueName = ""
	//routin key
	TransOSSRoutingKey = ""
)

then modify the auth of redis and mysql
```


## 3.Run Server
```
go run main.go//the basic server start up
```

## 4.Run Rabbitmq
```
go run msgQue/main/main.go
```
