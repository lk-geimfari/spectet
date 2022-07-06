package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tatsushid/go-fastping"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const TasksApiURL = "https://isaak.dev/index.json"
const CreateTaskApiURL = "https://httpbin.org/post"
const PacketBodyMessage = "Spectet Healtcheck"

type Task struct {
	TaskId   string `json:"task_id"`
	TaskType string `json:"task_type"`
	Hostname string `json:"hostname"`
	Port     uint16 `json:"port"`
}

// CheckHTTP Checks if host from the given Task is available over HTTP.
func (task Task) CheckHTTP() bool {
	var uri = task.Hostname
	if !strings.HasPrefix(uri, "https://") {
		uri = fmt.Sprintf("https://%s/", uri)
	}
	response, err := http.Get(uri)
	if err != nil {
		return false
	}
	return response.StatusCode >= 200 && response.StatusCode < 300
}

// CheckUDP Checks if host from the given Task is available over UDP.
func (task Task) CheckUDP() bool {
	var port = strconv.FormatUint(uint64(task.Port), 10)
	var address = fmt.Sprintf("%s:%s", task.Hostname, port)
	// Resolve the UDP address so that we can make use of
	// DialUDP with an actual IP.
	RemoteAddr, _ := net.ResolveUDPAddr("udp", address)

	connection, err := net.DialUDP("udp", nil, RemoteAddr)

	defer connection.Close()

	if err != nil {
		return false
	}

	// write a message to server
	var message = []byte(PacketBodyMessage)
	_, err = connection.Write(message)

	if err != nil {
		return false
	}

	// Receive response from server
	buf := make([]byte, len(PacketBodyMessage))
	amountByte, _, err := connection.ReadFromUDP(buf)

	if err != nil {
		return false
	}
	return len(message) == amountByte
}

// CheckTCP Checks if host from the given Task is available over TCP.
func (task Task) CheckTCP() bool {
	var port = strconv.FormatUint(uint64(task.Port), 10)
	var address = fmt.Sprintf("%s:%s", task.Hostname, port)
	_, err := net.Dial("tcp", address)
	if err != nil {
		return false
	}
	return true
}

// CheckDNS Checks if hostname from the given Task resolves IP/CNAME/NS over DNS.
func (task Task) CheckDNS() bool {
	var ips, err = net.LookupIP(task.Hostname)
	if err != nil || len(ips) == 0 {
		return false
	}

	cname, err := net.LookupCNAME(task.Hostname)
	if err != nil || len(cname) == 0 {
		return false
	}

	nameserver, err := net.LookupNS(task.Hostname)
	if err != nil || len(nameserver) == 0 {
		return false
	}

	return true
}

// CheckICMP Checks if hostname from the given Task is responding to ping over ICMP.
func (task Task) CheckICMP() bool {
	var pinger = fastping.NewPinger()
	resolvedAddress, err := net.ResolveIPAddr("ip:icmp", task.Hostname)

	if err != nil {
		fmt.Println(err)
		return false
	}

	var icmpStatus = false
	pinger.AddIPAddr(resolvedAddress)
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		icmpStatus = true
		fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}
	pinger.OnIdle = func() {
		icmpStatus = true
		fmt.Println("ICMP Checked!")
	}
	err = pinger.Run()

	if err != nil {
		fmt.Println(err)
		return false
	}
	return icmpStatus
}

// Perform Performs appropriate check for given Task.
func (task Task) Perform() bool {
	var status bool

	if task.TaskType == "tcp" {
		status = task.CheckTCP()
	} else if task.TaskType == "http" {
		status = task.CheckHTTP()
	} else if task.TaskType == "udp" {
		status = task.CheckUDP()
	} else if task.TaskType == "resolve" {
		status = task.CheckDNS()
	} else if task.TaskType == "icmp" {
		status = task.CheckICMP()
	}

	log.Printf("%s over %s is accesesible: %t\n", task.Hostname, task.TaskType, status)
	return status
}

type Tasks []Task

// SaveReport Saves the result of checking in the database.
func (task Task) SaveReport(result any) {

	var report = map[string]any{
		"task_id":     task.TaskId,
		task.TaskType: result,
	}
	jsonData, _ := json.Marshal(report)

	_, err := http.Post(CreateTaskApiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
}

// FetchTasks Fetches tasks to perform.
func FetchTasks() Tasks {
	response, err := http.Get(TasksApiURL)
	if err != nil {
		log.Println("No response from request")
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	var result Tasks
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println("Can not unmarshal JSON")
	}
	return result
}

func main() {
	for _, task := range FetchTasks() {
		var report = task.Perform()
		task.SaveReport(report)
	}
}
