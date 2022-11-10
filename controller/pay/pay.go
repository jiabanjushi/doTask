package pay

import (
	"sort"
)

func SortString(params map[string]string) string {
	var dataParams string
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	//拼接
	for _, k := range keys {
		dataParams = dataParams + k + "=" + params[k] + "&"
	}
	ff := dataParams[0 : len(dataParams)-1]
	return ff
}
