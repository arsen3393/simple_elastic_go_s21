package db

import (
	"Day03/ex01/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"strings"
)

type ElasticSearchStore struct {
	Es *elasticsearch.Client
}

func (s *ElasticSearchStore) GetPlaces(limit int, offset int) ([]types.Place, int, error) {
	const op = "ElasticSearchStore.GetPlaces"

	query := types.Limits{
		Size: limit,
		From: offset,
	}
	queryJson, err := json.Marshal(query)
	if err != nil {
		return nil, 0, errors.New(op + ": " + err.Error())
	}
	fmt.Printf("Query JSON: %s\n", string(queryJson))
	req := esapi.SearchRequest{
		Index:          []string{"places"},
		Body:           strings.NewReader(string(queryJson)),
		TrackTotalHits: true,
	}
	res, err := req.Do(context.Background(), s.Es)
	if err != nil {
		return nil, 0, errors.New(op + ": " + err.Error())
	}
	defer func() { _ = res.Body.Close() }()
	if res.IsError() {
		return nil, 0, errors.New(op + ": " + res.Status())
	}
	var resBody map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return nil, 0, errors.New(op + ": " + err.Error())
	}
	hits := resBody["hits"].(map[string]interface{})["hits"].([]interface{})
	places := make([]types.Place, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		placeBytes, err := json.Marshal(source)
		if err != nil {
			continue
		}
		var place types.Place
		if err := json.Unmarshal(placeBytes, &place); err != nil {
			continue
		}
		places = append(places, place)
	}
	totalHits := int(resBody["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	return places, totalHits, nil
}

func NewElasticSearchStore() (*ElasticSearchStore, error) {
	const op = "In NewElasticSearchStore"
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, errors.New(op + ": " + err.Error())
	}
	return &ElasticSearchStore{es}, nil
}
