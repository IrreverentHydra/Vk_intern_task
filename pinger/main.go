package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	backendAPI   = "http://vk_intern_task-backend-1:8080/ping_results"
	pingInterval = 30 * time.Second
)

type PingResult struct {
	IPAddress      string   `json:"ip_address"`
	PingTime       float64  `json:"ping_time"`
	LastSuccessful string   `json:"last_successful"`
}

func main() {
	for {
		containers := getDockerContainerIPs()
		for _, ip := range containers {
			pingContainer(ip)
		}
		time.Sleep(pingInterval)
	}
}

func getDockerContainerIPs() []string {
	cmd := exec.Command("docker", "ps", "-q")
	containerIDs, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Ошибка получения списка контейнеров: %v\n", err)
		return []string{}
	}

	ids := strings.Fields(string(containerIDs))
	var ips []string
	for _, id := range ids {
		cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", id)
		output, err := cmd.CombinedOutput()
		if err == nil {
			ip := strings.TrimSpace(string(output))
			if ip != "" {
				ips = append(ips, ip)
			}
		}
	}
	return ips
}

func pingContainer(ip string) {
	cmd := exec.Command("ping", "-c", "1", ip)
	start := time.Now()
	output, err := cmd.CombinedOutput()
	pingDuration := time.Since(start)

	result := PingResult{
		IPAddress: ip,
		PingTime:  float64(pingDuration.Microseconds()),
	}

	if err == nil {
		result.LastSuccessful = time.Now().Format(time.RFC3339)
	} else {
		fmt.Printf("Ошибка пинга %s: %s\n", ip, output)
	}

	sendPingResult(result)
}


func sendPingResult(result PingResult) {
	data, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Ошибка сериализации JSON: %v\n", err)
		return
	}

	resp, err := http.Post(backendAPI, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Ошибка отправки данных: %v\n", err)
		fmt.Println("Отправляемые данные:", string(data))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Ошибка ответа от сервера: %s\n", resp.Status)
		fmt.Println("Отправляемые данные:", string(data))
	}
}
