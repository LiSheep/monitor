package log

import (
	"log"
	"os"
	"time"
	"sync"
	"strconv"
	"fmt"
)

type Mlog struct {
	path string
	prefix string
	flag int
	log *log.Logger
	writer *os.File
	mu sync.RWMutex
	count int
	begTime time.Time
}


func NewMlog(path string, prefix string, flag int) (l *Mlog, err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	l = new(Mlog)
	l.writer = f
	l.path = path
	l.flag = flag
	l.prefix = prefix
	l.log = log.New(f, prefix, flag)
	l.count = 3
	l.begTime = time.Now()
	return
}

func (m *Mlog) rotateFile() {
	m.writer.Close()
	f, err := os.OpenFile(m.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic("error open"+err.Error())
		return
	}
	m.writer = f
	m.log = log.New(m.writer, m.path, m.flag)
}

func (m *Mlog) rotating() {
	fmt.Println("rotating")
	for i := m.count; i >= 0; i-- {
		filePath := m.path + "."+strconv.Itoa(i)
		_, err := os.Stat(filePath)
		if err == nil || os.IsExist(err) {
			if i == m.count {
				os.Remove(filePath)
			} else {
				os.Rename(filePath, m.path+"."+strconv.Itoa(i+1))
			}
		}
	}
	os.Rename(m.path, m.path+".0")
	m.rotateFile()
}

func (m *Mlog) judgeRotate() {
	now := time.Now()
	m.mu.Lock()
	if m.begTime.Day() < now.Day() {
		m.rotating()
		m.begTime = now
		fmt.Println(m.path, "rotate finish")
	}
	m.mu.Unlock()
}

func (m *Mlog) Println(v ...interface{}) {
	m.judgeRotate()
	m.mu.RLock()
	m.log.Println(v...)
	m.mu.RUnlock()
}

func (m *Mlog) Printf(format string, v ...interface{}) {
	m.judgeRotate()
	m.mu.RLock()
	m.log.Printf(format, v...)
	m.mu.RUnlock()
}

var logSys *Mlog
var logCpuCore *Mlog
var logCpuProgram *Mlog

const sysLogPath = "/usr/local/frigate/var/log/sys.log"
const cpuCorePath = "/usr/local/frigate/var/log/core.cpu.log"
const cpuProgramPath = "/usr/local/frigate/var/log/program.cpu.log"


var oldTime time.Time
func MLogInit() error {
	//err := NewMlog(sysLogPath, "", log.LstdFlags)
	//if err != nil {
	//	return err
	//}
	var err error
	oldTime = time.Now()
	logCpuCore, err = NewMlog(cpuCorePath, "", 0)
	if err != nil {
		return err
	}
	logCpuProgram, err =NewMlog(cpuProgramPath, "", 0)
	if err != nil {
		return err
	}
	return nil
}

func MLog(format string, v ...interface{}) {
	//logSys.Printf(format+"\n", v...)
}

func MCPUCore(str string) {
	logCpuCore.Println(str)
}

func MCPUProgram(str string) {
	logCpuProgram.Println(str)
}

func Test(){
	logCpuCore.judgeRotate()
}
