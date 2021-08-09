package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/joho/godotenv"
)

type GSheetData struct {
	Version  string `json:"version"`
	Encoding string `json:"encoding"`
	Feed struct {
		Entry    []struct {
			Cell struct {
				Row        string    `json:"row"`
				Col        string    `json:"col"`
				Value			 string 	 `json:"inputValue"`
			} `json:"gs$cell"`
		} `json:"entry"`
	} `json:"feed"`
}

type Payload struct {
	Data []Row
}
type Row struct {
	name	string
	comment string
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	response, err := getJsonResponse()
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, string(response))
}

func main() {
	godotenv.Load()
	router := httprouter.New()
	router.GET("/", Index)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func fetchData(url string) *http.Response {
	resp, _ := http.Get(url)

	return resp
}

func getJsonResponse() ([]byte, error) {
	url := os.Getenv("GOOGLE_SHEET_URL")
	response := fetchData(url)

	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	var gsheetData GSheetData
  json.Unmarshal([]byte(data), &gsheetData)

	return buildResponse(gsheetData)
}

func buildResponse(gsheetData GSheetData) ([]byte, error) {
	keys := getKeys(gsheetData)
	result := buildResult(keys, gsheetData)

	js, err := json.Marshal(result)

	return js, err
}

func getKeys(gsheetData GSheetData) ([]string) {
	entries := gsheetData.Feed.Entry
	var keys []string
	for _, s := range entries {
		if s.Cell.Row == "1" {
			keys = append(keys, s.Cell.Value)
		}
	}

	return keys
}

func buildResult(keys []string, gsheetData GSheetData) ([]map[string]string) {
	entries := gsheetData.Feed.Entry


	chunkSize := len(keys)

	var arrData [][]string

	for i := len(keys); i < len(entries); i += chunkSize {
		var arrRow []string
		for j := 0; j < len(keys); j++ {
			dataValue := entries[i+j].Cell.Value
			arrRow = append(arrRow, dataValue)

		}
		arrData = append(arrData, arrRow)
	}

	var result []map[string]string

	for _, data := range arrData {
		row := make(map[string]string)
		for i := 0; i <  len(keys); i++ {
			row[keys[i]] = data[i]
		}
		result = append(result, row)
	}

	return result
}