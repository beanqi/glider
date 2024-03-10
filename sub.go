package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type VmessInfo struct {
	V    string `json:"v"`
	Ps   string `json:"ps"`
	Add  string `json:"add"`
	Port string `json:"port"`
	Id   string `json:"id"`
	Aid  string `json:"aid"`
	Net  string `json:"net"`
	Type string `json:"type"`
	Host string `json:"host"`
	Path string `json:"path"`
	Tls  string `json:"tls"`
}

func SetForwards(conf *Config) {
	for _, subUrl := range conf.Subs {
		lines, err := ParseSub(subUrl)
		if err != nil {
			fmt.Printf("Error: %v", err)
			continue
		}
		for _, line := range lines {
			conf.Forwards = append(conf.Forwards, line)
		}
	}
}

func ParseSub(subUrl string) ([]string, error) {
	// Get http request to get subscription
	resp, err := http.Get(subUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Base64 decode the body
	decoded, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return nil, err
	}

	// Split the decoded body by lines
	lines := strings.Split(string(decoded), "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		decodedLine, err := decodeLine(line)
		if err != nil || line == "" {
			continue
		}
		result = append(result, decodedLine)
	}
	return result, nil
}

func decodeLine(line string) (string, error) {
	// every line is like "ss://YWVzLTEyOC1nY206ZDAwNThjZWMtZDk5Yi00Y2Q1LWI1ODEtMjJmMGU3MjczYTkx@8.8.8.8:80#%E5%89%A9%E4%BD%9xxxxx%9A17.22%20GB"
	// we need to decode YWVzLTEyOC1nY206ZDAwNThjZWMtZDk5Yi00Y2Q1LWI1ODEtMjJmMGU3MjczYTkx using base64, the content is between ss:// and @
	if strings.HasPrefix(line, "ss://") {
		ss := strings.Split(line, "ss://")[1]
		ss = strings.Split(ss, "@")[0]
		decoded, err := base64.StdEncoding.DecodeString(ss)
		if err != nil {
			return "", err
		}
		return line[:strings.Index(line, "ss://")+5] + string(decoded) + line[strings.Index(line, "@"):], nil
	}

	// Add more cases here for other prefixes like "vmess://"
	if strings.HasPrefix(line, "vmess://") {
		vmess := strings.Split(line, "vmess://")[1]
		decoded, err := base64.StdEncoding.DecodeString(vmess)
		if err != nil {
			return "", err
		}

		var info VmessInfo
		err = json.Unmarshal(decoded, &info)
		if err != nil {
			return "", err
		}

		// Generate the new vmess URL
		newUrl := fmt.Sprintf("vmess://%s@%s:%s", info.Id, info.Add, info.Port)
		if info.Net == "ws" {
			newUrl = fmt.Sprintf("ws://%s:%s%s?host=%s,vmess://%s", info.Add, info.Port, info.Path, info.Host, info.Id)
		}
		if info.Tls != "" {
			newUrl = fmt.Sprintf("tls://%s:%s,ws://,vmess://%s", info.Host, info.Port, info.Id)
		}

		return newUrl, nil
	}

	// TODO: Add more cases here for other prefixes like "trojan://"

	return line, nil
}
