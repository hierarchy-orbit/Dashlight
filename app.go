/*
	DESCRIPTION

	Dashligh
		A dashboard for a lighthouse validator node.
		The goal of this application is to be a light weight
		dashboard useful on a RPi or SBC. It uses the termui
		package and does not utilize a significant amount of
		resources.

	TODO:
		1 Add slash status
		2 Update the balance calculation. It is currently
			hard coded to format the box.
		3 Add a database back end to allow persistent
			data storage.
			- Once persistent storage is achieved the ability
				to graph data can be added.
		4 Complete testing in order to move out of Alpha!

*/

/*
String to check balance and slash status
curl -X POST -H "Content-Type: application/json" -d '{"pubkeys": ["0x90f70a6bbf31d38bb4e95a53ba87fc062b8858dcc45ec7c77174e891679f4e4edc2e6efb6f38aa11c7c66249c62cacdd"]}' http://localhost:5052/beacon/validators

[
  {
    "pubkey": "0x90f70a6bbf31d38bb4e95a53ba87fc062b8858dcc45ec7c77174e891679f4e4edc2e6efb6f38aa11c7c66249c62cacdd",
    "validator_index": 26595,
    "balance": 31204700300,
    "validator": {
      "pubkey": "0x90f70a6bbf31d38bb4e95a53ba87fc062b8858dcc45ec7c77174e891679f4e4edc2e6efb6f38aa11c7c66249c62cacdd",
      "withdrawal_credentials": "XXXXX",
      "effective_balance": 31000000000,
      "slashed": false,
      "activation_eligibility_epoch": 631,
      "activation_epoch": 1665,
      "exit_epoch": 18446744073709552000,
      "withdrawable_epoch": 18446744073709552000
    }
  }
]


*/

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type health struct {
	MemoryTotal int     `json:"sys_virt_mem_total"`
	MemoryUsed  int     `json:"sys_virt_mem_used"`
	MemoryFree  int     `json:"sys_virt_mem_available"`
	SystemLoad  float64 `json:"sys_loadavg_1"`
	Version     string
	PeerCount   string
	DBSize      string
}


var metrics = health{}
var memGauge = widgets.NewGauge()
var menuTest = widgets.NewParagraph()
var textInfo = widgets.NewList()
var baseURL = "http://localhost:5052"
var DBFile = "/var/lib/lighthouse/beacon-node/beacon/chain_db"

func main() {

	// Display the Help Message
	getVersion()
	menuTest.Title = metrics.Version
	menuTest.Text = "PRESS q TO QUIT"
	menuTest.SetRect(0, 0, 50, 4)
	menuTest.TextStyle.Fg = ui.ColorWhite
	menuTest.BorderStyle.Fg = ui.ColorCyan

	// Text information

	textInfo.Rows = []string{
		"Node Balance: 31.214355995 ETH",
		"System Load : ",
		"Peer Count  : ",
		"DB Size     : ",
	}

	textInfo.SetRect(0, 5, 50, 5+len(textInfo.Rows)+2)
	textInfo.TextStyle = ui.NewStyle(ui.ColorYellow)

	// Gauge to show percent memory usage
	memGauge.Title = "Mem Usage"
	memGauge.SetRect(0, len(textInfo.Rows)+9, 50, len(textInfo.Rows)+12)
	memGauge.Percent = 0
	memGauge.BarColor = ui.ColorGreen
	memGauge.BorderStyle.Fg = ui.ColorWhite
	memGauge.TitleStyle.Fg = ui.ColorCyan

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	showMemory()

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second * 5).C
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		case <-ticker:
			showMemory()
		}
	}

}

func showMemory() {
	getHealth()
	getPeers()
	getDBSize()
	if int(float64(metrics.MemoryUsed)/float64(metrics.MemoryTotal)*100) > 75 {
		memGauge.BarColor = ui.ColorRed
	} else {
		memGauge.BarColor = ui.ColorGreen
	}

	memGauge.Percent = int(float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal) * 100)

	textInfo.Rows = []string{
		"Node Balance: 31.214355995 ETH",
		"System Load : " + FloatToString(metrics.SystemLoad),
		"Peer Count  : " + metrics.PeerCount,
		"DB Size     : " + metrics.DBSize,
	}

	// textInfo.Text = "System Load: " + FloatToString(metrics.SystemLoad) + "\n" + "Peer Count: " + metrics.PeerCount

	ui.Render(menuTest)
	ui.Render(textInfo)
	ui.Render(memGauge)
}

func getHealth() {
	nodeHealth := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/node/health", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "LightDash")

	res, getErr := nodeHealth.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	jsonErr := json.Unmarshal(body, &metrics)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
}

func getPeers() {
	res, err := http.Get(baseURL + "/network/peer_count")
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	metrics.PeerCount = string(content)
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', -1, 32)
}

func getVersion() {
	res, err := http.Get(baseURL + "/node/version")
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	metrics.Version = string(content)
}

func getDBSize() {

	metrics.DBSize = strconv.FormatInt(DirSize(DBFile)/1024/1024/1024, 10) + " GB"
}

func DirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size
}
