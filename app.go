/*
Health Endpoint:
{
  "pid": 843,
  "pid_num_threads": 37,
  "pid_mem_resident_set_size": 3848122368,
  "pid_mem_virtual_memory_size": 7764897792,
  "sys_virt_mem_total": 8348585984,
  "sys_virt_mem_available": 2940694528,
  "sys_virt_mem_used": 5113991168,
  "sys_virt_mem_free": 251953152,
  "sys_virt_mem_percent": 64.77614,
  "sys_loadavg_1": 0.34,
  "sys_loadavg_5": 0.35,
  "sys_loadavg_15": 0.33
}
*/

package main

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"strconv"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type health struct {
	MemoryTotal int     `json:"sys_virt_mem_total"`
	MemoryUsed  int     `json:"sys_virt_mem_used"`
	MemoryFree  int     `json:"sys_virt_mem_available"`
	SystemLoad  float64 `json:"sys_loadavg_1"`
	Version		string
	PeerCount	string
}

/*
statusInfo := []string{
	"Connected Nodes: ",
	"System Load: ",
	"[2] interrupt.go",
	"[3] keyboard.go",
	"[4] output.go",
	"[5] random_out.go",
	"[6] dashboard.go",
	"[7] nsf/termbox-go",
}
*/

var metrics = health{}
var memGauge = widgets.NewGauge()
var menuTest = widgets.NewParagraph()
var textInfo = widgets.NewParagraph()
var healthURL = "http://localhost:5052/node/health"
var peerURL = "http://localhost:5052/network/peer_count"
var versionURL = "http://localhost:5052/node/version"

func main() {
	
	// Display the Help Message
	getVersion()
	menuTest.Title = metrics.Version
	menuTest.Text = "PRESS q TO QUIT"
	menuTest.SetRect(0, 0, 50, 4)
	menuTest.TextStyle.Fg = ui.ColorWhite
	menuTest.BorderStyle.Fg = ui.ColorCyan

	// Text information
	textInfo.Title = "Status Bar"
	textInfo.Text = "Loading..."
	textInfo.SetRect(0,5,25,10)

	// Gauge to show percent memory usage
	memGauge.Title = "Mem Usage"
	memGauge.SetRect(0, 10, 50, 13)
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
	ticker := time.NewTicker(time.Second*5).C
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
	if int(float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal)*100) > 75 {
		memGauge.BarColor = ui.ColorRed
	} else {
		memGauge.BarColor = ui.ColorGreen
	}

	memGauge.Percent = int(float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal)*100)
	textInfo.Text = "System Load: " + FloatToString(metrics.SystemLoad) + "\n" + "Peer Count: " + metrics.PeerCount

	ui.Render(menuTest)
	ui.Render(textInfo)
	ui.Render(memGauge)
}

func getHealth() {
	nodeHealth := http.Client {
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, healthURL, nil)
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
	resp, _ := http.Get(peerURL)
	bytes, _ := ioutil.ReadAll(resp.Body)

	metrics.PeerCount = string(bytes)
}

func FloatToString(input_num float64) string {
    // to convert a float number to a string
    return strconv.FormatFloat(input_num, 'f', -1, 32)
}

func getVersion() {
	resp, _ := http.Get(versionURL)
	bytes, _ := ioutil.ReadAll(resp.Body)

	metrics.Version = string(bytes)
}