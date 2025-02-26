package main

import (
	"Day03/ex01/db"
	"Day03/ex01/types"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
)

type Store interface {
	GetPlaces(limit int, offset int) ([]types.Place, int, error)
}

var base Store

type Paginator struct {
	Places []types.Place
	Total  int
	Page   int
	Last   int
}

func main() {
	var err error
	base, err = db.NewElasticSearchStore()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", HandlerGetPlaces)
	fmt.Println("Listening on port 8888")
	err = http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func HandlerGetPlaces(w http.ResponseWriter, r *http.Request) {
	var res Paginator
	var err error
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		http.Error(w, "In HandlerGetPlacesFunc: 'page' parameter is required", http.StatusBadRequest)
		fmt.Println("In HandlerGetPlacesFunc: 'page' parameter is required")
		return
	}
	if res.Page, err = strconv.Atoi(r.URL.Query().Get("page")); err != nil {
		http.Error(w, "In HandlerGetPlacesFunc: "+err.Error(), http.StatusBadRequest)
		return
	}
	//fmt.Println("debug")
	limit := 10
	offset := (res.Page - 1) * limit
	res.Places, res.Total, err = base.GetPlaces(limit, offset)
	if err != nil {
		http.Error(w, "In HandlerGetPlacesFunc: "+err.Error(), http.StatusBadRequest)
		return
	}
	res.Last = int(math.Ceil(float64(res.Total) / float64(limit)))

	if res.Page > res.Last || res.Page < 1 {
		http.Error(w, "Error 400\n BadRequest \nInvalid 'page' value: 'foo'", http.StatusBadRequest)
		return
	}
	tmpl, err := template.New("index.html").Funcs(
		template.FuncMap{
			"sum": sum,
			"sub": sub,
		},
	).ParseFiles("./template/index.html")

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, res)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func sum(x, y int) int {
	return x + y
}

func sub(x, y int) int {
	return x - y
}
