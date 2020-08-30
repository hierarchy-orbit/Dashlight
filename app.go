/*
	DESCRIPTION

	Dashligh
		A dashboard for a lighthouse validator node.
		The goal of this application is to be a light weight
		dashboard useful on a RPi or SBC. It uses the termui
		package and does not utilize a significant amount of
		resources.

	TODO:
		1 Add slash status [DONE]
		2 Update the balance calculation. It is currently
			hard coded to format the box. [DONE]
		3 Add a database back end to allow persistent
			data storage.
			- Once persistent storage is achieved the ability
				to graph data can be added.
		4 Complete testing in order to move out of Alpha!
		5 Create a configuration file feature to save settings.

*/

package main

import (
	"bytes"
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

type BeaconValidator struct {
	Balance   int64  `json:"balance"`
	Pubkey    string `json:"pubkey"`
	Validator struct {
		ActivationEligibilityEpoch int64  `json:"activation_eligibility_epoch"`
		ActivationEpoch            int64  `json:"activation_epoch"`
		EffectiveBalance           int64  `json:"effective_balance"`
		ExitEpoch                  int64  `json:"exit_epoch"`
		Pubkey                     string `json:"pubkey"`
		Slashed                    bool   `json:"slashed"`
		WithdrawableEpoch          int64  `json:"withdrawable_epoch"`
		WithdrawalCredentials      string `json:"withdrawal_credentials"`
	} `json:"validator"`
	ValidatorIndex int64 `json:"validator_index"`
}

type health struct {
	MemoryTotal int     `json:"sys_virt_mem_total"`
	MemoryUsed  int     `json:"sys_virt_mem_used"`
	MemoryFree  int     `json:"sys_virt_mem_available"`
	SystemLoad  float64 `json:"sys_loadavg_1"`
	Version     string
	PeerCount   string
	DBSize      string
	Balance     string
	slashed     string
}

var metrics = health{}
var memGauge = widgets.NewGauge()
var menuTest = widgets.NewParagraph()
var textInfo = widgets.NewList()
var testval string

const baseURL = "http://localhost:5052"
const DBFile = "/var/lib/lighthouse/beacon-node/beacon/chain_db"

var BeaconValidators []BeaconValidator

func main() {
	BeaconValidators = append([]BeaconValidator{})
	menuTest.Text = "  PRESS q TO QUIT"
	menuTest.SetRect(0, 0, 50, 4)
	menuTest.TextStyle.Fg = ui.ColorWhite
	menuTest.BorderStyle.Fg = ui.ColorCyan
	menuTest.Title = "\tDashLight Ethereum Validator Monitor"
	getVersion()
	menuTest.Text = menuTest.Text + "\nVer. = " + metrics.Version

	// Text information

	textInfo.Rows = []string{
		"Public Key  : XXX",
		"Is Slashed  : Unknown",
		"Node Balance: 00000 ETH",
		"System Load : ",
		"Peer Count  : ",
		"DB Size     : ",
	}

	textInfo.SetRect(0, 5, 50, 5+len(textInfo.Rows)+2)
	textInfo.TextStyle = ui.NewStyle(ui.ColorYellow)

	// Gauge to show percent memory usage
	memGauge.Title = "System Memory Usage"
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
	getBalance()
	if int(float64(metrics.MemoryUsed)/float64(metrics.MemoryTotal)*100) > 75 {
		memGauge.BarColor = ui.ColorRed
	} else {
		memGauge.BarColor = ui.ColorGreen
	}

	memGauge.Percent = int(float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal) * 100)

	if BeaconValidators[0].Validator.Slashed {
		textInfo.TextStyle = ui.NewStyle(ui.ColorRed)
	} else {
		textInfo.TextStyle = ui.NewStyle(ui.ColorYellow)
	}

	textInfo.Rows = []string{
		"Public Key  : " + BeaconValidators[0].Pubkey,
		"Is Slashed  : " + strconv.FormatBool(BeaconValidators[0].Validator.Slashed),
		"Node Balance: " + IntToString(BeaconValidators[0].Balance) + " ETH",
		"System Load : " + FloatToString(metrics.SystemLoad),
		"Peer Count  : " + metrics.PeerCount,
		"DB Size     : " + metrics.DBSize,
	}

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

func IntToString(input_num int64) string {
	// to convert a int number to a string and also gwei to ether
	num := int(input_num / 1000000)
	return FloatToString(float64(num) / 1000)
	//return strconv.Itoa(input_num / 1000)
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

	metrics.DBSize = strconv.FormatInt(getDirSize(DBFile)/1024/1024/1024, 10) + " GB"
}

func getDirSize(path string) int64 {
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

func getBalance() {

	jsonKeys := `{"pubkeys": ["0x90f70a6bbf31d38bb4e95a53ba87fc062b8858dcc45ec7c77174e891679f4e4edc2e6efb6f38aa11c7c66249c62cacdd"]}`

	resp, err := http.Post(baseURL+"/beacon/validators", "application/json", bytes.NewBuffer([]byte(jsonKeys)))
	if err != nil {
		log.Fatalln(err)
	} else {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &BeaconValidators)
		testval = string(bodyBytes)

	}

	defer resp.Body.Close()
}
