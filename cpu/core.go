package cpu

import (
	"time"
	proc "github.com/c9s/goprocinfo/linux"
)

type CPUCoreRecord struct {
	Time		time.Time	`json:"-"`
	Id			int			`json:"id"`
	User		uint64		`json:"user"`
	Nice		uint64		`json:"nice"`
	System		uint64		`json:"system"`
	Idle		uint64		`json:"idle"`
	IOWait		uint64		`json:"IOWait"`
	IRQ			uint64		`json:"IRQ"`
	SoftIRQ		uint64		`json:"softIRQ"`
	Steal		uint64		`json:"steal"`
	Guest		uint64		`json:"guest"`
	GuestNice	uint64		`json:"guest_nice"`
	TotalUsed	uint64		`json:"total_used"`
}

type CPUCoreMonitor struct {
	data chan CPUProgramRecord
	stop chan int
	duration time.Duration
}

type CPUCoreRecordResult struct {
	Time		string		`json:"time"`
	Records		[]CPUCoreRecord	`json:"records"`
}

type CoreResultCallback func (CPUCoreRecordResult)

func NewCoreRecorder(duration time.Duration) (m *CPUCoreMonitor) {
	m = new(CPUCoreMonitor)
	m.duration = duration
	return
}

func (m *CPUCoreMonitor) calacCoreCPU(record1 , record2 []CPUCoreRecord) CPUCoreRecordResult {
	var result CPUCoreRecordResult
	result.Records = make([]CPUCoreRecord, len(record1))
	result.Time = record2[0].Time.Format("2006-01-02 15:04:05")
	for i := 0; i < len(record1); i++ {
		off := uint64(record2[i].Time.Sub(record1[i].Time).Seconds())/uint64(m.duration)
		if off == 0 {
			off = 1
		}
		result.Records[i].Id		= record1[i].Id
		result.Records[i].User		= (record2[i].User - record1[i].User)/uint64(off)
		result.Records[i].Nice		= (record2[i].Nice - record1[i].Nice)/uint64(off)
		result.Records[i].System	= (record2[i].System - record1[i].System)/uint64(off)
		result.Records[i].Idle		= (record2[i].Idle - record1[i].Idle)/uint64(off)
		result.Records[i].IOWait	= (record2[i].IOWait - record1[i].IOWait)/uint64(off)
		result.Records[i].IRQ		= (record2[i].IRQ - record1[i].IRQ)/uint64(off)
		result.Records[i].SoftIRQ	= (record2[i].SoftIRQ - record1[i].SoftIRQ)/uint64(off)
		result.Records[i].Steal		= (record2[i].Steal - record1[i].Steal)/uint64(off)
		result.Records[i].Guest		= (record2[i].Guest - record1[i].Guest)/uint64(off)
		result.Records[i].GuestNice	= (record2[i].GuestNice - record1[i].GuestNice)/uint64(off)
		result.Records[i].TotalUsed	= (record2[i].TotalUsed - record1[i].TotalUsed)/uint64(off)
	}
	return result
}


func (m *CPUCoreMonitor) recordCPU() ([]CPUCoreRecord, error){
	statFile := "/proc/stat"
	stat, err := proc.ReadStat(statFile)
	if (err != nil) {
		return nil, err
	}
	records := make([]CPUCoreRecord, len(stat.CPUStats))
	for i, s := range(stat.CPUStats) {
		records[i] = CPUCoreRecord{
			Time: time.Now(),
			Id: i,
			User: s.User,
			Nice: s.Nice,
			System:s.System,
			Idle: s.Idle,
			IOWait:s.IOWait,
			IRQ: s.IRQ,
			SoftIRQ: s.SoftIRQ,
			Steal: s.Steal,
			Guest: s.Guest,
			GuestNice: s.GuestNice,
		}
		records[i].TotalUsed = s.User + s.Nice + s.System + s.IOWait + s.IRQ + s.SoftIRQ + s.Steal + s.Guest + s.GuestNice
	}
	return records, nil
}

func(m *CPUCoreMonitor) Monitor(cb CoreResultCallback) (error) {
	r1, err := m.recordCPU()
	if err != nil {
		return err
	}
	for {
		time.Sleep(m.duration)
		r2, err := m.recordCPU()
		if err != nil {
			return err
		}
		cb(m.calacCoreCPU(r1, r2))
		r1 = r2
	}
}