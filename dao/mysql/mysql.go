/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
	"github.com/wangyi/GinTemplate/model"
	"time"
)

var (
	DB  *gorm.DB
	err error
)

func Init() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		viper.GetString("mysql.dbname"),
	)
	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		fmt.Println("数据库链接失败", err)
		panic(err)
		return err
	}

	//设置连接池
	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetMaxOpenConns(100)

	////////////////////////////////////////////////////////////////////////模型初始化
	model.CheckIsExistModelUser(DB)
	model.CheckIsExistModelLog(DB)
	model.CheckIsExistModelConfig(DB)
	model.CheckIsExistModelStatistics(DB)
	model.CheckIsExistModelVip(DB)
	model.CheckIsExistModelMenu(DB)  //create menu
	model.CheckIsExistModelRole(DB)  //create role
	model.CheckIsExistModelAdmin(DB) //create  administrator
	model.CheckIsExistModelSlideshow(DB)
	model.CheckIsExistModelCountry(DB)
	model.CheckIsExistModelGoods(DB)
	model.CheckIsExistModelTask(DB)
	model.CheckIsExistGetTask(DB)
	model.CheckIsExistModelTaskOrder(DB)
	model.CheckIsExistModelAccountChange(DB)
	model.CheckIsExistModelRecord(DB)
	model.CheckIsExistModelBankCard(DB)
	model.CheckIsExistModelBankCardInformation(DB)
	model.CheckIsExistModelBankPay(DB)
	model.CheckIsExistModelPayChannels(DB)
	//设置时区
	config := model.Config{}
	err2 := DB.Where("id=?", 1).First(&config).Error
	if err2 == nil {
		loc, err := time.LoadLocation(config.TimeZone)
		if err == nil {
			time.Local = loc // -> this is setting the global timezone
		}

	}

	////////////////////////////////////////////////////////////////////////模型初始化
	return err
}

func Close() {
	defer DB.Close()
}
