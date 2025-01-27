package objectstorage

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/webbgeorge/castkeeper/pkg/util"
)

type S3ObjectStorage struct {
	HTTPClient *http.Client
	S3Client   *s3.Client
	BucketName string
	Prefix     string
}

func (s *S3ObjectStorage) SaveRemoteFile(ctx context.Context, remoteLocation, podcastGUID, fileName string) error {
	err := util.ValidateExtURL(remoteLocation)
	if err != nil {
		return fmt.Errorf("invalid remoteLocation '%s': %w", remoteLocation, err)
	}

	s3Key := fmt.Sprintf("%s/%s", podcastGUID, fileName)

	req, err := http.NewRequest(http.MethodGet, remoteLocation, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to download file with status '%d'", resp.StatusCode)
	}

	uploader := manager.NewUploader(s.S3Client)
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(s.Prefix + s3Key),
		Body:   resp.Body,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3ObjectStorage) ServeFile(ctx context.Context, r *http.Request, w http.ResponseWriter, podcastGUID, fileName string) error {
	s3Key := fmt.Sprintf("%s/%s", podcastGUID, fileName)

	res, err := s.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(s.Prefix + s3Key),
	})
	if err != nil {
		return err
	}

	// TODO support range requests/chunked downloads
	_, err = io.Copy(w, res.Body)
	return err
}
