package cpu

import (
	proc "github.com/c9s/goprocinfo/linux"
	"time"
	"strconv"
	"sync"
	"io/ioutil"
)

type CPUProgramRecord struct {
	TimeStr		string		`json:"time"`
	Time		time.Time	`json:"-"`
	Pid			uint64		`json:"pid"`
	Command		string		`json:"command"`
	User		uint64		`json:"user"`
	CUser		int64		`json:"cuser"`
	Sys			uint64		`json:"sys"`
	CSys		int64		`json:"csys"`
	CoreId		int64		`json:"core_id"`
	Stat		string		`json:"stat"`
	Rss			int64		`json:"rss"`
	TotalUsed	uint64		`json:"total_used"`
}

type CPUProgramMonitor struct {
	duration time.Duration
	cache programCache
}

const programName = "monitor"

func NewProgramMonitor(duration time.Duration) (m *CPUProgramMonitor) {
	m = new(CPUProgramMonitor)
	m.duration = duration
	return
}

func (m *CPUProgramMonitor) recordCPU(pid uint64) (record CPUProgramRecord, err error) {
	statFile := "/proc/"+strconv.FormatUint(pid, 10)+"/stat"
	stat, err := proc.ReadProcessStat(statFile)
	if err != nil {
		return
	}
	record = CPUProgramRecord {
		Time:	time.Now(),
		Pid:	stat.Pid,
		User:	stat.Utime,
		CUser:	stat.Cutime,
		CSys:	stat.Cstime,
		Sys:	stat.Stime,
		CoreId:	stat.Processor,
		Stat:	stat.State,
		Rss:	stat.Rss,
	}
	record.TotalUsed = record.User + uint64(record.CUser) + uint64(record.CSys) + record.Sys
	return record, nil
}

func (m *CPUProgramMonitor) calacProgramCPU(record1, record2 CPUProgramRecord) (r CPUProgramRecord) {
	rtime := record2.Time
	off := uint64(record2.Time.Sub(record1.Time).Seconds())/uint64(m.duration.Seconds())
	if off == 0 {
		off = 1
	}
	off = off * uint64(m.duration.Seconds())
	r = CPUProgramRecord{
		Time:		rtime,
		Pid:		record1.Pid,
		User:		(record2.User - record1.User)/uint64(off),
		CUser:		(record2.CUser - record1.CUser)/int64(off),
		Sys:		(record2.Sys - record1.Sys)/uint64(off),
		CSys:		(record2.CSys - record1.CSys)/int64(off),
		CoreId:		(record2.CoreId),
		Stat:		(record2.Stat),
		Rss:		(record2.Rss),
		TotalUsed:	(record2.TotalUsed - record1.TotalUsed)/uint64(off),
	}
	r.TimeStr = rtime.Format("2006-01-02 15:04:05")
	return
}

type programInfo struct {
	pid uint64
	cmdline string
}

type programCache struct {
	rwMutex sync.RWMutex
	record map[uint64] *programInfo
}

type ProgramResultCallback func (CPUProgramRecord, error)

func (m *CPUProgramMonitor) monitorRouting(cb ProgramResultCallback, pid uint64) {
	cmd, err := proc.ReadProcessCmdline("/proc/"+strconv.FormatUint(pid, 10)+"/cmdline")
	if err != nil {
		// program down
		return
	}
	r1, err := m.recordCPU(pid)
	if err != nil {
		m.cache.rwMutex.Lock()
		delete(m.cache.record, pid)
		m.cache.rwMutex.Unlock()
		r1.Command = cmd
		cb(r1, err)
		return
	}
	var r CPUProgramRecord
	var r2 CPUProgramRecord
	for {
		time.Sleep(m.duration)
		r2, err = m.recordCPU(pid)
		if err != nil {
			m.cache.rwMutex.Lock()
			delete(m.cache.record, pid)
			m.cache.rwMutex.Unlock()
			r1.Command = cmd
			cb(r1, err)
			return
		}
		r = m.calacProgramCPU(r1, r2)
		r1 = r2
		r.Command = cmd
		cb(r, err)
	}
}

func ListProgram(path string) ([]uint64, error) {
	pids := make([]uint64, 0, 20)
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return  nil, err
	}
	for _, info := range(infos) {
		pid, err := strconv.ParseUint(info.Name(), 10, 64)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func (m *CPUProgramMonitor) Monitor(cb ProgramResultCallback, cmds []string) error {
	m.cache.record = make(map[uint64] *programInfo)
	for {
		pids, err := ListProgram("/proc")
		if err != nil {
			return err
		}
		for _, pid := range(pids) {
			m.cache.rwMutex.RLock()
			_, ok := m.cache.record[pid]
			m.cache.rwMutex.RUnlock()
			if ok { // program exist
				// do nothing
			} else { // program not exist
				ps := strconv.FormatUint(pid, 10)
				cmd, err := proc.ReadProcessCmdline("/proc/"+ps+"/cmdline")
				if err != nil {
					continue
				}
				pro := new(programInfo)
				pro.cmdline = cmd
				pro.pid = pid
				m.cache.rwMutex.Lock()
				m.cache.record[pid] = pro
				m.cache.rwMutex.Unlock()
				go m.monitorRouting(cb, pid)
			}
		}
		time.Sleep(m.duration)
	}
}
