package operation

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func UpdatePropsInFiles(fileSlice []string, mapToReplace map[string]string) error {
	for _, filePath := range fileSlice {
		inputFile, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("UpdatePropsInFiles: ", filePath)
			return fmt.Errorf("UpdatePropsInFiles : %w", err)
		}
		out := string(inputFile)
		for oldStr, newStr := range mapToReplace {
			out = strings.ReplaceAll(out, oldStr, newStr)
		}

		err = os.WriteFile(filePath, []byte(out), 0777)
		if err != nil {
			fmt.Printf("UpdatePropsInFiles %s: %e", filePath, err)
			return fmt.Errorf("UpdatePropsInFiles : %w", err)
		}
	}
	return nil

}

func CheckInFile(fileListToCheck []string, sliceToCheck []string) error {

	for _, fileToCheckPath := range fileListToCheck {
		fileToCheck, err := os.ReadFile(fileToCheckPath)
		if err != nil {
			return fmt.Errorf("UpdatePropsInFiles : %w", err)
		}
		data := string(fileToCheck)

		for _, toCheck := range sliceToCheck {
			if strings.Contains(data, toCheck) {
				return fmt.Errorf("CheckInFile : %w", fmt.Errorf("file %s still contains %s", fileToCheckPath, toCheck))
			}
		}
	}

	return nil
}

func CheckCopiedFiles(srcDir, targetDir string) error {
	allSourceFilesPath := make([]string, 0)
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		allSourceFilesPath = append(allSourceFilesPath, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("CheckCopiedFiles : %w", err)
	}
	for _, filePath := range allSourceFilesPath {
		filePath = strings.Replace(filePath, srcDir, targetDir, 1)
		_, err = os.Stat(filePath)
		if err != nil {
			slog.Error("CheckCopiedFiles : file " + filePath + " does not exist")
			return fmt.Errorf("CheckCopiedFiles : %w", err)
		}
	}

	return nil
}
