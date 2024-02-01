package cos

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/wire"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"github.com/zeromicro/go-zero/core/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type CosSDK struct {
	stsClient     *sts.Client
	cosClient     *cos.Client
	fileStsClient *sts.Client
	fileCosClient *cos.Client
	CDNConf       *config.CDNConfig
}

func NewCosSDK(config *config.Config) (*CosSDK, error) {
	bucketURL, err := url.Parse(config.CosConfig.CosHost())
	if err != nil {
		return nil, err
	}
	fileBucketURL, err := url.Parse(config.FileCosConfig.CosHost())
	if err != nil {
		return nil, err
	}
	return &CosSDK{
		stsClient: sts.NewClient(
			config.CosConfig.SecretId,
			config.CosConfig.SecretKey,
			nil),
		cosClient: cos.NewClient(&cos.BaseURL{
			BucketURL: bucketURL,
		}, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  config.CosConfig.SecretId,
				SecretKey: config.CosConfig.SecretKey,
			},
		}),
		fileStsClient: sts.NewClient(
			config.FileCosConfig.SecretId,
			config.FileCosConfig.SecretKey,
			nil),
		fileCosClient: cos.NewClient(&cos.BaseURL{
			BucketURL: fileBucketURL,
		}, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  config.FileCosConfig.SecretId,
				SecretKey: config.FileCosConfig.SecretKey,
			},
		}),
		CDNConf: config.CdnConfig,
	}, nil
}
func (s *CosSDK) GenerateURL(path string, ttl int) string {
	if ttl < s.CDNConf.MinTTL {
		ttl = s.CDNConf.MinTTL
	} else if ttl > s.CDNConf.MaxTTL {
		ttl = s.CDNConf.MaxTTL
	}
	url := s.CDNConf.Url
	key := s.CDNConf.Key
	now := time.Now().Add(-time.Duration(s.CDNConf.MaxTTL-ttl) * time.Second).Unix()
	signKey := "sign"
	timeKey := "t"
	ttlFormat := 10
	var requestURL string
	tsFormat := strconv.FormatInt(now, ttlFormat)
	sign := fmt.Sprintf("%s%s%s", key, path, tsFormat)
	signMD5 := md5.Sum([]byte(sign))
	signHex := hex.EncodeToString(signMD5[:])
	requestURL = fmt.Sprintf("%s%s?%s=%s&%s=%s", url, path, signKey, signHex, timeKey, tsFormat)
	return requestURL
}
func (s *CosSDK) GetCredential(ctx context.Context, opt *sts.CredentialOptions, isFile bool) (*sts.CredentialResult, error) {
	_, span := trace.TracerFromContext(ctx).Start(ctx, "sts/GetCredential", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	if isFile {
		return s.fileStsClient.GetCredential(opt)
	}
	return s.stsClient.GetCredential(opt)
}

func (s *CosSDK) GetPresignedURL(ctx context.Context, httpMethod, name, ak, sk string, expired time.Duration, opt interface{}, signHost ...bool) (*url.URL, error) {
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, "cos/Object/GetPresignedURL", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	return s.cosClient.Object.GetPresignedURL(ctx, httpMethod, name, ak, sk, expired, opt, signHost...)
}

func (s *CosSDK) Delete(ctx context.Context, name string, opt ...*cos.ObjectDeleteOptions) (*cos.Response, error) {
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, "cos/Object/Delete", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	return s.cosClient.Object.Delete(ctx, name, opt...)
}

func (s *CosSDK) BatchImageAuditing(ctx context.Context, opt *cos.BatchImageAuditingOptions) (*cos.BatchImageAuditingJobResult, *cos.Response, error) {
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, "cos/CI/BatchImageAuditing", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	return s.cosClient.CI.BatchImageAuditing(ctx, opt)
}

var CosSet = wire.NewSet(
	NewCosSDK,
)
