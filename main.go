package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://chromasdk.io:54236"

var quit = make(chan bool)
var sessionID int

type appCreationRequest struct {
	Title           string                   `json:"title"`
	Description     string                   `json:"description"`
	Author          appCreationRequestAuthor `json:"author"`
	DeviceSupported []string                 `json:"device_supported"`
	Category        string                   `json:"category"`
}

type appCreationRequestAuthor struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
}

type appCreationResponse struct {
	SessionID int    `json:"sessionid"`
	URI       string `json:"uri"`
}

type effectCreationRequest struct {
	Effect string      `json:"effect"`
	Param  effectParam `json:"param"`
}

type effectCreationResponse struct {
	ID     string `json:"id"`
	Result int    `json:"result"`
}

type effectParam struct {
	Color int `json:"color"`
}

type effectApplyRequest struct {
	ID string `json:"id"`
}

func main() {
	createApp()

	go pingHeartbeat()

	fmt.Println(sessionID)

	createAndApplyEffect(200)

	// TODO Test latency/request limits

	<-quit // Keep the program alive until we kill it with a keyboard shortcut
}

func makeRequest(method string, url string, body []byte) (*http.Response, error) {
	fmt.Println("URL:>", url)

	// TODO Do I need to do anything special to handle not passing a body?
	var jsonStr = []byte(body)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func structToBytes(theStruct interface{}) []byte {
	resultString, err := json.Marshal(theStruct)
	if err != nil {
		panic(err)
	}

	return resultString
}

func pingHeartbeat() {
	// TODO Make a way to end this
	for range time.Tick(time.Second * 1) {
		_, err := makeRequest(http.MethodPut, getSessionURL()+"/heartbeat", nil)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func getSessionURL() string {
	return baseURL + "/sid=" + strconv.Itoa(sessionID)
}

func createApp() {
	app := appCreationRequest{
		Title:       "Razer Chroma Go Wrapper",
		Description: "Poots",
		Author: appCreationRequestAuthor{
			Name:    "Jesse Millar",
			Contact: "jessemillar.com",
		},
		DeviceSupported: []string{
			"keyboard",
			"mouse",
			"headset",
			"mousepad",
			"keypad",
			"chromalink",
		},
		Category: "application",
	}

	resp, err := makeRequest(http.MethodPost, getSessionURL()+"/razer/chromasdk", structToBytes(app))
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var data appCreationResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err.Error())
	}

	sessionID = data.SessionID
}

func createAndApplyEffect(color int) {
	effect := effectCreationRequest{
		Effect: "CHROMA_STATIC",
		Param: effectParam{
			Color: color,
		},
	}

	effectID := createEffect(effect)
	applyEffect(effectID)
}

func createEffect(effect effectCreationRequest) string {
	resp, err := makeRequest(http.MethodPost, getSessionURL()+"/chromalink", structToBytes(effect))
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var data effectCreationResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err.Error())
	}

	return data.ID
}

func applyEffect(effectID string) {
	requestBody := effectApplyRequest{
		ID: effectID,
	}

	_, err := makeRequest(http.MethodPut, getSessionURL()+"/effect", structToBytes(requestBody))
	if err != nil {
		log.Fatalln(err)
	}
}
