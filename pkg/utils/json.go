package utils

import (
	"encoding/json"

	"github.com/ppxb/oreo-admin-go/pkg/log"
)

func Struct2Json(obj interface{}) string {
	str, err := json.Marshal(obj)
	if err != nil {
		log.Error("[STRUCT2JSON] Can not convert: %v", err)
	}
	return string(str)
}

func Json2Struct(str string, obj interface{}) {
	err := json.Unmarshal([]byte(str), obj)
	if err != nil {
		log.Error("[JSON2Struct] Can not convert: %v", err)
	}
}

func Struct2StructByJson(source interface{}, target interface{}) {
	jsonStr := Struct2Json(source)
	Json2Struct(jsonStr, target)
}

func JsonWithSort(str string) string {
	m := make(map[string]interface{})
	Json2Struct(str, &m)
	return Struct2Json(m)
}
