package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func Marshal(info map[string]interface{}, sortKey []string, isIndent bool) (string, error) {

	if len(info) != len(sortKey) {
		return "", errors.New("inconsistent array length")
	}

	var jsonBuf strings.Builder
	jsonBuf.WriteString("{")
	for i := 0; i < len(sortKey); i++ {
		k := sortKey[i]
		jsonBuf.WriteString("\"" + k + "\":")
		var content string
		mydata := reflect.ValueOf(info).MapIndex(reflect.ValueOf(k))
		switch reflect.ValueOf(mydata.Interface()).Kind() {
		case reflect.String:

			content = "\"" + reflect.ValueOf(mydata.Interface()).String() + "\""
		case reflect.Float64:
			v := reflect.ValueOf(mydata.Interface())
			strings := strconv.FormatInt(int64(v.Float()), 10)
			content = strings
		default:
			content = "null"
		}
		if i < len(sortKey)-1 {
			jsonBuf.WriteString(content + ",")
		} else {
			jsonBuf.WriteString(content)
		}
	}
	jsonBuf.WriteString("}")
	if !isIndent {
		return jsonBuf.String(), nil
	}
	var retBuf bytes.Buffer
	err := json.Indent(&retBuf, []byte(jsonBuf.String()), "", "\t")

	if err != nil {
		return "", err
	}
	retBuf.WriteTo(os.Stdout)
	return retBuf.String(), nil

}
