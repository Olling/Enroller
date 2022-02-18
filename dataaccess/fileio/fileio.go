package fileio

import (
	"os"
	"fmt"
	"path"
	"sync"
	"time"
	"strings"
	"syscall"
	"strconv"
	"os/exec"
	"io/ioutil"
	"math/rand"
	"encoding/json"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enroller/dataaccess/config"
)

var (
	SyncOutputMutex		sync.Mutex
)

func WriteToFile(filepath string, content string, appendToFile bool, filemode os.FileMode) (err error) {
	SyncOutputMutex.Lock()
	defer SyncOutputMutex.Unlock()

	if appendToFile {
		file, fileerr := os.OpenFile(filepath, os.O_APPEND, filemode)
		defer file.Close()
		if fileerr != nil {
			return fileerr
		}

		_, writeerr := file.WriteString(content)
		return writeerr
	} else {
		err := ioutil.WriteFile(filepath, []byte(content), filemode)
		if err != nil {
			slog.PrintError("Error while writing file", err)
			return err
		}
		return nil
	}
}

func WriteStructToFile(s interface{}, filepath string, appendToFile bool) (err error) {
	json, err := utils.StructToJson(s)

	if err != nil {
		slog.PrintError("Could not convert struct to json", err)
		return err
	}

	return WriteToFile(filepath, json, appendToFile, 0664)

}

func CheckScriptPath(filepath string) error {
	if filepath == "" {
		slog.PrintError("ScriptPath is empty")
		return fmt.Errorf("ScriptPath is empty")
	}

	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		slog.PrintError("ScriptPath does not exist: '" + filepath + "'")
		return fmt.Errorf("ScriptPath does not exist")
	}

	return nil
}

func LoadFromFile(s interface{}, filepath string) error {
	file, err := os.Open(filepath)
	defer file.Close()

	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&s)
	file.Close()

	return nil
}


func GetFileList(directoryPath string) ([]os.FileInfo, error) {
	filelist, err := ioutil.ReadDir(directoryPath)
	return filelist, err
}

func LoadScripts(scripts map[string]objects.Script) error {
	scripts = make(map[string]objects.Script)

	filelist, err := GetFileList(config.Configuration.ScriptDirectory)
	if err != nil {
		slog.PrintDebug("Failed to load script list", err)
		return err
	}

	for _,directory := range filelist {
		if !directory.IsDir() {
			slog.PrintDebug("Ignoring the following file in the script path", directory.Name())
			continue
		}

		var script objects.Script
		scriptID := directory.Name()
		scriptPath := path.Join(config.Configuration.ScriptDirectory, scriptID)

		err = LoadFromFile(&script, path.Join(scriptPath, scriptID + ".json"))
		if err != nil{
			slog.PrintError("Failed to get script information from", path.Join(scriptPath, scriptID + ".json"))
			continue
		}

		//TODO there might be a reference problem here
		scripts[scriptID] = script
	}

	return nil
}


func FileExist(filepath string) bool {
	_, existsErr := os.Stat(filepath)

	if os.IsNotExist(existsErr) {
		return false
	} else {
		return true
	}
}


func RunScript(scriptPath string, server objects.Server, scriptID string, timeout int) error {
	err := CheckScriptPath(scriptPath)
	if err != nil {
		return err
	}

	tempDirectory := path.Join(config.Configuration.TempPath, server.ServerID, strconv.Itoa(rand.Intn(200)))
	inventoryPath := path.Join(tempDirectory, "single.inventory")

	logPath := path.Join(config.Configuration.LogPath, scriptID, server.ServerID + ".log")

	_, err = os.Stat(path.Join(config.Configuration.LogPath, scriptID))
	if os.IsNotExist(err) {
		err := os.MkdirAll(path.Join(config.Configuration.LogPath, scriptID), 0744)
		if err != nil {
			slog.PrintDebug("Failed to create log directory", err)
		}
	}

	outfile, err := os.Create(logPath)
	if err != nil {
		slog.PrintError("Error creating logfile", outfile.Name, err)
	}
	defer outfile.Close()

	_, existsErr := os.Stat(tempDirectory)
	if os.IsNotExist(existsErr) {
		createErr := os.MkdirAll(tempDirectory, 0755)
		if createErr != nil {
			slog.PrintError(createErr)
			return fmt.Errorf("Could not create temp directory: " + tempDirectory)
		}
	}

	json, _ := utils.GetInventoryInJSON([]objects.Server{server})
	json = strings.Replace(json, "\"", "\\\"", -1)
	inventory := "#!/bin/bash\necho \"" + json + "\""

	WriteToFile(inventoryPath, inventory, false, 0755)

	cmd := exec.Command("/bin/bash", scriptPath, inventoryPath, server.ServerID)
	cmd.Stdout = outfile
	cmd.Stderr = outfile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err = cmd.Start(); err != nil {
		slog.PrintError("Could not start the script", scriptID, err)
		return err
	}

	timer := time.AfterFunc(time.Duration(timeout) * time.Second, func() {
		slog.PrintError("The script ", scriptID + "(" + server.ServerID + ")", "has reached the timeout - Killing process", cmd.Process.Pid)
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, 15)
		}
	})

	execErr := cmd.Wait()
	timer.Stop()

	if execErr != nil {
		slog.PrintError("Error while excecuting script", scriptID, "Please see the log for more info:", logPath)
		return execErr
	}

	return nil
}


func EnrollServer(server objects.Server) error {
	err := RunScript(config.Configuration.EnrollmentScriptPath, server, "Enroll", config.Configuration.Timeout)
	if err != nil {
		slog.PrintError("Error running script against", server.ServerID, "(" + server.IP + "):", err)
		utils.Notification("Enrolld failure", "Failed to enroll the following new server: " + server.ServerID + "(" + server.IP + ")", server)

		return err
	}

	slog.PrintInfo("Enrolld script successful: " + server.ServerID)
	return err
}
