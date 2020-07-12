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
	ConsumerKey    string `yaml:"ConsumerKey"`
	ConsumerSecret string `yaml:"ConsumerSecret"`
	AccessToken    string `yaml:"AccessToken"`
	AccessSecret   string `yaml:"AccessSecret"`
	ReadTweet      int64  `yaml:"ReadTweet"`
}

func (config TwitterConfig) connect() {
	consumerKey := config.ConsumerKey
	consumerSecret := config.ConsumerSecret
	accessToken := config.AccessToken
	accessSecret := config.AccessSecret

	if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessSecret == "" {
		log.Fatal().Msgf("Twitter: Consumer key/secret and Access token/secret required")
	}

	authConfig := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := authConfig.Client(oauth1.NoContext, token)

	// Twitter Client
	client := twitter.NewClient(httpClient)
	if config.ReadTweet != 0 {
		tweet, _, _ := client.Statuses.Show(config.ReadTweet, nil)
		ParseTweet(tweet)
		return
	}

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = ParseTweet

	log.Info().Msg("Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Follow:        []string{"842095599100997636"},
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal().Msgf("Twitter", err.Error())
	}

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msgf("Received signal %d", <-ch)

	log.Info().Msg("Stopping Stream...")
	stream.Stop()
}
