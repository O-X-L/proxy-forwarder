package log

import (
	"fmt"
	"proxy_forwarder/meta"
	"time"
)

func log(lvl string, pkg string, msg string) {
	if meta.LOG_TIME {
		fmt.Printf(
			"%s | %s | %s | %s\n",
			time.Now().Format(meta.LOG_TIME_FORMAT),
			lvl, pkg, msg,
		)
	} else {
		fmt.Printf("%s | %s | %s\n", lvl, pkg, msg)
	}
}

func logConn(lvl string, pkg string, src string, dst string, msg string) {
	if meta.LOG_TIME {
		fmt.Printf(
			"%s | %s | %s | %s <=> %s | %s\n",
			time.Now().Format(meta.LOG_TIME_FORMAT),
			lvl, pkg, src, dst, msg,
		)
	} else {
		fmt.Printf("%s | %s | %s <=> %s | %s\n", lvl, pkg, src, dst, msg)
	}
}

func ErrorS(pkg string, msg string) {
	log("ERROR", pkg, msg)
}

func Error(pkg string, err error) {
	log("ERROR", pkg, fmt.Sprintf("%s", err))
}

func ConnErrorS(pkg string, src string, dst string, msg string) {
	logConn("ERROR", pkg, src, dst, msg)
}

func ConnError(pkg string, src string, dst string, err error) {
	logConn("ERROR", pkg, src, dst, fmt.Sprintf("%s", err))
}

func Debug(pkg string, msg string) {
	if meta.DEBUG {
		log("DEBUG", pkg, msg)
	}
}

func ConnDebug(pkg string, src string, dst string, msg string) {
	if meta.DEBUG {
		logConn("DEBUG", pkg, src, dst, msg)
	}
}

func Info(pkg string, msg string) {
	log("INFO", pkg, msg)
}

func ConnInfo(pkg string, src string, dst string, msg string) {
	logConn("INFO", pkg, src, dst, msg)
}

func Warn(pkg string, msg string) {
	log("WARN", pkg, msg)
}
