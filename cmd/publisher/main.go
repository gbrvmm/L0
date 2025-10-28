package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	stan "github.com/nats-io/stan.go"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	var file string
	flag.StringVar(&file, "file", "sample/model.json", "path to json file to publish")
	flag.Parse()

	clusterID := getenv("STAN_CLUSTER_ID", "test-cluster")
	clientID := getenv("STAN_CLIENT_ID", "publisher-1")
	url := getenv("STAN_URL", "nats://localhost:4222")
	channel := getenv("CHANNEL", "orders")

	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("read file: %v", err)
	}

	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(url))
	if err != nil {
		log.Fatalf("stan connect: %v", err)
	}
	defer sc.Close()

	if err := sc.Publish(channel, b); err != nil {
		log.Fatalf("publish: %v", err)
	}
	log.Printf("published %d bytes to channel %s", len(b), channel)
}
