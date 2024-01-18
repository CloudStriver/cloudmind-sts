package provider

import (
	"github.com/CloudStriver/cloudmind-sts/biz/application/service"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/mapper/user"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/stores/redis"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/sdk/cos"
	"github.com/google/wire"
)

var AllProvider = wire.NewSet(
	ApplicationSet,
	InfrastructureSet,
)

var ApplicationSet = wire.NewSet(
	service.AuthSet,
	service.CosSet,
)

var InfrastructureSet = wire.NewSet(
	config.NewConfig,
	redis.NewRedis,
	cos.NewCosSDK,
	MapperSet,
)

var MapperSet = wire.NewSet(
	user.NewMongoMapper,
)
