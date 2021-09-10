package common

import (
	"errors"
	"reflect"
	"strconv"
)

func Marshal(info map[string]interface{}, r []string) (string, error) {

	if len(info) != len(r) {
		return "", errors.New("inconsistent array length")
	}
	json := "{"
	for i := 0; i < len(r); i++ {
		k := r[i]
		mydata := reflect.ValueOf(info).MapIndex(reflect.ValueOf(k))
		switch reflect.ValueOf(mydata.Interface()).Kind() {
		case reflect.String:
			json += "\"" + k + "\":\"" + reflect.ValueOf(mydata.Interface()).String() + "\","
		case reflect.Float64:
			v := reflect.ValueOf(mydata.Interface())
			strings := strconv.FormatInt(int64(v.Float()), 10)
			json += "\"" + k + "\":" + strings + ","
		default:
			json += "\"" + k + "\":" + "\"\"" + ","
		}

	}
	return json[:len(json)-1] + "}", nil

}
