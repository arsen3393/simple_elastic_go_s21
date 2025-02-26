package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const csvFile = "../../materials/data.csv"
const mappingFile = "./schema.json"
const indexName = "places"

func main() {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatal("Error in creating clinet", err)
	}
	fmt.Println("parsing files...")
	data, err := parseCsvFile(csvFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("parsing files completed\n")
	mapping, err := readMappingFile(mappingFile)
	if err != nil {
		log.Fatal(err)
	}
	createIndexMapping(es, mapping)
	fmt.Println("loading data...\n")
	err = loadData(es, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("done")
}

func createIndexMapping(es *elasticsearch.Client, mapping string) {
	if response, err := es.Indices.Delete([]string{indexName},
		es.Indices.Delete.WithIgnoreUnavailable(true)); err != nil || response.IsError() {
		log.Fatal("Can't delete index", err) // удаляем если уже существует такой индекс
	}

	res, err := es.Indices.Create(indexName, es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		log.Fatal("Can't create index", err)
	}
	if res.IsError() {
		log.Fatal("Can't create index", res)
	}
	defer func() { _ = res.Body.Close() }()
}

type Data struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Phone    string   `json:"phone"`
	Location location `json:"location"`
}

type location struct {
	Longitude float64 `json:"lon"`
	Latitude  float64 `json:"lat"`
}

func readMappingFile(path string) (string, error) {
	const op = "readMappingFile function process"
	res, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New(op + ": " + err.Error())
	}
	return string(res), nil
}

func parseCsvFile(path string) ([]Data, error) {
	const op = "parseCsvFile function process"
	var result []Data
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New(op + ": " + err.Error())
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	_, _ = reader.Read() // пропускаем первую строку
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.New(op + ": " + err.Error())
		}
		data, err := makeData(record)
		if err != nil {
			return nil, errors.New(op + ": " + err.Error())
		}
		result = append(result, data)
	}
	return result, nil
}

func makeData(record []string) (Data, error) {
	const op = "makeData function process"
	if len(record) != 6 {
		return Data{}, fmt.Errorf("Invalid person slice: %v", record)
	}
	lon, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return Data{}, errors.New(op + ": " + err.Error())
	}
	lat, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return Data{}, errors.New(op + ": " + err.Error())
	}
	return Data{
		Id:      record[0],
		Name:    record[1],
		Address: record[2],
		Phone:   record[3],
		Location: location{
			Longitude: lon,
			Latitude:  lat,
		},
	}, nil
}

func makeConfig(es *elasticsearch.Client) esutil.BulkIndexerConfig {
	return esutil.BulkIndexerConfig{
		Index:         indexName,
		Client:        es,
		NumWorkers:    2,
		FlushBytes:    5000,
		FlushInterval: time.Second * 30,
	}
}

func loadData(es *elasticsearch.Client, data []Data) error {
	var placeHolder uint64
	const op = "loadData function process"
	bi, err := esutil.NewBulkIndexer(makeConfig(es))
	if err != nil {
		return errors.New(op + ": " + err.Error())
	}
	for _, d := range data {
		dInfo, err := json.Marshal(d)
		if err != nil {
			return errors.New("cannot marshaling data in" + op + ": " + err.Error())
		}
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: d.Id,
				Body:       bytes.NewReader(dInfo),
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&placeHolder, 1)
				},
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Println("ERROR:", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
	}
	err = bi.Close(context.Background())
	if err != nil {
		return errors.New(op + ": " + err.Error())
	}
	return nil
}
