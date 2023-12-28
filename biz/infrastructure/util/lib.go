package util

import (
	"fmt"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/log"
	"github.com/bytedance/sonic"
	"math/rand"
	"time"
)

func JSONF(v any) string {
	data, err := sonic.Marshal(v)
	if err != nil {
		log.Error("JSONF fail, v=%v, err=%v", v, err)
	}
	return string(data)
}

//func ParsePagination(opts *basic.PaginationOptions) (p *pagination.PaginationOptions) {
//	if opts == nil {
//		p = &pagination.PaginationOptions{}
//	} else {
//		p = &pagination.PaginationOptions{
//			Limit:     opts.Limit,
//			Offset:    opts.Offset,
//			Backward:  opts.Backward,
//			LastToken: opts.LastToken,
//		}
//	}
//	return
//}

func GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	return code
}
