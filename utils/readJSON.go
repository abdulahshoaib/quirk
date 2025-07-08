package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func readJsonFile(filePath string) ([]map[string]interface{}, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read input file %s: %v", filePath, err)
	}
	defer f.Close()

	var records []map[string]interface{}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&records); err != nil {
		return nil, fmt.Errorf("unable to parse file as JSON for %s: %v", filePath, err)
	}

	return records, nil
}

func printJSONRecords(records []map[string]interface{}) {
	fmt.Println("JSON Data:")
	for i, record := range records {
		fmt.Printf("Record %d:\n", i+1)
		for key, value := range record {
			fmt.Printf("  %s: %v\n", key, value)
		}
		fmt.Println()
	}
}

// func main() {

// 	jsonrecords, err := readJsonFile("data.json")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	printJSONRecords(jsonrecords)
// }
