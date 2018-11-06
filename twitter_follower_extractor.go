package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	envAccessToken    = "TWITTER_ACCESS_TOKEN"
	envAccessSecret   = "TWITTER_ACCESS_SECRET"
	envConsumerKey    = "TWITTER_CONSUMER_KEY"
	envConsumerSecret = "TWITTER_CONSUMER_SECRET"
)

type twitterConfig struct {
	accessToken    string
	tokenSecret    string
	consumerKey    string
	consumerSecret string
}

func getTwitterClient(c twitterConfig) *twitter.Client {
	log.Printf("Getting new Twitter client\n")
	oc := oauth1.NewConfig(c.consumerKey, c.consumerSecret)
	ot := oauth1.NewToken(c.accessToken, c.tokenSecret)
	hc := oc.Client(oauth1.NoContext, ot)
	return twitter.NewClient(hc)
}

func getTwitterUser(client *twitter.Client, screenName *string) *twitter.User {
	log.Printf("Getting info for Twitter user %s \n", *screenName)
	user, resp, err := client.Users.Show(&twitter.UserShowParams{
		ScreenName: *screenName,
	})
	if err != nil {
		log.Println("Failed getting twitter user")
		log.Fatalf("%#v\n%#v", resp, err)
	}
	return user
}

var followers = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "twitter_follower_count",
		Help: "Twitter follower count",
	})

func init() {
	log.Printf("Registering Prometheus collectors\n")
	prometheus.MustRegister(followers)
}

func main() {
	var (
		screenName  = flag.String("twitter.screenname", "intigriti", "ScreenName used to track followers")
		addr        = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
		metricsPath = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	c := twitterConfig{
		accessToken:    os.Getenv(envAccessToken),
		tokenSecret:    os.Getenv(envAccessSecret),
		consumerKey:    os.Getenv(envConsumerKey),
		consumerSecret: os.Getenv(envConsumerSecret),
	}

	client := getTwitterClient(c)

	go func() {
		for {
			user := getTwitterUser(client, screenName)
			followers.Set(float64(user.FollowersCount))
			time.Sleep(2 * time.Second)
		}
	}()

	log.Printf("starting prometheus http\nListening on %s%s", *addr, *metricsPath)
	http.Handle(*metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
