package main

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nu7hatch/gouuid"
)

const (
	bucketName = "pcc-final-project-media"
)

var (
	svc *s3.S3
)

func init() {
	sess := session.New(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewSharedCredentials("", "pcc"),
	})
	svc = s3.New(sess, nil)
}

func uploadMediaToS3(r io.ReadSeeker) (string, error) {
	key, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	fmt.Println("Uploading", key.String(), "to s3")

	params := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(fmt.Sprintf("/img/%s.png", key.String())),
		Body:        r,
		ACL:         aws.String("public-read"),
		ContentType: aws.String("image/png"),
	}

	resp, err := svc.PutObject(params)
	if err != nil {
		panic(err)
	}

	fmt.Println("Media uploaded")

	fmt.Println(resp)

	return key.String(), nil
}
