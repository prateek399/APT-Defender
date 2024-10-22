package service

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	lock "github.com/subchen/go-trylock"
)

var mu = lock.New()

func UploadBuild(inputf multipart.File, filename string) model.APIResponse {
	if ok := mu.TryLock(extras.LOCK_TIME_OUT); !ok {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, fmt.Errorf("timeout while acquiring lock"))
	}

	defer func() {
		mu.Unlock()
		// fmt.Printf("lock released")
	}()

	defer inputf.Close()

	filePath := ""
	curDir := extras.CONFIG_BASE_PATH
	// fmt.Println(curDir)
	if filename == "main" {
		filePath = filepath.Join(curDir, "main")
	} else if filename == "initSystem" {
		filePath = filepath.Join(curDir, "InitSystem", "initSystem")
	} else {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, fmt.Errorf("invalid filename"))
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(inputf)
	fmt.Println("buf --  ", len(buf.Bytes()))
	if err := execSha256ChecksumValidation(inputf); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	existingf, err := os.Open(filePath)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	// storing backup build if new build crashes
	restoref, err := os.OpenFile(filepath.Join(extras.TEMP_BUILD_PATH, "temp", filename), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		existingf.Close()
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}
	defer restoref.Close()

	existingBuf := new(bytes.Buffer)
	existingBuf.ReadFrom(existingf)

	if _, err = io.Copy(restoref, bytes.NewReader(existingBuf.Bytes())); err != nil {
		existingf.Close()
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	ok := false
	defer func(ok *bool) {
		if !*ok {
			// fmt.Println(len("In order to restore"))
			RestoreOldBuild(filename)
		}
	}(&ok)

	fmt.Println("before copying :", len(existingBuf.Bytes()))
	// if _, err := io.Copy(existingf, inputf); err != nil {
	// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	// }
	//Copy main to temp
	//rm main
	//main store

	if err := CopyFileContent(buf.Bytes(), filename); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}
	if filePath == "InitSystem/initSystem" {
		fmt.Println("copying")
		execCopyCommand()
	}
	ok = true

	fmt.Println("service restarting....")
	execServiceRestart(filename)
	fmt.Println("service restarted....")
	return model.NewSuccessResponse(extras.ERR_SUCCESS, "Build Uploaded Successfully")
}

func CopyFileContent(inputf []byte, filename string) error {
	filePath := extras.CONFIG_BASE_PATH
	fmt.Println("in==>>>. ", len(inputf))

	if filename == "initSystem" {
		filePath = filePath + "/InitSystem/initSystem"
	} else {
		filePath = filePath + "/" + filename
	}

	if err := os.Remove(filePath); err != nil {
		log.Println("error while removing", err)
	}

	// outf, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0777)
	// if err != nil {
	// 	log.Println("error while opening: ", err)
	// }
	// defer outf.Close()
	err := os.WriteFile(filePath, inputf, 0777)
	log.Println(err)
	// _, err = outf.Write(inputf)
	newBuf, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("error while reading: ", err)
	}

	// buf := new(bytes.Buffer)
	// buf.ReadFrom(outf)

	log.Println("New Buffer size: -----> ", len(newBuf))

	return err
}

