package filter

import (
	"github.com/CloudStriver/ToolGood"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
)

func NewFilter(config *config.Config) *ToolGood.IllegalWordsSearch {
	filter := ToolGood.NewIllegalWordsSearch()
	filter.UseDuplicateWordFilter = config.FilterConfig.UseDuplicateWordFilter
	filter.UseIgnoreCase = config.FilterConfig.UseIgnoreCase
	filter.UseSimplifiedChineseConverter = config.FilterConfig.UseSimplifiedChineseConverter
	filter.UseDBCcaseConverter = config.FilterConfig.UseDBCcaseConverter
	filter.LoadFromDB(&ToolGood.MongoConfig{
		DataSource: config.Mongo.URL,
		DB:         config.Mongo.DB,
	})
	return filter
}
