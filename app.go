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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type health struct {
	MemoryTotal int     `json:"sys_virt_mem_total"`
	MemoryUsed  int     `json:"sys_virt_mem_used"`
	MemoryFree  int     `json:"sys_virt_mem_available"`
	SystemLoad  float32 `json:"sys_loadavg_1"`
}

func main() {
	url := "http://localhost:5052/node/health"

	nodeHealth := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
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

	metrics := health{}
	jsonErr := json.Unmarshal(body, &metrics)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	fmt.Println("Total Memory: ", metrics.MemoryTotal/1024/1024)
	fmt.Println("Used Memory : ", metrics.MemoryUsed/1024/1024)
	fmt.Println("Free Memory : ", metrics.MemoryFree/1024/1024)
	fmt.Println("System Load : ", metrics.SystemLoad)

}
