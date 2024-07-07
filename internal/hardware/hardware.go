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

	output := fmt.Sprintf("Hostname: %s\nTotal Memory: %.2f MB\nUsed Memory: %.2f MB\nOS: %s", hostStat.Hostname, float32(vmStat.Total)/MB, float32(vmStat.Used)/MB, runTimeOS)

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
		output = fmt.Sprintf("CPU: %s\nCores: %d", cpuStat[0].ModelName, cpuStat[0].Cores)
	} else {
		output = fmt.Sprintf("CPU: %s\nCores: %d", cpuStat[0].ModelName, len(cpuStat))
	}

	return output, nil
}

func GetDiskSection() (string, error) {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return "", err
	}
	output := fmt.Sprintf("Total Disk Space: %d GB\nFree Disk Space: %d GB", diskStat.Total/GB, diskStat.Free/GB)

	return output, nil
}
