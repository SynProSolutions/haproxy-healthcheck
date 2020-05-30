package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-systemd/activation"
	"github.com/shirou/gopsutil/cpu"
)

// NOTE: `apt-get install gcc libc6-dev`

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

// stateFile exists and is valid, then message is returned, to support
// https://cbonte.github.io/haproxy-dconv/2.0/configuration.html#5.2-agent-check
var stateFile = "/var/run/haproxy-healthcheck"

// CPUData is a global data structure written by an evaluation goroutine
// and read by all TCP connections
type CPUData struct {
	stats           [2]cpu.TimesStat
	isActive        bool
	currentInstance int
	rwMutex         sync.Mutex
}

var cpuData CPUData

func init() {
	cpuData.currentInstance = 1
	cpuData.isActive = false
	cpuData.stats[0] = currentCPUTimes()
}

// report userHZ, usually being 100
func getClockTicksPerSecond() (ticksPerSecond uint64) {
	var scClkTck C.long
	scClkTck = C.sysconf(C._SC_CLK_TCK)
	ticksPerSecond = uint64(scClkTck)
	return
}

var userHz = float64(getClockTicksPerSecond())

// report second entry of /proc/uptime, see proc(5)
func getIdleTime() (uptime float64) {
	contents, _ := ioutil.ReadFile("/proc/uptime")

	reader := bufio.NewReader(bytes.NewBuffer(contents))
	line, _, _ := reader.ReadLine()
	fields := strings.Fields(string(line))

	val, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return
	}
	return val
}

func initValue() int {
	var prev cpu.TimesStat
	// zero value of float64 == 0.0, so if we subtract this, it's the initial value we want
	return differenceValue(prev, currentCPUTimes())
}

func differenceValue(prev, cur cpu.TimesStat) int {

	cpuUse := [2]float64{prev.User, cur.User}
	cpuNice := [2]float64{prev.Nice, cur.Nice}
	cpuSystem := [2]float64{prev.System, cur.System}
	cpuIdle := [2]float64{prev.Idle, cur.Idle}
	cpuIowait := [2]float64{prev.Iowait, cur.Iowait}
	cpuIrQ := [2]float64{prev.Irq, cur.Irq}
	cpuSoftirq := [2]float64{prev.Softirq, cur.Softirq}
	cpuSteal := [2]float64{prev.Steal, cur.Steal}

	// implementation based on procps (proc/sysinfo.c + vmstat.c), see
	// https://sources.debian.org/src/procps/2:3.3.16-5/vmstat.c/?hl=361#L306
	duse := (cpuUse[1] - cpuUse[0] + cpuNice[1] - cpuNice[0]) * userHz
	dsys := (cpuSystem[1] - cpuSystem[0] + cpuIrQ[1] - cpuIrQ[0] + cpuSoftirq[1] - cpuSoftirq[0]) * userHz
	didl := (cpuIdle[1] - cpuIdle[0]) * userHz
	diow := (cpuIowait[1] - cpuIowait[0]) * userHz
	dstl := cpuSteal[1] - cpuSteal[0]

	div := duse + dsys + didl + diow + dstl

	if div == 0 {
		div = 1
		didl = 1
	}

	divo2 := div / 2
	result := ((100*didl + divo2) / div)

	return int(result)
}

func currentCPUTimes() cpu.TimesStat {
	cpuTimes, err := cpu.Times(false)
	if err != nil {
		panic(err)
	}
	return cpuTimes[0]
}

func handleRequest(conn net.Conn) {
	cpuData.rwMutex.Lock()

	if !cpuData.isActive {
		data, err := ioutil.ReadFile(stateFile)
		if err != nil {
			// we might end up here if file doesn't exist, e.g.
			// on first invocation within socket activation,
			// let's try to report CPU data then instead
			cpuData.isActive = true
		} else {
			var output string
			content := strings.TrimSpace(string(data))
			switch content {
			case "down", "drain", "failed", "maint", "stopped", "ready", "up":
				output = content + " \n"
			default:
				log.Printf("WARN: unsupported instructions in file detected: '%s'", content)
				cpuData.isActive = true
			}
			conn.Write([]byte(string(output)))
		}
	}

	if cpuData.isActive {
		var prev cpu.TimesStat
		var cur cpu.TimesStat

		if cpuData.currentInstance == 0 {
			prev = cpuData.stats[1]
			cur = cpuData.stats[0]
		} else {
			prev = cpuData.stats[0]
			cur = cpuData.stats[1]
		}

		diff := differenceValue(prev, cur)

		conn.Write([]byte(fmt.Sprintf("%d%% \n", diff)))
	}

	cpuData.rwMutex.Unlock()
	conn.Close()
}

func evaluateCPU() {
	for {
		cpuData.rwMutex.Lock()
		inst := cpuData.currentInstance
		cpuData.stats[inst] = currentCPUTimes()
		cpuData.currentInstance = (inst + 1) % len(cpuData.stats)
		cpuData.rwMutex.Unlock()

		time.Sleep(time.Second)
	}
}

func updateServerState() {
	for {
		cpuData.rwMutex.Lock()
		_, err := os.Stat(stateFile)
		cpuData.isActive = os.IsNotExist(err)
		cpuData.rwMutex.Unlock()

		time.Sleep(time.Second)
	}
}

func main() {
	flag.StringVar(&stateFile, "f", "/var/run/haproxy-healthcheck", "Specify healthcheck status file.")
	flag.Parse()

	fmt.Printf("INFO: Healtcheck status file set to %s\n", stateFile)

	// without socket activation:
	// listener, err := net.Listen("tcp", "localhost:3333")
	// with socket activation:
	listeners, err := activation.Listeners()
	if err != nil {
		log.Panicf("Cannot retrieve listeners: %s", err)
	}
	if len(listeners) != 1 {
		log.Panicln("This program is meant to be invoked via socket activation")
	}

	listener := listeners[0]
	defer listener.Close()

	go evaluateCPU()
	go updateServerState()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panicln(err)
		}

		go handleRequest(conn)
	}
}
