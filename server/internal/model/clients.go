package model

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// ReadMapFromFile reads a map from a JSON file
func ReadMapFromFile(filename string) (map[string]string, error) {
	// Read file contents
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Unmarshal JSON data into a map
	var data map[string]string
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return data, nil
}

func WriteMapToFile(filename string, data map[string]string) error {
	// Convert the map to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	// Write JSON data to file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

const ClientPkFilePath = "resources/ClientPk.json"
const PkCLientFilePath = "resources/PkClient.json"

type ClientsMap struct {
	ClientPK map[string]string
	Clients  map[string]*Client
	PkClient map[string]string
	mu       sync.RWMutex
}

func NewClientsMap() *ClientsMap {
	ret := new(ClientsMap)
	m, err := ReadMapFromFile(ClientPkFilePath)
	if err != nil {
		ret.ClientPK = make(map[string]string)
	} else {
		ret.ClientPK = m
	}

	m, err = ReadMapFromFile(PkCLientFilePath)
	if err != nil {
		ret.PkClient = make(map[string]string)
	} else {
		ret.PkClient = m
	}
	ret.Clients = make(map[string]*Client)
	for key := range ret.ClientPK {
		ret.Clients[key] = NewClient(key, ret.ClientPK[key])
	}
	return ret
}
func AddClient(m *ClientsMap, client *Client) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.ClientPK[client.Username]; exists {
		return false
	}
	m.Clients[client.Username] = client
	m.ClientPK[client.Username] = client.PublicKey
	m.PkClient[client.PublicKey] = client.Username
	WriteMapToFile(PkCLientFilePath, m.PkClient)
	WriteMapToFile(ClientPkFilePath, m.ClientPK)
	return true
}
func AddClientAsync(m *ClientsMap, client *Client) bool {
	resultChan := make(chan bool)
	go func() {
		result := AddClient(m, client)
		resultChan <- result
	}()
	return <-resultChan
}
func GetClient(m *ClientsMap, UserPk string) *Client {
	if _, exists := m.ClientPK[UserPk]; exists {
		return m.Clients[UserPk]
	}
	if username, exists := m.PkClient[UserPk]; exists {
		return m.Clients[username]
	}
	return nil
}
