package main

import (
	captcha "github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/captcha/puzzle_captcha"
	"github.com/CloudStriver/cloudmind-sts/provider"
	"github.com/CloudStriver/go-pkg/utils/kitex/middleware"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	authservice "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/sts/stsservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	etcd "github.com/kitex-contrib/registry-etcd"
	"net"
)

func main() {
	initCaptcha()
	klog.SetLogger(log.NewKlogLogger())
	s, err := provider.NewStsServerImpl()
	if err != nil {
		panic(err)
	}
	addr, err := net.ResolveTCPAddr("tcp", s.ListenOn)
	if err != nil {
		panic(err)
	}

	r, err := etcd.NewEtcdRegistry(s.Config.EtcdConf.Hosts)
	if err != nil {
		panic(err)
	}
	svr := authservice.NewServer(
		s,
		server.WithServiceAddr(addr),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithRegistry(r),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: s.Name}),
		server.WithMiddleware(middleware.LogMiddleware(s.Name)),
	)

	err = svr.Run()
	if err != nil {
		log.Error(err.Error())
	}
}
func initCaptcha() {
	if err := captcha.LoadBackgroudImages("./biz/infrastructure/util/captcha/images/puzzle_captcha/backgroud"); err != nil {
		panic(err)
	}
	if err := captcha.LoadBlockImages("./biz/infrastructure/util/captcha/images/puzzle_captcha/block"); err != nil {
		panic(err)
	}

}
