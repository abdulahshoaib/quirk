package utils

import (
	"encoding/csv"
	"fmt"
	// "log"
	"os"
	"strconv"
	"strings"
)

func readCsvFile(filePath string) ([][]interface{}, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read input file %s: %v", filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse file as CSV for %s: %v", filePath, err)
	}

	return Convertdatatype(records), nil
}

func convertType(s string) interface{} {

	if strings.ToLower(s) == "true" {
		return true
	}
	if strings.ToLower(s) == "false" {
		return false
	}

	if i, err := strconv.Atoi(s); err == nil {
		// fmt.Printf("Converted %d to int\n", i)
		return i
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		// fmt.Printf("Converted %f to float", f)
		return f
	}

	return s
}

func Convertdatatype(str [][]string) [][]interface{} {
	if len(str) == 0 {
		return nil
	}

	var result [][]interface{}

	headerrow := make([]interface{}, len(str[0]))
	for i, val := range str[0] {
		headerrow[i] = val
	}
	result = append(result, headerrow)

	for _, row := range str[1:] {
		convertedrow := make([]interface{}, len(row))
		for i, val := range row {
			convertedrow[i] = convertType(val)
		}
		result = append(result, convertedrow)
	}

	return result
}
func printCSV(csvrecords [][]interface{}) {

	fmt.Println("CSV Data:")
	for _, row := range csvrecords {
		fmt.Println(row)
	}
	fmt.Println()
}

// func main() {
// 	csvrecords, err := readCsvFile("data.csv")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	printCSV(csvrecords)

// }
