package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	sc "github.com/gr4c2-2000/go-binlog2api/internal/go-binlog2api"
)

func main() {
	go func() {
		sc.Logger.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	sc.LoggerInit()
	sc.Logger.Info("Logger Initialized")

	go binlogListener()

	runtime.Goexit()
	//defer scripts.LoggerClose()
	fmt.Println("Exit")

}
