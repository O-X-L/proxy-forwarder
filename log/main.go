package log

import (
	"fmt"
	"proxy_forwarder/meta"
	"time"
)

func ErrorS(pkg string, msg string) {
	fmt.Printf(
		"%s | ERROR | %s | %s\n",
		time.Now().Format(meta.LOG_TIME_FORMAT),
		pkg, msg,
	)
}

func Error(pkg string, msg error) {
	ErrorS(pkg, fmt.Sprintf("%s", msg))
}

func ConnErrorS(pkg string, src string, dst string, msg string) {
	fmt.Printf(
		"%s | ERROR | %s | %s <=> %s | %s\n",
		time.Now().Format(meta.LOG_TIME_FORMAT),
		src, dst, pkg, msg,
	)
}

func ConnError(pkg string, src string, dst string, msg error) {
	ConnErrorS(pkg, src, dst, fmt.Sprintf("%s", msg))
}

func conn(lvl string, pkg string, src string, dst string, msg string) {
	if lvl == "DEBUG" {
		if meta.DEBUG {
			fmt.Printf(
				"%s | DEBUG | %s | %s <=> %s | %s\n",
				time.Now().Format(meta.LOG_TIME_FORMAT),
				pkg, src, dst, msg,
			)
		}
	} else {
		fmt.Printf(
			"%s | %s | %s | %s <=> %s | %s\n",
			time.Now().Format(meta.LOG_TIME_FORMAT),
			lvl, pkg, src, dst, msg,
		)
	}
}

func l(lvl string, pkg string, msg string) {
	if lvl == "DEBUG" {
		if meta.DEBUG {
			fmt.Printf(
				"%s | DEBUG | %s | %s\n",
				time.Now().Format(meta.LOG_TIME_FORMAT),
				pkg, msg,
			)
		}
	} else {
		fmt.Printf(
			"%s | %s | %s | %s\n",
			time.Now().Format(meta.LOG_TIME_FORMAT),
			lvl, pkg, msg,
		)
	}
}

func Debug(pkg string, msg string) {
	l("DEBUG", pkg, msg)
}

func ConnDebug(pkg string, src string, dst string, msg string) {
	conn("DEBUG", pkg, src, dst, msg)
}

func Info(pkg string, msg string) {
	l("INFO", pkg, msg)
}

func ConnInfo(pkg string, src string, dst string, msg string) {
	conn("INFO", pkg, src, dst, msg)
}
