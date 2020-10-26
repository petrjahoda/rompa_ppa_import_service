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
		LogInfo(device.Name, "Data product Id: "+strconv.Itoa(productId)+", open order product id: "+strconv.Itoa(openOrderProductId))
		LogInfo(device.Name, "FailId: "+strconv.Itoa(failId)+", fail datetime: "+failDateTime.String())
		if orderId == 100 {
			LogInfo(device.Name, "Internal order opened, updating")
			orderIdToInsert := GetOrderIdForProductId(device, productId)
			device.UpdateTerminalInputOrder(openTerminalInputOrderId, orderIdToInsert)
			CreateFail(failId, failDateTime, device)
			continue
		}
		if openTerminalInputOrderId > 0 {
			LogInfo(device.Name, "PPA order opened")
			if productId == openOrderProductId {
				LogInfo(device.Name, "Product match")
				CreateFail(failId, failDateTime, device)
				continue
			} else {
				LogInfo(device.Name, "Product did not match")
				orderIdToInsert := GetOrderIdForProductId(device, productId)
				OpenNewTerminalInputOrder(device, orderIdToInsert, failDateTime)
				CloseOpenTerminalInputOrder(device, openTerminalInputOrderId, failDateTime)
				CreateFail(failId, failDateTime, device)
				continue
			}
		} else {
			LogInfo(device.Name, "No order opened, creating")
			orderIdToInsert := GetOrderIdForProductId(device, productId)
			OpenNewTerminalInputOrder(device, orderIdToInsert, failDateTime)
			CreateFail(failId, failDateTime, device)
		}
	}
}

func CreateFail(failId int, failDateTime time.Time, device Device) {
	var terminalInputFail TerminalInputFail
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	terminalInputFail.FailID = failId
	terminalInputFail.DeviceID = device.OID
	terminalInputFail.UserID = 1
	terminalInputFail.DT = failDateTime
	db.NewRecord(terminalInputFail)
	db.Create(&terminalInputFail)
	openTerminalInputOrderId, _ := CheckOpenTerminalInputOrder(device)
	latestTerminalInputFailId := GetLatestTerminalInputFailId(device, failId, failDateTime)
	var terminalInputOrderFail TerminalInputOrderTerminalInputFail
	terminalInputOrderFail.TerminalInputOrderID = openTerminalInputOrderId
	terminalInputOrderFail.TerminalInputFailID = latestTerminalInputFailId
	db.Create(&terminalInputOrderFail)
}

func GetLatestTerminalInputFailId(device Device, failId int, failDateTime time.Time) int {
	var terminalInputFail TerminalInputFail
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0
	}
	db.Where("DeviceID = ?", device.OID).Where("FailID = ?", failId).Where("DT = ?", failDateTime).Find(&terminalInputFail)
	return terminalInputFail.OID
}

func CloseOpenTerminalInputOrder(device Device, openTerminalInputOrderId int, failDateTime time.Time) {
	var terminalInputOrder TerminalInputOrder
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	db.Model(&terminalInputOrder).Where("OID = ?", openTerminalInputOrderId).Update("DTE", failDateTime)
}

func OpenNewTerminalInputOrder(device Device, orderIdToInsert int, failDateTime time.Time) {
	var terminalInputOrder TerminalInputOrder
	terminalInputOrder.OrderID = orderIdToInsert
	terminalInputOrder.DeviceID = device.OID
	terminalInputOrder.DTS = failDateTime
	terminalInputOrder.UserID = 1
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	db.Create(&terminalInputOrder)
}

func GetOrderIdForProductId(device Device, productId int) int {
	var order Order
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0
	}
	db.Where("ProductId = ?", productId).Find(&order)
	if order.OID > 0 {
		return order.OID
	}
	var newOrder Order
	newOrder.Name = device.Name + strconv.Itoa(productId)
	newOrder.ProductID = productId
	db.Save(&newOrder)
	var brandNewOrder Order
	db.Where("ProductId = ?", productId).Find(&brandNewOrder)
	return brandNewOrder.OID
}

func CheckProductForOpenOrder(orderId int) int {
	var order Order
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0
	}
	db.Where("OID = ?", orderId).Find(&order)
	return order.ProductID
}

func CheckOpenTerminalInputOrder(device Device) (terminalInputOrderId int, orderId int) {
	var openOrder TerminalInputOrder
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0, 0
	}
	db.Where("DeviceID = ?", device.OID).Where("DTE is null").Find(&openOrder)
	return openOrder.OID, openOrder.OrderID

}

func CheckProductInDatabase(device Device, productName string) int {
	var product Product
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0
	}
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
	defer db.Close()
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return 0
	}
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
