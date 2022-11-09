/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wangyi/GinTemplate/common"
	"github.com/wangyi/GinTemplate/dao/mysql"
	"github.com/wangyi/GinTemplate/dao/redis"
	"github.com/wangyi/GinTemplate/logger"
	"github.com/wangyi/GinTemplate/process"
	"github.com/wangyi/GinTemplate/router"
	"github.com/wangyi/GinTemplate/setting"
	"github.com/wangyi/GinTemplate/tools"
	"github.com/zh-five/xdaemon"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	port     string
	daemon   bool
	rootPath string
)
var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "启动服务",
	Example: "go-fly server",
	Run:     run,
}

func init() {
	serverCmd.PersistentFlags().StringVarP(&rootPath, "rootPath", "r", "", "程序根目录")
	serverCmd.PersistentFlags().StringVarP(&port, "port", "p", "8787", "监听端口号")
	serverCmd.PersistentFlags().BoolVarP(&daemon, "daemon", "d", false, "是否为守护进程模式")
}

func run(cmd *cobra.Command, args []string) {
	//初始化目录
	initDir()
	//初始化守护进程
	initDaemon()
	if noExist, _ := tools.IsFileNotExist(common.LogDirPath); noExist {
		if err := os.MkdirAll(common.LogDirPath, 0777); err != nil {
			log.Println(err.Error())
		}
	}
	isMainUploadExist, _ := tools.IsFileExist(common.Upload)
	if !isMainUploadExist {
		err := os.MkdirAll(common.Upload, os.ModePerm)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	// UploadGoods  商品图片地址
	isMainUploadGoodsExist, _ := tools.IsFileExist(common.UploadGoods)
	if !isMainUploadGoodsExist {
		err := os.MkdirAll(common.UploadGoods, os.ModePerm)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	//任务弹窗图片地址
	isMainUploadUploadTaskExist, _ := tools.IsFileExist(common.UploadTask)
	if !isMainUploadUploadTaskExist {
		err := os.MkdirAll(common.UploadTask, os.ModePerm)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	//加载配置
	if err := setting.Init(); err != nil {
		fmt.Println("配置文件初始化事变", err)
		return
	}
	//初始化日志
	if err := logger.Init(); err != nil {
		fmt.Println("日志初始化失败", err)
		return
	}
	defer zap.L().Sync() //缓存日志追加到日志文件中
	//链接数据库
	if err := mysql.Init(); err != nil {
		fmt.Println("mysql 链接失败,", err)
		return
	}
	defer mysql.Close()
	//redis 初始化
	//4.初始化redis连接
	if err := redis.Init(); err != nil {
		fmt.Println("redis文件初始化失败：", err)
		return
	}
	defer redis.Close()

	//进程运行
	//支付任务订单超时检查
	go process.OrderTimeout(mysql.DB)
	//充值订单超时检查
	go process.RechargeTimeout(mysql.DB)
	//定时更新 每日日数据
	go process.UpdateStatistics(mysql.DB)
	//每日任务
	go process.TimeTask(mysql.DB)
	router.Setup()
}

//初始化目录
func initDir() {

	if rootPath == "" {
		rootPath = tools.GetRootPath()
	}
	log.Println("程序运行路径:" + rootPath)
	common.RootPath = rootPath
	common.LogDirPath = rootPath + "/logs/"
	common.ConfigDirPath = rootPath + "/config/"
	common.StaticDirPath = rootPath + "/static/"
	common.UploadDirPath = rootPath + "/static/upload/"

}

//初始化守护进程
func initDaemon() {

	//启动进程之前要先杀死之前的金额

	pid, err := ioutil.ReadFile("Project.sock")
	if err != nil {
		return
	}
	pidSlice := strings.Split(string(pid), ",")
	var command *exec.Cmd
	for _, pid := range pidSlice {
		if runtime.GOOS == "windows" {
			command = exec.Command("taskkill.exe", "/f", "/pid", pid)
		} else {
			fmt.Println("成功结束进程:", pid)
			command = exec.Command("kill", pid)
		}
		command.Start()
	}

	if daemon == true {
		d := xdaemon.NewDaemon(common.LogDirPath + "Project.log")
		d.MaxError = 10
		d.Run()
	}
	//记录pid
	err = ioutil.WriteFile(common.RootPath+"/Project.sock", []byte(fmt.Sprintf("%d,%d", os.Getppid(), os.Getpid())), 0666)
	if err != nil {

		return
	}
}
