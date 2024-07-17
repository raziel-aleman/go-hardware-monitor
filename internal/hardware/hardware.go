package hardware

import (
	"fmt"
	"regexp"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	_ = 1 << (10 * iota)
	KB
	MB
	GB
	TB
)

func GetSystemSection() (string, error) {
	runTimeOS := runtime.GOOS

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}

	hostStat, err := host.Info()
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("Operating System: %s\nPlatform: %s\nHostname: %s\nNumber of processes: %d\nTotal Memory: %.2f MB\nFree Memory: %.2f MB\nPercentage used memory: %.2f%%", runTimeOS, hostStat.Platform, hostStat.Hostname, hostStat.Procs, float64(vmStat.Total)/MB, float64(vmStat.Free)/MB, vmStat.UsedPercent)

	return output, nil
}

func GetCpuSection() (string, error) {
	cpuStat, err := cpu.Info()
	if err != nil {
		return "", err
	}

	var output string
	pattern := "Apple"
	matched, err := regexp.MatchString(pattern, cpuStat[0].ModelName)

	if err != nil {
		return "", err
	}

	if matched {
		output = fmt.Sprintf("Model Name: %s\nCores: %d", cpuStat[0].ModelName, cpuStat[0].Cores)
	} else {
		output = fmt.Sprintf("Model Name: %s\nFamily: %s\nSpeed: %.2f Mhz\nCores: %d", cpuStat[0].ModelName, cpuStat[0].Family, cpuStat[0].Mhz, len(cpuStat))
	}

	return output, nil
}

func GetDiskSection() (string, error) {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return "", err
	}
	output := fmt.Sprintf("Total Disk Space: %d GB\nUsed Disk Space: %d GB\nFree Disk Space: %d GB\nPercentage disk space usage: %.2f%%", diskStat.Total/GB, diskStat.Used/GB, diskStat.Free/GB, diskStat.UsedPercent)

	return output, nil
}
