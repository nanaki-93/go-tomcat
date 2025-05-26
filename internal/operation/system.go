package operation

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
)

func PrintCmd(cmd *exec.Cmd) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	slog.Info("Executing command",
		"command", cmd.String(),
		"working_dir", cmd.Dir)

}

func isFreePort(portToCheck string) bool {

	conn, _ := net.Listen("tcp", net.JoinHostPort(hostToCheck, portToCheck))
	if conn != nil {
		err := conn.Close()
		if err != nil {
			CheckErr(err, "Error closing connection of portToCheck:"+portToCheck)
		}
		return true
	}
	slog.Info("Port not available", portToCheck)
	return false
}

func sliceToSlice[T any](items []T, selector func(T) int) []int {
	used := make([]int, 0, len(items))
	for _, item := range items {
		used = append(used, selector(item))
	}
	return used
}
