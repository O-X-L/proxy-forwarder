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
	var forwardProxy string
	var noLogTime bool
	var listenUDP bool
	listenerParams := "?sniffing=true"

    fmt.Println("PROJECT: github.com/O-X-L/proxy-forwarder (MIT, © OXL IT Service 2025)")
    fmt.Printf("FORKED FROM: github.com/go-gost/gost (MIT, © ginuerzh 2016)\n\n")

	flag.StringVar(&listenPort, "P", "", "Listen port")
	flag.StringVar(&forwardProxy, "F", "", "Proxy server to forward the traffic to")
	flag.BoolVar(&tproxyMode, "T", false, "Run in TProxy mode")
	flag.StringVar(&tproxyMark, "M", "", "Mark to set for TPRoxy traffic")
	flag.BoolVar(&printVersion, "V", false, "Show version")
	flag.BoolVar(&meta.DEBUG, "D", false, "Enable debug mode")
	flag.BoolVar(&listenUDP, "U", false, "Enable UDP listeners")
	flag.StringVar(&metricsAddr, "metrics", "", "Set a metrics service address (prometheus)")
	flag.BoolVar(&noLogTime, "no-log-time", false, "Do not add timestamp to logs")
	flag.Parse()

	if printVersion {
		fmt.Printf("Proxy-Forwarder Version: %s\nForked from Gost-Version: %s\n\n", meta.VERSION_FWD, meta.VERSION_GOST)
		os.Exit(0)
	}

	if listenPort == "" || forwardProxy == "" {
		fmt.Printf("Proxy-Forwarder %s\n\n", meta.VERSION_FWD)
		fmt.Println("USAGE:")
		fmt.Println("  -P 'Listen port' (required)")
		fmt.Println("  -F 'Proxy server to forward the traffic to' (required, Example: 'http://192.168.0.1:3128')")
		fmt.Println("  -T 'Run in TProxy mode' (default: false)")
		fmt.Println("  -M 'Mark to set for TProxy traffic' (default: None)")
		fmt.Println("  -V 'Show version'")
		fmt.Println("  -D 'Enable debug mode'")
		fmt.Println("  -metrics 'Set a metrics service address (prometheus)' (Example: '127.0.0.1:9000', Docs: 'https://gost.run/en/tutorials/metrics/')")
		fmt.Println("  -no-log-time 'Do not add timestamp to logs'")
		fmt.Printf("\n\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(forwardProxy, "http://") && !strings.HasPrefix(forwardProxy, "https://") {
		fmt.Println("The forward-proxy must include its protocol! (http/https)")
		os.Exit(1)
	}

	meta.LOG_TIME = !noLogTime

	nodes = []string{forwardProxy}

	if tproxyMode {
		listenerParams += "&tproxy=true"
		if tproxyMark != "" {
			listenerParams += fmt.Sprintf("&so_mark=%s", tproxyMark)
		}
	}

    if listenUDP {
        services = []string{
            fmt.Sprintf("redirect://127.0.0.1:%s%s", listenPort, listenerParams),
            fmt.Sprintf("redirect://[::1]:%s%s", listenPort, listenerParams),
            fmt.Sprintf("redu://127.0.0.1:%s%s", listenPort, listenerParams),
            fmt.Sprintf("redu://[::1]:%s%s", listenPort, listenerParams),
        }

    } else {
        services = []string{
            fmt.Sprintf("redirect://127.0.0.1:%s%s", listenPort, listenerParams),
            fmt.Sprintf("redirect://[::1]:%s%s", listenPort, listenerParams),
        }
    }
}

func main() {
	p := &program{}
	if err := svc.Run(p); err != nil {
		log.Fatal(err)
	}
}
