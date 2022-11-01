package common

import "sync"

var (
	PageSize                   uint         = 9
	VisitorPageSize            uint         = 10
	Version                    string       = "0.6.0"
	VisitorExpire              float64      = 600
	Upload                     string       = "static/upload/"
	UploadGoods                string       = "static/goods/"
	UploadTask                 string       = "static/taskDialog/"
	Dir                        string       = "config/"
	MysqlConf                  string       = Dir + "mysql.json"
	RpcServer                  string       = "0.0.0.0:8082"
	RpcStatus                  bool         = false
	SecretToken                string       = "AaBbCcDd123AaBbCcDd"
	AesKey                     string       = "ggcjd4slhtjyxl16"
	WsBreakTimeout             int64        = 5     //断线超时秒数
	IsCompireTemplate          bool         = false //是否编译静态模板到二进制
	IsTry                      bool         = false
	TryDeadline                int64        = 30 * 24 * 3600 //试用十四天
	WeixinTemplateHost         string       = "http://wechat.sopans.com/api/wechat/templateMessage"
	DomainWhiteList            string       = "*" //域名白名单
	RootPath                   string       = ""  //自动获取，程序运行根路径
	LogDirPath                 string       = ""  //自动获取，程序运行根路径/logs/
	ConfigDirPath              string       = ""  //自动获取，程序运行根路径/config/
	StaticDirPath              string       = ""  //自动获取，程序运行根路径/static/
	UploadDirPath              string       = ""  //自动获取，程序运行根路径/static/upload/
	LockForGlobalChangeBalance sync.RWMutex       //全局的余额操作锁(会影响性能,但是安全)
)
