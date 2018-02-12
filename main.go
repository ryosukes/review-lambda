package main

import (
	"fmt"
	"math/rand"
	"time"

	slack "review-ojisan/slack"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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
	prURL := fmt.Sprintf("%s", request.Body)
	message := generateMessage(reviewer, prURL)

	sl := slack.NewSlack(config.Slack.URL, message, config.Slack.UserName, ":eyes:", "", config.Slack.Channel)

	sl.Send()

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func loadConfig() {
	var file = "./config/config.toml"
	_, err := toml.DecodeFile(file, &config)

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
	return config.Slack.Group + " " + reviewer.SlackAccount + " " + reviewer.Name + "さん、コードレビューをお願いします！ " + prURL
}
