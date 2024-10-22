package queues

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
)

func ScanFileThroughQuickScope(task Task) (bool, error) {
	fp := extras.SANDBOX_FILE_PATHS + fmt.Sprintf("%d", task.Id)
	args := []string{"python3", "/home/prateek/Qu1cksc0pe/qu1cksc0pe.py", "--file"}

	fod := model.FileOnDemand{}
	queryString := fmt.Sprintf("SELECT * FROM %s WHERE id = %d", FileOnDemandTable, task.Id)

	fileOnDemand := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fod,
	}
	dao.GormOperations(&fileOnDemand, config.Db, dao.EXEC)

	ext := filepath.Ext(fod.FileName)
	if ext == "" {
		bytes, _ := os.ReadFile(fp)
		kind, _ := filetype.Match(bytes)
		ext = kind.Extension
	}

	if ext == ".txt" {
		util.EmbedTextInHTML(fp)
		ext = ".html"
	}

	args = append(args, fp)

	switch ext {
	case ".doc", ".docm", ".docx", "xls", ".xlsm", ".xlsx", ".pdf", ".one", ".htm", ".html", ".rtf":
		args = append(args, "--docs")
	case ".zip", ".rar", ".ace":
		args = append(args, "--archive")
	case ".elf":
		args = append(args, "--sigcheck")
	default:
		return false, fmt.Errorf("file type not supported: %s", ext)
	}

	// slog.Printf("Executing command: %s %v", args[0], args[1:])

	cmd := exec.Command(args[0], args[1:]...)
	var stdout, stderr bytes.Buffer
	stdin, err := cmd.StdinPipe()
	if err != nil {
		// slog.Printf("ERROR CREATING STDIN PIPE: %v", err)
		return false, err
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		// slog.Printf("ERROR STARTING COMMAND: %v", err)
		return false, err
	}

	// Provide input if needed
	go func() {
		defer stdin.Close()
		inputs := []string{"Y\n"} // Predefine inputs for different prompts
		for _, input := range inputs {
			fmt.Fprint(stdin, input)
		}
	}()

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		// slog.Printf("ERROR WHILE EXECUTING COMMAND: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		return false, err
	}

	trimmedString := strings.TrimSpace(stdout.String())

	// Print the byte values of the string to check for hidden characters
	found := false
	i := 0
	for i+4 < len(trimmedString) {
		if trimmedString[i] == 'I' && trimmedString[i+1] == '0' && trimmedString[i+2] == '\x04' && trimmedString[i+3] == 'C' {
			found = true
			break
		}
		i++
	}

	var malwareFound bool = false
	if found || strings.Contains(trimmedString, "I0C") || strings.Contains(strings.ToLower(trimmedString), "macro found") || strings.Contains(strings.ToLower(trimmedString), "rule name") {
		malwareFound = true
	}

	// slog.Println(stdout.String())
	return malwareFound, nil
}
