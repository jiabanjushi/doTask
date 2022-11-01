package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
	"sync"
	"testing"
)

func TestCheckIsExistModelTask(t *testing.T) {
	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name string
		args args
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CheckIsExistModelTask(tt.args.db)
		})
	}
}

func TestTask_CreateNoOverLayTask(t1 *testing.T) {

	str := "123 456 8977"
	fmt.Println()
	count := strings.Index(str, " ")

	fmt.Println(str[:count])
	fmt.Println(strings.TrimSpace(str[count : len(str)-1]))
}

var lock = &sync.Mutex{} // 创建互锁
type single struct {
	Name string
	P    sync.RWMutex
} // 创建结构体

var singleInstance *single // 创建指针

func getInstance() *single {
	if singleInstance == nil { //!!!注意这里check nil了两次
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			fmt.Println("创建单例")
			singleInstance = &single{}
		} else {
			fmt.Println("单例对象已创建")
		}
	} else {
		fmt.Println("单例对象已创建")
	}
	return singleInstance
}

func TestTask_GetList(t1 *testing.T) {

}
