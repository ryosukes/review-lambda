package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	slack "review-ojisan/slack"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Config from config directory
type Config struct {
	Reviewers []ReviewerConfig `toml:"Reviewer"`
	Slack     SlackConfig      `toml:"Slack"`
}

// ReviewerConfig from config.toml
type ReviewerConfig struct {
	Name         string `toml:"name"`
	SlackAccount string `toml:"slack_account"`
}

// SlackConfig from config.toml
type SlackConfig struct {
	URL      string
	UserName string
	Channel  string
	Group    string
}

var config Config

// HandleRequest is Lambda Handler
func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	loadConfig()

	reviewer := selectReviewer()

	u, _ := url.Parse("https://dummy.com/?" + request.Body)
	query := u.Query()
	prURL := fmt.Sprintf("%s", query["text"])
	message := generateMessage(reviewer, prURL)

	sl := slack.NewSlack(config.Slack.URL, message, config.Slack.UserName, "", "http://3.bp.blogspot.com/-0SY0brETIYs/VaMNiZlDbUI/AAAAAAAAvZQ/hrfERj3OB4A/s800/man_49.png", config.Slack.Channel)

	sl.Send()

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func loadConfig() {
	var BUCKET = os.Getenv("BUCKET")
	var KEY = os.Getenv("KEY")

	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(endpoints.ApNortheast1RegionID),
	})

	file, _ := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(BUCKET),
		Key:    aws.String(KEY),
	})

	defer file.Body.Close()

	brb := new(bytes.Buffer)
	brb.ReadFrom(file.Body)

	_, err := toml.DecodeReader(brb, &config)

	if err != nil {
		panic(err)
	}
}

func selectReviewer() ReviewerConfig {
	reviewerCount := len(config.Reviewers)
	rand.Seed(time.Now().UnixNano())
	reviewerNum := rand.Intn(reviewerCount)

	return config.Reviewers[reviewerNum]
}

func generateMessage(reviewer ReviewerConfig, prURL string) string {
	r := strings.NewReplacer("[", "", "]", "")
	return config.Slack.Group + " " + reviewer.SlackAccount + " " + reviewer.Name + "さん、コードレビューをお願いします！ " + r.Replace(prURL)
}
