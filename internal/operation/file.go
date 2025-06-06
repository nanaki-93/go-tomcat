package operation

import (
	"fmt"
	"os"
	"strings"
)

func ReplaceInFile(fileSlice []string, mapToReplace map[string]string) error {
	for _, filePath := range fileSlice {
		inputFile, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("ReplaceInFile: ", filePath)
			return fmt.Errorf("ReplaceInFile : %w", err)
		}
		out := string(inputFile)
		for oldStr, newStr := range mapToReplace {
			out = strings.ReplaceAll(out, oldStr, newStr)
		}

		err = os.WriteFile(filePath, []byte(out), 0777)
		if err != nil {
			fmt.Printf("ReplaceInFile %s: %e", filePath, err)
			return fmt.Errorf("ReplaceInFile : %w", err)
		}
	}
	return nil

}

func CheckInFile(fileListToCheck []string, sliceToCheck []string) error {

	for _, fileToCheckPath := range fileListToCheck {
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
	}

	return nil
}
