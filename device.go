package main

import (
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func (device Device) ProcessData(data string) {
	for _, line := range strings.Split(strings.TrimSuffix(data, "\n"), "\n") {
		parsedLine := strings.Split(line, "|")
		failName := parsedLine[0]
		productName := parsedLine[1]
		failId := CheckFailInDatabase(device, failName)
		productId := CheckProductInDatabase(device, productName)
		openTerminalInputOrderId, orderId := CheckOpenTerminalInputOrder(device)
		openOrderProductId := CheckProductForOpenOrder(orderId)
		failDateTime, err := time.Parse("02.01.2006 15:04:05", parsedLine[2])
		if err != nil {
			LogError(device.Name, "Problem parsing time: "+err.Error())
			continue
		}
		LogInfo(device.Name, "Open Order: "+strconv.Itoa(orderId)+" with terminal_input_order "+strconv.Itoa(openTerminalInputOrderId))
		LogInfo(device.Name, strconv.Itoa(productId)+"..."+strconv.Itoa(openOrderProductId))
		if openTerminalInputOrderId > 0 {
			if productId == openOrderProductId {
				LogInfo(device.Name, "Adding new fail to database: ["+strconv.Itoa(failId)+"] "+failName+"    ["+strconv.Itoa(productId)+"] "+productName+"    "+failDateTime.String()+"--"+parsedLine[2])
			} else {
				LogInfo(device.Name, "Closing old terminal_input_order")
				LogInfo(device.Name, "Creating new terminal_input_order")
				openTerminalInputOrderId, orderId = CheckOpenTerminalInputOrder(device)
				LogInfo(device.Name, "Adding new fail to database: ["+strconv.Itoa(failId)+"] "+failName+"    ["+strconv.Itoa(productId)+"] "+productName+"    "+failDateTime.String()+"--"+parsedLine[2])
			}
		} else {
			LogInfo(device.Name, "Creating new terminal_input_order")
			openTerminalInputOrderId, orderId = CheckOpenTerminalInputOrder(device)
			LogInfo(device.Name, "Adding new fail to database: ["+strconv.Itoa(failId)+"] "+failName+"    ["+strconv.Itoa(productId)+"] "+productName+"    "+failDateTime.String()+"--"+parsedLine[2])
		}
	}
}

func CheckProductForOpenOrder(orderId int) int {
	var order Order
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)

	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0
	}
	defer db.Close()
	db.Where("OID = ?", orderId).Find(&order)
	return order.ProductID
}

func CheckOpenTerminalInputOrder(device Device) (terminalInputOrderId int, orderId int) {
	var openOrder TerminalInputOrder
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)

	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0, 0
	}
	defer db.Close()
	db.Where("DeviceID = ?", device.OID).Where("DTE is null").Find(&openOrder)
	return openOrder.OID, openOrder.OrderID

}

func CheckProductInDatabase(device Device, productName string) int {
	var product Product
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)

	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0
	}
	defer db.Close()
	db.Where("Name = ?", productName).Find(&product)
	if product.OID > 0 {
		LogInfo(device.Name, "Product ["+productName+"] exists")
		return product.OID
	}
	var newProduct Product
	newProduct.Name = productName
	newProduct.Barcode = productName
	newProduct.ProductStatusID = 1
	db.Save(&newProduct)
	var anotherProduct Product
	db.Where("Name = ?", productName).Find(&anotherProduct)
	return anotherProduct.OID
}

func CheckFailInDatabase(device Device, failName string) int {
	var fail Fail
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)

	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0
	}
	defer db.Close()
	db.Where("Name = ?", failName).Find(&fail)
	if fail.OID > 0 {
		LogInfo(device.Name, "Fail ["+failName+"] exists")
		return fail.OID
	}
	var newFail Fail
	newFail.Name = failName
	newFail.Barcode = failName
	newFail.FailTypeID = 101
	db.Save(&newFail)
	var anotherFail Fail
	db.Where("Name = ?", failName).Find(&anotherFail)
	return anotherFail.OID
}

func (device Device) DeleteData() {
	var err = os.Remove("/home/" + strconv.Itoa(device.OID) + "/data.txt")
	if err != nil {
		LogError(device.Name, "Problem deleting file: "+err.Error())
		return
	}
	LogInfo(device.Name, "File deleted")
}

func (device Device) DownloadDataFromFile() string {
	data, err := ioutil.ReadFile("/home/" + strconv.Itoa(device.OID) + "/data.txt")
	if err != nil {
		LogError(device.Name, "Problem reading file: "+err.Error())
		return ""
	}
	return string(data)
}

func (device Device) Sleep(start time.Time) {
	if time.Since(start) < (downloadInSeconds * time.Second) {
		sleepTime := downloadInSeconds*time.Second - time.Since(start)
		LogInfo(device.Name, "Sleeping for "+sleepTime.String())
		time.Sleep(sleepTime)
	}
}
