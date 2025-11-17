package resp

import (
	"github.com/dromara/carbon/v2"
)

type Time struct {
	CreatedAt carbon.DateTime `json:"createdAt"`
	UpdatedAt carbon.DateTime `json:"updatedAt"`
}

type Base struct {
	Id uint `json:"id"`
	Time
}

type Resp struct {
	Code      int         `json:"code"`
	Data      interface{} `json:"data"`
	Msg       string      `json:"msg"`
	RequestId string      `json:"requestId"`
}

type Page struct {
	PageNum  uint  `json:"pageNum"`
	PageSize uint  `json:"pageSize"`
	Total    int64 `json:"total"`
}

type PageData struct {
	Page
	List interface{} `json:"list"`
}
