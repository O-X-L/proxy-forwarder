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
	var tproxyMode bool
	var tproxyMark string
	listenerParams := "?sniffing=true"

	flag.StringVar(&listenPort, "P", "", "Listen port")
	flag.Var(&nodes, "F", "Proxy server to forward the traffic to")
	flag.BoolVar(&tproxyMode, "T", false, "Run in TProxy mode")
	flag.StringVar(&tproxyMark, "M", "100", "Mark to set for TPRoxy traffic")
	flag.StringVar(&metricsAddr, "m", "", "Set a metrics service address (prometheus)")
	flag.BoolVar(&printVersion, "V", false, "Show version")
	flag.BoolVar(&debug, "D", false, "Enable debug mode")
	flag.Parse()

	if printVersion {
		fmt.Printf("\nProxy-Forwarder Version: %s\nGost Version: %s\n\n", meta.VERSION_FWD, meta.VERSION_GOST)
		os.Exit(0)
	}

	if listenPort == "" || len(nodes) == 0 {
		fmt.Printf("Proxy-Forwarder %s\n\n", meta.VERSION_FWD)
		fmt.Println("USAGE:")
		fmt.Println("  -P 'Listen port' (required)")
		fmt.Println("  -F 'Proxy server to forward the traffic to' (required, Example: 'http://192.168.0.1:3128')")
		fmt.Println("  -T 'Run in TProxy mode' (default: false)")
		fmt.Println("  -M 'Mark to set for TProxy traffic' (default: 100)")
		fmt.Println("  -m 'Set a metrics service address (prometheus)' (Example: '127.0.0.1:9000', Docs: 'https://gost.run/en/tutorials/metrics/')")
		fmt.Println("  -V 'Show version'")
		fmt.Printf("  -D 'Enable debug mode'\n\n")
		os.Exit(1)
	}

	if tproxyMode {
		listenerParams += fmt.Sprintf("tproxy=true&so_mark=%s", tproxyMark)
	}

	services = []string{
		fmt.Sprintf("redirect://127.0.0.1:%s%s", listenPort, listenerParams),
		fmt.Sprintf("redirect://[::1]:%s%s", listenPort, listenerParams),
		fmt.Sprintf("redu://127.0.0.1:%s%s", listenPort, listenerParams),
		fmt.Sprintf("redu://[::1]:%s%s", listenPort, listenerParams),
	}

	logger.SetDefault(xlogger.NewLogger())
}

func main() {
	p := &program{}
	if err := svc.Run(p); err != nil {
		log.Fatal(err)
	}
}
