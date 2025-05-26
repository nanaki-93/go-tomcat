/*
Copyright © 2025 Marco Andreose <andreose.marco93@gmail.com>
*/
package cmd

import (
	"fmt"
	"github.com/nanaki-93/go-tomcat/internal/operation"
	"github.com/spf13/cobra"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var jspFolderSuffix = filepath.Join("src", "main", "webapp")

// startCmd represents the master command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update the app's jsp in the tomcat server",
	Long:  `update the app's jsp in the tomcat server. It will copy all the jsp files from the source folder to the destination folder in the tomcat server.`,
	Run:   execUpdateCmd,
	Args:  validateArgs(),
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func execUpdateCmd(cmd *cobra.Command, args []string) {
	const poolSize = 100
	pool := make(chan struct{}, poolSize)
	wg := new(sync.WaitGroup)

	appName := args[0]

	tm, err := createTomcatManager(CliBasePath, appName)
	operation.CheckErr(err)

	isRunning := false
	for _, tomcat := range tm.TomcatProps.RunningTomcats {
		if tomcat.AppTomcatName == appName {
			isRunning = true
			break
		}
	}

	if !isRunning {
		slog.Error("the tomcat server is not running", "appName", appName)
		slog.Error("you need to start the tomcat server before updating the jsp files")
		return
	}
	sourcePath := filepath.Join(tm.TomcatConfig.AppConfig.ProjectPath, jspFolderSuffix)
	destPath := filepath.Join(tm.TomcatPaths.HomeAppTomcat, "webapps", tm.TomcatConfig.AppConfig.ContextFileName)

	sourceFileList, err := os.ReadDir(sourcePath)
	startTime := time.Now()
	slog.Warn("Start at: time", "time", startTime.Format(time.RFC3339))
	counter := 0
	for _, file := range sourceFileList {
		if file.IsDir() || !strings.Contains(file.Name(), ".jsp") {
			continue
		}
		sourceFilePath := filepath.Join(sourcePath, file.Name())
		destFilePath := filepath.Join(destPath, file.Name())
		counter++
		pool <- struct{}{} // acquire slot
		wg.Add(1)
		go copyRoutine(wg, pool, sourceFilePath, destFilePath)
	}
	wg.Wait()
	endTime := time.Now()
	slog.Info("End at: time", "time", endTime.Format(time.RFC3339))
	slog.Info("Duration: time", "time", endTime.Sub(startTime).String())
	if counter == 0 {
		slog.Warn("No jsp files found to update in the source path", "sourcePath", sourcePath)
	} else {
		slog.Info("Total files copied: ", "files", counter)
		slog.Info("all the jsp files are updated in the tomcat server")
	}

	wg.Wait()
}
func copyRoutine(wg *sync.WaitGroup, pool chan struct{}, src, dst string) {
	defer wg.Done()
	defer func() { <-pool }() // release slot
	err := copyFileContents(src, dst)
	operation.CheckErr(err)
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copyFileContents : %w", err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("copyFileContents : %w", err)
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copyFileContents : %w", err)
	}
	return out.Sync()
}
