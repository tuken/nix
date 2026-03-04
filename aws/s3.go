package aws

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client *s3.Client
	ctx    context.Context
}

// NewS3Client S3のクライアントを生成する 環境変数に「AWS_ACCESS_KEY_ID」「AWS_SECRET_ACCESS_KEY」「AWS_REGION": "ap-northeast-1"」をセットしておく必要がある
func NewS3Client() *S3Client {

	clt := S3Client{
		ctx: context.Background(),
	}

	cfg, err := config.LoadDefaultConfig(clt.ctx)
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	clt.client = s3.NewFromConfig(cfg)

	return &clt
}

func (c S3Client) GetSignedURL(bucket, key string) (string, error) {

	pre := s3.NewPresignClient(c.client)

	input := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	res, err := pre.PresignGetObject(c.ctx, &input, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", fmt.Errorf("error presign s3 object (bucket=%s, key=%s): %v", bucket, key, err)
	}

	return res.URL, nil
}

// Upload 指定されたバケットのキーオブジェクトを上書きアップロードする `mime`は`obj`のContentTypeを指定する
func (c S3Client) Upload(obj io.Reader, bucket, key, mime string) error {

	input := s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        obj,
		ContentType: aws.String(mime),
	}
	_, err := c.client.PutObject(c.ctx, &input)
	if err != nil {
		return fmt.Errorf("error upload s3 (bucket=%s, key=%s): %v", bucket, key, err)
	}

	return nil
}
