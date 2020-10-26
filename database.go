package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"time"
)

type Fail struct {
	OID        int    `gorm:"column:OID"`
	Name       string `gorm:"column:Name"`
	Barcode    string `gorm:"column:Barcode"`
	FailTypeID int    `gorm:"column:FailTypeID"`
}

func (Fail) TableName() string {
	return "fail"
}

type Device struct {
	OID        int    `gorm:"column:OID"`
	CustomerID int    `gorm:"column:CustomerID"`
	Name       string `gorm:"column:Name"`
	DeviceType int    `gorm:"column:DeviceType"`
	Setting    string `gorm:"column:Setting"`
}

func (Device) TableName() string {
	return "device"
}

func (device Device) UpdateTerminalInputOrder(openTerminalInputOrderId int, orderIdToInsert int) {
	var terminalInputOrder TerminalInputOrder
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	db.Model(&terminalInputOrder).Where("OID = ?", openTerminalInputOrderId).Update("OrderID", orderIdToInsert)
}

type TerminalInputOrder struct {
	OID      int       `gorm:"column:OID"`
	DTS      time.Time `gorm:"column:DTS"`
	DTE      time.Time `gorm:"column:DTE; default:null"`
	Interval float32   `gorm:"column:Interval"`
	OrderID  int       `gorm:"column:OrderID"`
	UserID   int       `gorm:"column:UserID"`
	DeviceID int       `gorm:"column:DeviceID"`
}

func (TerminalInputOrder) TableName() string {
	return "terminal_input_order"
}

type TerminalInputOrderTerminalInputFail struct {
	TerminalInputOrderID int `gorm:"column:TerminalInputOrderID"`
	TerminalInputFailID  int `gorm:"column:TerminalInputFailID"`
}

func (TerminalInputOrderTerminalInputFail) TableName() string {
	return "terminal_input_order_terminal_input_fail"
}

type User struct {
	OID   int    `gorm:"column:OID"`
	Login string `gorm:"column:Login"`
}

func (User) TableName() string {
	return "user"
}

type Order struct {
	OID       int    `gorm:"column:OID"`
	Name      string `gorm:"column:Name"`
	ProductID int    `gorm:"column:ProductID"`
}

func (Order) TableName() string {
	return "order"
}

type Product struct {
	OID             int    `gorm:"column:OID"`
	Name            string `gorm:"column:Name"`
	Barcode         string `gorm:"column:Barcode"`
	ProductStatusID int    `gorm:"column:ProductStatusID"`
}

func (Product) TableName() string {
	return "product"
}

type Package struct {
	OID           int    `gorm:"column:OID"`
	Barcode       int    `gorm:"column:Barcode"`
	PackageTypeID int    `gorm:"column:PackageTypeID"`
	OrderID       string `gorm:"column:OrderID"`
}

func (Package) TableName() string {
	return "package"
}

type TerminalInputFail struct {
	OID      int       `gorm:"column:OID"`
	DT       time.Time `gorm:"column:DT"`
	FailID   int       `gorm:"column:FailID"`
	UserID   int       `gorm:"column:UserID; default:null"`
	DeviceID int       `gorm:"column:DeviceID"`
}

func (TerminalInputFail) TableName() string {
	return "terminal_input_fail"
}

func CheckDatabase() bool {
	var connectionString string
	var dialect string
	if DatabaseType == "postgres" {
		connectionString = "host=" + DatabaseIpAddress + " sslmode=disable port=" + DatabasePort + " user=" + DatabaseLogin + " dbname=" + DatabaseName + " password=" + DatabasePassword
		dialect = "postgres"
	} else if DatabaseType == "mysql" {
		connectionString = DatabaseLogin + ":" + DatabasePassword + "@tcp(" + DatabaseIpAddress + ":" + DatabasePort + ")/" + DatabaseName + "?charset=utf8&parseTime=True&loc=Local"
		dialect = "mysql"
	}
	db, err := gorm.Open(dialect, connectionString)
	defer db.Close()
	if err != nil {
		LogWarning("MAIN", "Database zapsi2 does not exist")
		return false
	}
	LogDebug("MAIN", "Database zapsi2 exists")
	return true
}

func CheckDatabaseType() (string, string) {
	var connectionString string
	var dialect string
	if DatabaseType == "postgres" {
		connectionString = "host=" + DatabaseIpAddress + " sslmode=disable port=" + DatabasePort + " user=" + DatabaseLogin + " dbname=" + DatabaseName + " password=" + DatabasePassword
		dialect = "postgres"
	} else if DatabaseType == "mysql" {
		connectionString = DatabaseLogin + ":" + DatabasePassword + "@tcp(" + DatabaseIpAddress + ":" + DatabasePort + ")/" + DatabaseName + "?charset=utf8&parseTime=True&loc=Local"
		dialect = "mysql"
	}
	return connectionString, dialect
}
