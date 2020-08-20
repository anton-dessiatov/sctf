package app

import (
	"fmt"

	"github.com/anton-dessiatov/sctf/direct/dal"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jinzhu/gorm"
)

type App struct {
	DB  *gorm.DB
	AWS *session.Session
}

func New() (*App, error) {
	config := readConfig()
	db := dal.MakeDB()

	creds := credentials.NewStaticCredentials(config.AWS.AccessKey, config.AWS.SecretKey, "")
	awsSession, err := session.NewSession(aws.NewConfig().WithCredentials(creds).
		WithRegion(config.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("session.NewSession: %w", err)
	}

	if config.AWS.AssumeRoleARN != "" {
		awsSession, err = session.NewSession(aws.NewConfig().WithCredentials(
			stscreds.NewCredentials(awsSession, config.AWS.AssumeRoleARN)).
			WithRegion(config.AWS.Region))
		if err != nil {
			return nil, fmt.Errorf("session.NewSession(role_arn): %w", err)
		}
	}

	return &App{
		DB:  db,
		AWS: awsSession,
	}, nil
}

var Instance *App
