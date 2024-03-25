package service

import (
	"context"
	"github.com/CloudStriver/ToolGood"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/consts"
	gensts "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/sts"
	"github.com/google/wire"
	"github.com/samber/lo"
	"sync"
)

type IFilterService interface {
	ReplaceContent(ctx context.Context, req *gensts.ReplaceContentReq) (resp *gensts.ReplaceContentResp, err error)
	FindAllContent(ctx context.Context, req *gensts.FindAllContentReq) (resp *gensts.FindAllContentResp, err error)
}

type FilterService struct {
	Config *config.Config
	Filter *ToolGood.IllegalWordsSearch
}

func (s *FilterService) ReplaceContent(ctx context.Context, req *gensts.ReplaceContentReq) (resp *gensts.ReplaceContentResp, err error) {
	contents := make([]string, len(req.Contents))
	wg := sync.WaitGroup{}
	wg.Add(len(req.Contents))
	for i, content := range req.Contents {
		go func(i int, content string) {
			defer wg.Done()
			contents[i] = s.Filter.Replace(content, consts.ReplaceChar)
		}(i, content)
	}

	wg.Wait()

	return &gensts.ReplaceContentResp{
		Content: contents,
	}, nil
}

func (s *FilterService) FindAllContent(ctx context.Context, req *gensts.FindAllContentReq) (resp *gensts.FindAllContentResp, err error) {
	keywords := make([]*gensts.Keywords, len(req.Contents))
	wg := sync.WaitGroup{}
	wg.Add(len(req.Contents))
	for i, content := range req.Contents {
		go func(i int, content string) {
			defer wg.Done()
			keywords[i] = &gensts.Keywords{
				Keywords: lo.Map[*ToolGood.IllegalWordsSearchResult, string](s.Filter.FindAll(content), func(item *ToolGood.IllegalWordsSearchResult, index int) string {
					return item.Keyword
				}),
			}
		}(i, content)
	}
	wg.Wait()
	return &gensts.FindAllContentResp{
		Keywords: keywords,
	}, nil
}

var FilterSet = wire.NewSet(
	wire.Struct(new(FilterService), "*"),
	wire.Bind(new(IFilterService), new(*FilterService)),
)
