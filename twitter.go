package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type TwitterConfig struct {
	ReadTweet int64  `yaml:"ReadTweet"`
	UserId    string `yaml:"UserId"`
}

var twitterClient *twitter.Client

func (config TwitterConfig) connect() {
	consumerKey := os.Getenv("TwitterConsumerKey")
	consumerSecret := os.Getenv("TwitterConsumerSecret")
	accessToken := os.Getenv("TwitterAccessToken")
	accessSecret := os.Getenv("TwitterAccessSecret")

	if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessSecret == "" {
		log.Fatal().Msgf("Twitter: Consumer key/secret and Access token/secret required")
	}

	authConfig := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := authConfig.Client(oauth1.NoContext, token)

	// Twitter Client
	twitterClient = twitter.NewClient(httpClient)
	if config.ReadTweet != 0 {
		tweet, _, _ := twitterClient.Statuses.Show(config.ReadTweet, nil)
		ParseTweet(tweet)
		return
	}

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = ParseTweet

	log.Info().Msg("Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Follow:        []string{config.UserId},
		StallWarnings: twitter.Bool(true),
	}
	stream, err := twitterClient.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal().Msgf("Twitter error: %s", err.Error())
	}

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msgf("Received signal %d", <-ch)

	log.Info().Msg("Stopping Stream...")
	stream.Stop()
}

func GetTweetBefore(lastTweetId int64) *twitter.Tweet {
	pastTweet, _, _ := twitterClient.Statuses.Show(lastTweetId, nil)
	if pastTweet.QuotedStatusID != 0 {
		pastTweet, _, _ = twitterClient.Statuses.Show(pastTweet.QuotedStatusID, nil)
	}
	return pastTweet
}
