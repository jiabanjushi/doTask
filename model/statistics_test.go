package model

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestCheckIsExistModelStatistics(t *testing.T) {

	st := Statistics{RegisterNum: 1}

	var i interface{} = st
	value := reflect.ValueOf(i)
	//注册人数数据更新
	RegisterNum := value.FieldByName("RegisterNum")
	fmt.Println(RegisterNum)

}

//包含上下限 [min, max]
func getRandomWithAll(min, max int) int64 {
	rand.Seed(time.Now().UnixNano())
	return int64(rand.Intn(max-min+1) + min)
}

// 不包含上限 [min, max)
func getRandomWithMin(min, max int) int64 {
	rand.Seed(time.Now().UnixNano())
	return int64(rand.Intn(max-min) + min)
}

// 不包含下限 (min, max]
func getRandomWithMax(min, max int) int64 {
	var res int64
	rand.Seed(time.Now().UnixNano())
Restart:
	res = int64(rand.Intn(max-min+1) + min)
	if res == int64(min) {
		goto Restart
	}
	return res
}

// 都不包含 (min, max)
func getRandomWithNo(min, max int) int64 {
	var res int64
	rand.Seed(time.Now().UnixNano())
Restart:
	res = int64(rand.Intn(max-min) + min)
	if res == int64(min) {
		goto Restart
	}
	return res

}

//var globalUserMap = sync.Map{}
//
//func GetUserLock(UserId int) sync.RWMutex {
//	lok, _ := globalUserMap.LoadOrStore(UserId, sync.RWMutex{})
//	return lok.(sync.RWMutex)
//}

//type UA struct {
//	sys  sync.RWMutex
//	Name string
//}
//
//func (U *UA) GET() {
//	U.sys = GetUserLock(14)
//	U.sys.Lock()
//	go func() {
//		time.Sleep(4 * time.Second)
//		fmt.Println(U.Name + "执行了")
//		U.sys.Unlock()
//	}()
//
//}
