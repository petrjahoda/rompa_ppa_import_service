# ROMPA PPA Import Service


## Installation
* use docker image from https://hub.docker.com/r/petrjahoda/rompa_ppa_import_service
* use linux, mac or windows (using nssm) version and make it run like a service

## Settings
Settings are read from config.json.<br>
This file is created with default values, when not found.
* DatabaseType: "mysql"
* IpAddress:    "zapsidatabase"
* DatabaseName: "zapsi2"
* Port:         "3306"
* Login:        "zapsi_uzivatel"
* Password:     "zapsi"

## Tables used
* READ Device (deviceId=1000)
* READ/WRITE Fail
* READ/WRITE Product
* READ/WRITE Order
* READ/WRITE Terminal_input_order
* READ/WRITE Terminal_input_fail

## Description
Go service that downloads data every 10 seconds from mapped network file for every device that is DeviceType=1000.<br>
File is processed line by line and new product, order and fails are created.




www.zapsi.eu © 2020
