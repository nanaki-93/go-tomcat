package operation

import (
	"log/slog"
	"os"
)

func CheckErr(err error, msg ...interface{}) {
	if err != nil {
		if len(msg) == 0 {
			slog.Error("Error:", err)
		} else {
			slog.Error("Error", err, ":", msg)
		}
		os.Exit(1)
	}
}