func RestoreOldBuild(filename string) {
	filePath := filepath.Join(extras.TEMP_BUILD_PATH, "temp", filename)
	buildf, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Println("error while creating: ", err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(buildf)

	if err := os.Remove(filepath.Join(extras.CONFIG_BASE_PATH, filename)); err != nil {
		log.Println(err)
	}

	outf, err := os.OpenFile(filepath.Join(extras.CONFIG_BASE_PATH, filename), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Println("error while open: ", err)
	}

	if _, err := io.Copy(outf, bytes.NewReader(buf.Bytes())); err != nil {
		log.Println("error while copy: ", err)
	}
}

func execSha256ChecksumValidation(inputf multipart.File) error {
	filePath, err := os.Getwd()
	if err != nil {
		return err
	}

	f, err := os.Create("tempBuild")
	if err != nil {
		return err
	}
	defer f.Close()

	inputBuf := new(bytes.Buffer)
	inputBuf.ReadFrom(inputf)
	if _, err := io.Copy(f, bytes.NewReader(inputBuf.Bytes())); err != nil {
		return err
	}

	cmd := exec.Command("sha256sum", f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	shaFile, err := os.Create(filePath + ".sha256")
	if err != nil {
		fmt.Println("Error creating SHA256 file:", err)
		return err
	}
	defer shaFile.Close()

	defer func() {
		if err := os.Remove(filePath + ".sha256"); err != nil {
			log.Println("error while removing: ", err)
		}
		if err := os.Remove("tempBuild"); err != nil {
			log.Println("error while removing: ", err)
		}
	}()

	// Set the output of the command to write to the SHA256 file
	cmd.Stdout = shaFile

	err = cmd.Run()
	if err != nil {
		log.Println("Command:", cmd)
		log.Println("Error executing command:", err)
		log.Println("Stdout:", cmd.Stdout)
		log.Println("Stderr:", cmd.Stderr)
		return err
	}

	if err := executeCommandForValidatingBinary(filePath); err != nil {
		return err
	}
	return nil
}

func executeCommandForValidatingBinary(filePath string) error {
	cmd := exec.Command("sha256sum", "-c", filePath+".sha256")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Println("Command:", cmd)
		log.Println("Error executing command:", err)
		log.Println("Stdout:", cmd.Stdout)
		log.Println("Stderr:", cmd.Stderr)
		return err
	}

	return nil
}

func execCopyCommand() {
	filePath := extras.CONFIG_BASE_PATH

	cmd := exec.Command("cp", filePath+"/InitSystem/initSystem", "/var/www/html/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Println("Command:", cmd)
		log.Println("Error executing command:", err)
		log.Println("Stdout:", cmd.Stdout)
		log.Println("Stderr:", cmd.Stderr)
	}

	perm := exec.Command("chmod", "777", "/var/www/html/initSystem")
	perm.Stdout = os.Stdout
	perm.Stderr = os.Stderr

	err = perm.Run()
	if err != nil {
		log.Println("Command:", perm)
		log.Println("Error executing command:", err)
		log.Println("Stdout:", perm.Stdout)
		log.Println("Stderr:", perm.Stderr)
	}
}

func execServiceRestart(filename string) error {
	ok := false
	defer func(ok *bool) {
		if !*ok {
			// fmt.Println(len("In order to restore"))
			RestoreOldBuild(filename)
			go func() {
				time.Sleep(3 * time.Second)
				log.Println("Restarting old build...")
				execOldServiceRestart(filename)
				log.Println("Restarted old build...")
			}()
		}
	}(&ok)

	filePath := "main"
	if filename == "initSystem" {
		filePath = "InitSystem/initSystem"
	}

	gocmd := exec.Command(filepath.Join(extras.CONFIG_BASE_PATH, filePath))

	goOutput, err := gocmd.CombinedOutput()
	if err != nil {
		log.Println("Command:", gocmd)
		log.Println("Error executing command:", err)
		log.Println("Stdout:", gocmd.Stdout)
		log.Println("Stderr:", gocmd.Stderr)
		return err
	}

	log.Println("Checking GLIBC error")
	if strings.Contains(string(goOutput), "`GLIBC_2.34' not found") || strings.Contains(string(goOutput), "`GLIBC_2.32' not found") {
		log.Println("incompatible GLIBC_2.34 and GLIBC_2.32 not found")
		return fmt.Errorf("incompatible GLIBC version")
	}
	log.Println("Checked GLIBC error")

	cmd := exec.Command("service", "backend", "restart")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Println("Command:", cmd)
		log.Println("Error executing command:", err)
		log.Println("Stdout:", cmd.Stdout)
		log.Println("Stderr:", cmd.Stderr)
		return err
	}
	log.Println("Error in output string: ", err)

	ok = true
	return nil
}

func execOldServiceRestart(filename string) {
	filePath := extras.TEMP_BUILD_PATH + "/temp"

	buildPath := extras.CONFIG_BASE_PATH + "/"
	if filename == "initSystem" {
		buildPath = "/var/www/html/"
	}

	cpCmd := exec.Command("cp", filepath.Join(filePath, filename), buildPath)
	cpCmd.Stdout = os.Stdout
	cpCmd.Stderr = os.Stderr

	err := cpCmd.Run()
	if err != nil {
		log.Println("Command:", cpCmd)
		log.Println("Error executing command:", err)
		log.Println("Stdout:", cpCmd.Stdout)
		log.Println("Stderr:", cpCmd.Stderr)
	}

	cmd := exec.Command("service", "backend", "restart")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Starting old build...")
	err = cmd.Run()
	if err != nil {
		log.Println("Command:", cmd)
		log.Println("Error executing command:", err)
		log.Println("Stdout:", cmd.Stdout)
		log.Println("Stderr:", cmd.Stderr)
	} else {
		log.Println("Startied old build...")
	}
}
