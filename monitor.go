package main

import (
	"github.com/tenchlee/monitor/cpu"
	"time"
	"fmt"
	"encoding/json"
	"github.com/tenchlee/monitor/log"
	"runtime"
	"os"
	"syscall"
	"os/signal"
)

func programResultCallback(r cpu.CPUProgramRecord, err error) {
	if err != nil {
		//log.MLog("pid :%d(%s) down", r.Pid, r.Command)
		return
	}
	if r.TotalUsed < 10 {
		return
	}
	x, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
	}
	log.MCPUProgram(string(x))
}

func coreResultCallback(r cpu.CPUCoreRecordResult) {
	x, err := json.Marshal(r)
	if err != nil {
		log.MLog("core err: %s", err.Error())
	}
	log.MCPUCore(string(x))
}

func usage() {
	fmt.Errorf("%s\n", `
		monitor run\r\n
	`)
}

func signalRegister() {
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGHUP)
		for {
			<- ch
			fmt.Println("recv hup")
			log.Test()
		}
	}()
}

func main() {
	runtime.GOMAXPROCS(1)
	log.MLogInit()
	signalRegister()
	args := os.Args[1:]
	go cpu.NewProgramMonitor(time.Second*3).Monitor(programResultCallback, args)
	cpu.NewCoreRecorder(time.Second).Monitor(coreResultCallback)
}