package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H 結構體用于表示 HTTP 響應的格式
type H struct {
	Code  int         `json:"code"` // 這 `` 是我加的
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
	Rows  interface{} `json:"rows"`
	Total interface{} `json:"total"`
}

// Resp 函數用于發送 JSON 格式的 HTTP 響應
func Resp(w http.ResponseWriter, code int, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := H{
		Code: code,
		Data: data,
		Msg:  msg,
	}
	ret, err := json.Marshal(h)
	// 處理 JSON 編碼錯誤
	if err != nil {
		fmt.Println(err)
	}
	w.Write(ret)
}

// RespList 函數用于發送帶有列表數據和總數的 JSON 格式的 HTTP 響應
func RespList(w http.ResponseWriter, code int, data interface{}, total interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := H{
		Code:  code,
		Rows:  data,
		Total: total,
	}
	ret, err := json.Marshal(h)
	// 處理 JSON 編碼錯誤
	if err != nil {
		fmt.Println(err)
	}
	w.Write(ret)
}

func RespFail(w http.ResponseWriter, msg string) {
	Resp(w, -1, nil, msg)
}

func RespOK(w http.ResponseWriter, data interface{}, msg string) {
	Resp(w, 0, data, msg)
}

func RespOKList(w http.ResponseWriter, data interface{}, total interface{}) {
	RespList(w, 0, data, total)
}
