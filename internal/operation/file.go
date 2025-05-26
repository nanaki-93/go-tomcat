package operation

import (
	"fmt"
	"os"
	"strings"
)

func ReplaceInFile(inFilePath, outFilePath string, mapToReplace map[string]string) error {
	inputFile, err := os.ReadFile(inFilePath)
	if err != nil {
		return fmt.Errorf("ReplaceInFile : %w", err)
	}
	out := string(inputFile)
	for oldStr, newStr := range mapToReplace {
		out = strings.ReplaceAll(out, oldStr, newStr)
	}

	return os.WriteFile(outFilePath, []byte(out), 0777)

}

func CheckInFile(fileToCheckPath string, sliceToCheck []string) error {
	fileToCheck, err := os.ReadFile(fileToCheckPath)
	if err != nil {
		return fmt.Errorf("ReplaceInFile : %w", err)
	}
	data := string(fileToCheck)

	for _, toCheck := range sliceToCheck {
		if strings.Contains(data, toCheck) {
			return fmt.Errorf("CheckInFile : %w", fmt.Errorf("file %s still contains %s", fileToCheckPath, toCheck))
		}
	}

	return nil
}
