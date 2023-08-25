package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"strings"
	"sync"

	"proxy_forwarder/gost/core/logger"
	xlogger "proxy_forwarder/gost/x/logger"
	"proxy_forwarder/meta"

	"github.com/judwhite/go-svc"
)

var (
	cfgFile      string
	outputFormat string
	services     stringList
	nodes        stringList
	debug        bool
	apiAddr      string
	metricsAddr  string
)

func init() {
	args := strings.Join(os.Args[1:], "  ")

	if strings.Contains(args, " -- ") {
		var (
			wg  sync.WaitGroup
			ret int
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for wid, wargs := range strings.Split(" "+args+" ", " -- ") {
			wg.Add(1)
			go func(wid int, wargs string) {
				defer wg.Done()
				defer cancel()
				worker(wid, strings.Split(wargs, "  "), &ctx, &ret)
			}(wid, strings.TrimSpace(wargs))
		}

		wg.Wait()

		os.Exit(ret)
	}
}

func worker(id int, args []string, ctx *context.Context, ret *int) {
	cmd := exec.CommandContext(*ctx, os.Args[0], args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("_GOST_ID=%d", id))

	cmd.Run()
	if cmd.ProcessState.Exited() {
		*ret = cmd.ProcessState.ExitCode()
	}
}

func init() {
	var printVersion bool
	var listenPort string

	// flag.Var(&services, "L", "Listen addresses")
	flag.StringVar(&listenPort, "P", "4128", "Listen port")
	flag.Var(&nodes, "F", "Proxy server to forward the traffic to")
	flag.StringVar(&metricsAddr, "metrics", "", "metrics service address")
	flag.BoolVar(&printVersion, "V", false, "print version")
	flag.BoolVar(&debug, "D", false, "debug mode")
	flag.Parse()

	services = []string{
		fmt.Sprintf("redirect://127.0.0.1:%s", listenPort),
		fmt.Sprintf("redirect://[::1]:%s", listenPort),
		fmt.Sprintf("redu://127.0.0.1:%s", listenPort),
		fmt.Sprintf("redu://[::1]:%s", listenPort),
	}

	if printVersion {
		fmt.Printf("Proxy-Forwarder Version: %s | Gost Version: %s", meta.VERSION_FWD, meta.VERSION_GOST)
		os.Exit(0)
	}

	logger.SetDefault(xlogger.NewLogger())
}

func main() {
	p := &program{}
	if err := svc.Run(p); err != nil {
		log.Fatal(err)
	}
}
