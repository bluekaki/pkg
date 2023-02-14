package pod

import (
	"os"
	"strings"
)

var (
	groupName string
	podName   string
)

func init() {
	// project-58d56db7b7-pzx7m
	hostname := strings.TrimSpace(os.Getenv("HOSTNAME"))
	if hostname == "" {
		panic("no hostname found")
	}

	index := strings.LastIndex(hostname, "-")
	if index <= 0 {
		panic("hostname illegal")
	}

	groupName = hostname[:index]
	podName = hostname[index+1:]
}

func GroupName() string {
	return groupName
}

func PodName() string {
	return podName
}
