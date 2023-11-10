package main

import (
	"flag"
	"time"
)

func main() {
	var (
		updateInterval    time.Duration
		individualScrapes string
	)
	flag.DurationVar(&updateInterval, "update", time.Hour*24, "Duration for computing global stats")
	flag.StringVar(&individualScrapes, "scrape", "response", "Whether to compute stats responsively (ie. only in response to requests from the api) or proactively (ie. run background individual stats computations, requiring more computation but potentially lower-latency for higher-traffic deployments)")

	flag.Parse()
}
