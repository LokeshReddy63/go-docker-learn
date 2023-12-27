package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ServerInfo struct {
	Status       string `json:"status"`
	Hostname     string `json:"hostname"`
	HostIP       string `json:"host_ip"`
	OtherContent string `json:"other_content,omitempty"`
}

// Struct to store details of each request
type RequestDetails struct {
	IP      string    `json:"ip"`
	Time    time.Time `json:"time"`
	Content string    `json:"content"`
}

// Path to the history file inside the volume
const historyFilePath = "./log/history.json"

func main() {
	// Define a handler function for the / endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get hostname
		hostname, err := os.Hostname()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get host IP address
		hostIP, err := getHostIP()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a ServerInfo struct with the response details
		info := ServerInfo{
			Status:   "ok",
			Hostname: hostname,
			HostIP:   hostIP,
		}

		// If it's a direct call to the API, include history in OtherContent
		if r.Method == http.MethodGet {
			info.OtherContent = getHistoryContent()
		}

		// Convert the struct to JSON
		jsonResponse, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header to application/json
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON response
		_, err = w.Write(jsonResponse)
		if err != nil {
			fmt.Println("Error writing JSON response:", err)
		}

		// Log the request details to the history
		logRequestDetails(r)
	})

	// Start the server on port 8000
	port := 8000
	fmt.Printf("Server is running on port %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error starting the server:", err)
	}
}

func getHostIP() (string, error) {
	// Retrieve the host's IP address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		// Check if the address is not a loopback address and is IPv4
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("Unable to determine host IP address")
}

func getHistoryContent() string {
	// Read the history file
	data, err := os.ReadFile(historyFilePath)
	if err != nil {
		return fmt.Sprintf("Error reading history file: %v", err)
	}

	// Unmarshal the JSON data into a slice of RequestDetails
	var history []RequestDetails
	if err := json.Unmarshal(data, &history); err != nil {
		return fmt.Sprintf("Error unmarshalling history: %v", err)
	}

	// Generate a string with each log line on a new line
	var content string
	for _, req := range history {
		content += fmt.Sprintf(":: %s - %s - %s ::", req.Time.Format(time.RFC3339), req.Content, req.IP)
	}

	return content
}

func logRequestDetails(r *http.Request) {
	// Get the IP address of the requestor
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Println("Error extracting IP address:", err)
		return
	}

	// Create a new request details entry
	requestDetails := RequestDetails{
		IP:      ip,
		Time:    time.Now(),
		Content: "Get Requested",
	}

	// Append the request details to the history
	appendHistory(requestDetails)
}

func appendHistory(details RequestDetails) {
	// Read existing history from the file or create a new one
	var history []RequestDetails
	data, err := os.ReadFile(historyFilePath)
	if err == nil {
		err = json.Unmarshal(data, &history)
		if err != nil {
			fmt.Println("Error unmarshalling history:", err)
		}
	}

	// Append the new request details
	history = append(history, details)

	// Marshal the history struct into JSON
	data, err = json.Marshal(history)
	if err != nil {
		fmt.Println("Error marshalling history:", err)
		return
	}

	// Ensure the directory structure exists before writing the file
	if err := os.MkdirAll(filepath.Dir(historyFilePath), 0755); err != nil {
		fmt.Println("Error creating directory structure:", err)
		return
	}

	// Write the JSON data to the history file
	err = os.WriteFile(historyFilePath, data, 0644)
	if err != nil {
		fmt.Println("Error writing history file:", err)
	}
}
