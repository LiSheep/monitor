package cpu

import (
	"time"
)

type CPUMonitor struct {
	Duration	time.Duration
}



type CPUProgramRecordResult struct {
	Time		time.Time			`json:"time"`
	Records		[]CPUProgramRecord	`json:"records"`
}



