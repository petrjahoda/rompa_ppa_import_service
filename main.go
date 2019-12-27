package main

import (
	"github.com/jinzhu/gorm"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var (
	activeDevices  []Device
	runningDevices []Device
	workplaceSync  sync.Mutex
)

const version = "2019.4.2.21"
const deleteLogsAfter = 240 * time.Hour
const downloadInSeconds = 10

func main() {
	LogDirectoryFileCheck("MAIN")
	LogInfo("MAIN", "Program version "+version+" started")
	CreateConfigIfNotExists()
	LoadSettingsFromConfigFile()
	LogDebug("MAIN", "Using ["+DatabaseType+"] on "+DatabaseIpAddress+":"+DatabasePort+" with database "+DatabaseName)
	databaseAvailable := false
	for {
		start := time.Now()
		LogInfo("MAIN", "Program running")
		databaseAvailable = CheckDatabase()
		if databaseAvailable {
			UpdateActiveDevices("MAIN")
			DeleteOldLogFiles()
			LogInfo("MAIN", "Active devices: "+strconv.Itoa(len(activeDevices))+", running devices: "+strconv.Itoa(len(runningDevices)))
			for _, activeDevice := range activeDevices {
				activeWorkplaceIsRunning := CheckWorkplace(activeDevice)
				if !activeWorkplaceIsRunning {
					go RunDevice(activeDevice)
				}
			}

		}
		if time.Since(start) < (downloadInSeconds * time.Second) {
			sleepTime := downloadInSeconds*time.Second - time.Since(start)
			LogInfo("MAIN", "Sleeping for "+sleepTime.String())
			time.Sleep(sleepTime)
		}
	}
}

func CheckWorkplace(device Device) bool {
	for _, runningDevice := range runningDevices {
		if runningDevice.Name == device.Name {
			return true
		}
	}
	return false
}

func UpdateActiveDevices(reference string) {
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)

	if err != nil {
		LogError(reference, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	db.Where("DeviceType = ?", 100).Where("CustomerID = ?", 1000).Find(&activeDevices)
}

func RunDevice(device Device) {
	LogInfo(device.Name, "Device started running")
	networkFolderAvailable := false
	workplaceSync.Lock()
	runningDevices = append(runningDevices, device)
	workplaceSync.Unlock()
	deviceIsActive := true
	for deviceIsActive {
		if !networkFolderAvailable {
			networkFolderAvailable = MapNetworkFolder(device.OID, device.Setting, device.Name)
		}
		start := time.Now()
		data := device.DownloadDataFromFile()
		if len(data) > 0 {
			device.ProcessData(data)
			device.DeleteData()
		}
		LogInfo(device.Name, "Processing takes "+time.Since(start).String())
		device.Sleep(start)
		deviceIsActive = CheckActive(device)
	}
	RemoveWorkplaceFromRunningDevices(device)
	LogInfo(device.Name, "Device not active, stopped running")
}

func MapNetworkFolder(deviceId int, settings string, deviceName string) bool {
	LogInfo(deviceName, "Creating directory data")
	cmd := exec.Command("mkdir", "/home/"+strconv.Itoa(deviceId))
	err := cmd.Run()
	if err != nil {
		LogError(deviceName, "Problem creating directory, already exists: "+err.Error())
	}
	LogInfo(deviceName, "Directory created")
	LogInfo(deviceName, "Mapping network directory")
	cmd = exec.Command("mount", "-t", "cifs", "-v", "-o", "username=zapsi,password=Jahoda123,domain=ROMPACZ", settings, "/home/"+strconv.Itoa(deviceId))
	LogInfo(deviceName, cmd.String())
	err = cmd.Run()
	if err != nil {
		LogError(deviceName, "Problem mapping network directory: "+err.Error())
		return false
	}
	LogInfo(deviceName, "Network directory mapped successfully")
	return true
}

func CheckActive(device Device) bool {
	for _, activeDevice := range activeDevices {
		if activeDevice.Name == device.Name {
			LogInfo(device.Name, "Device still active")
			return true
		}
	}
	LogInfo(device.Name, "Device not active")
	return false
}

func RemoveWorkplaceFromRunningDevices(device Device) {
	for idx, runningDevice := range runningDevices {
		if device.Name == runningDevice.Name {
			workplaceSync.Lock()
			runningDevices = append(runningDevices[0:idx], runningDevices[idx+1:]...)
			workplaceSync.Unlock()
		}
	}
}
