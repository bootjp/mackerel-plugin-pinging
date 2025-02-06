package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/jessevdk/go-flags"
	ping "github.com/prometheus-community/pro-bing"
)

// version by Makefile
var version string

type cmdOpts struct {
	Host       string `long:"host" description:"Hostname to ping" required:"true"`
	Timeout    int    `long:"timeout" default:"1000" description:"timeout millisec per ping"`
	Interval   int    `long:"interval" default:"10" description:"sleep millisec after every ping"`
	Count      int    `long:"count" default:"10" description:"Count Sending ping"`
	Size       int    `long:"size" default:"56" description:"Payload size"`
	Privileged bool   `long:"privileged" default:"false" description:"Use privileged ICMP raw socket"`
	KeyPrefix  string `long:"key-prefix" description:"Metric key prefix" required:"true"`
	Version    bool   `short:"v" long:"version" description:"Show version"`
}

func round(f float64) int64 {
	return int64(math.Round(f)) - 1
}

func resolveHost(Host string) (*ping.Pinger, error) {
	pinger := ping.New(Host)
	err := pinger.Resolve()
	return pinger, err
}

func rttMilliSec(rtt time.Duration) float64 {
	return float64(rtt.Nanoseconds()) / 1000 / 1000
}

func getStats(opts cmdOpts) error {
	pinger, err := resolveHost(opts.Host)
	if err != nil {
		errorNow := uint64(time.Now().Unix())
		fmt.Printf("pinging.%s_rtt_count.success\t%f\t%d\n", opts.KeyPrefix, 0.0, errorNow)
		fmt.Printf("pinging.%s_rtt_count.error\t%f\t%d\n", opts.KeyPrefix, float64(opts.Count), errorNow)
		return err
	}
	defer pinger.Stop()

	stats := &ping.Statistics{}

	// preflight
	pinger.Timeout = time.Millisecond * time.Duration(opts.Timeout)
	pinger.Interval = time.Millisecond * time.Duration(opts.Interval)
	pinger.Count = opts.Count
	pinger.Size = opts.Size

	// privileged is use raw socket icmp ping.
	// privileged == false is use udp ping.
	pinger.SetPrivileged(opts.Privileged)
	pinger.OnFinish = func(s *ping.Statistics) {
		stats = s
	}

	err = pinger.Run()
	if err != nil {
		log.Printf("error in preflight: %v", err)
		// ignore error
	}

	now := time.Now().Unix()

	// heuristics error count
	errorCnt := opts.Count - len(stats.TTLs)
	successCnt := len(stats.TTLs)

	fmt.Printf("pinging.%s_rtt_count.success\t%d\t%d\n", opts.KeyPrefix, successCnt, now)
	fmt.Printf("pinging.%s_rtt_count.error\t%d\t%d\n", opts.KeyPrefix, errorCnt, now)
	if successCnt > 0 {
		fmt.Printf("pinging.%s_rtt_ms.max\t%f\t%d\n", opts.KeyPrefix, rttMilliSec(stats.MaxRtt), now)
		fmt.Printf("pinging.%s_rtt_ms.min\t%f\t%d\n", opts.KeyPrefix, rttMilliSec(stats.MinRtt), now)
		fmt.Printf("pinging.%s_rtt_ms.average\t%f\t%d\n", opts.KeyPrefix, rttMilliSec(stats.AvgRtt), now)
		fmt.Printf("pinging.%s_rtt_ms.90_percentile\t%f\t%d\n", opts.KeyPrefix, rttMilliSec(stats.Rtts[round(float64(successCnt)*0.90)]), now)
	}

	return nil
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opts := cmdOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	_, err := psr.Parse()

	if opts.Version {
		fmt.Printf(`%s %s
Compiler: %s %s
`,
			os.Args[0],
			version,
			runtime.Compiler,
			runtime.Version())
		return 0
	}
	if err != nil {
		return 1
	}
	err = getStats(opts)
	if err != nil {
		log.Printf("%v", err)
		return 1
	}
	return 0
}
