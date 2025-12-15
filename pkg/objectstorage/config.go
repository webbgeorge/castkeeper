package objectstorage

import (
	"context"
	"errors"
	"path"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

func ConfigureObjectStorage(ctx context.Context, cfg config.Config) (ObjectStorage, error) {
	httpClient := framework.NewHTTPClient(time.Minute * 15)

	switch cfg.ObjectStorage.Driver {
	case config.ObjectStorageDriverLocal:
		return &LocalObjectStorage{
			HTTPClient: httpClient,
			Root: config.MustOpenLocalFSRoot(
				path.Join(cfg.DataPath, "objects"),
			),
		}, nil

	case config.ObjectStorageDriverS3:
		// uses aws environment variables to configure the SDK
		awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
		s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.UsePathStyle = cfg.ObjectStorage.S3ForcePathStyle
		})

		return &S3ObjectStorage{
			HTTPClient: httpClient,
			S3Client:   s3Client,
			BucketName: cfg.ObjectStorage.S3Bucket,
			Prefix:     cfg.ObjectStorage.S3Prefix,
		}, nil

	default:
		return nil, errors.New("unknown objectstorage driver")
	}
}
