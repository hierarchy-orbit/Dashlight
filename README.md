# Dashlight
Ethereum 2.0 Lighthouse Validator CLI Dashboard

## Description
A dashboard for a lighthouse validator node.
		The goal of this application is to be a light weight
		dashboard useful on a RPi or SBC. It uses the termui
		package and does not utilize a significant amount of
		resources.

Currently this software is considered pre-alpha.
USE AT YOUR OWN RISK!

![Dashlight Screencap](/assets/dashlight-top-screenshot.png)

## Features

- Dashlight will provide a simple CLI dashboard for 
- Currently Dashlight will only work with the Lighthouse client.
- Created with Go with binaries for x64 and arm64.
- Utilizes termui to provide visual widgets.

## TODO

	TODO:
		- [ ] Add slash status
		- [ ] Update the balance calculation. It is currently
			hard coded to format the box.
		- [ ] Add a database back end to allow persistent
			data storage.
			- Once persistent storage is achieved the ability
				to graph data can be added.
		- [ ] Complete testing in order to move out of *Alpha!*

## Usage

git clone https://github.com/NakamotoNetwork/Dashlight.git

\$ cd Dashlight  
\$ go install  
\$ go build  
\$ ./dashlight  
