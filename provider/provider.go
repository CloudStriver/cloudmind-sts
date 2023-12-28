package provider

import (
	"github.com/CloudStriver/cloudmind-sts/biz/application/service"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/mapper/auth"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/stores/redis"
	"github.com/google/wire"
)

var AllProvider = wire.NewSet(
	ApplicationSet,
	InfrastructureSet,
)

var ApplicationSet = wire.NewSet(
	service.AuthSet,
)

var InfrastructureSet = wire.NewSet(
	config.NewConfig,
	redis.NewRedis,
	MapperSet,
)

var MapperSet = wire.NewSet(
	auth.NewMongoMapper,
)
