package operation

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
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

// YesNoPrompt asks yes/no questions using the label.
func YesNoPrompt(label string, def bool) bool {
	choices := "Y/n"
	if !def {
		choices = "y/N"
	}

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		_, err := fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		if err != nil {
			return def
		}
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return def
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}

// StringPrompt asks for a string value using the label
func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		_, err := fmt.Fprint(os.Stderr, label+" ")
		if err != nil {
			slog.Error("Error prompting :", label, err)
		}
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func GetOrderedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
