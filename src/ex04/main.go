package main

import (
	"Day03/ex03/db"
	"Day03/ex03/types"
	"Day03/ex04/middleware/jwtauth"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
)

type Store interface {
	GetPlaces(limit int, offset int) ([]types.Place, int, error)
	GetClosest(lat, lon float64) ([]types.Place, error)
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
	fmt.Println("Listening on port 8888")
	http.HandleFunc("/", HandlerGetPlaces)
	http.HandleFunc("/api/places", HandlerApiGetPlaces)
	http.HandleFunc("/api/recommend", jwtauth.JwtMiddleware(HandlerApiClosestPlaces))
	http.HandleFunc("/api/get_token", HandlerGetToken)
	err = http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func HandlerGetToken(w http.ResponseWriter, r *http.Request) {
	const op = "HandlerGetToken"
	tokenString, err := jwtauth.GenerateJwt()
	if err != nil {
		http.Error(w, op, http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(map[string]string{"token": tokenString})
	if err != nil {
		http.Error(w, op, http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	func() {
		_, err = w.Write(response)
		if err != nil {
			log.Println(op, err)
		}
	}()
	w.WriteHeader(http.StatusOK)
}

func HandlerApiClosestPlaces(w http.ResponseWriter, r *http.Request) {
	const op = "HandlerClosestPlaces"
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		http.Error(w, op+": "+err.Error(), http.StatusBadRequest)
	}
	lon, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if err != nil {
		http.Error(w, op+": "+err.Error(), http.StatusBadRequest)
	}
	place, err := base.GetClosest(lat, lon)
	if err != nil {
		http.Error(w, op+": "+err.Error(), http.StatusBadRequest)
	}
	res := types.NewResponse(place)
	response, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		http.Error(w, op+":"+err.Error(), http.StatusInternalServerError)
	}
	func() {
		_, err = w.Write(response)
		if err != nil {
			log.Println(op + ":" + err.Error())
		}
	}()
	w.WriteHeader(http.StatusOK)

}

func HandlerApiGetPlaces(w http.ResponseWriter, r *http.Request) {
	const op = "HandlerApiPlaces"
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var res Paginator
	var err error
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		http.Error(w, fmt.Sprintf("%s query string param is empty", op), http.StatusBadRequest)
		return
	}
	if res.Page, err = strconv.Atoi(r.URL.Query().Get("page")); err != nil {
		http.Error(w, "In HandlerGetPlacesFunc: "+err.Error(), http.StatusBadRequest)
		return
	}
	limit := 10
	offset := (res.Page - 1) * limit
	res.Places, res.Total, err = base.GetPlaces(limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s query string param is empty", op), http.StatusBadRequest)
	}
	res.Last = int(math.Ceil(float64(res.Total) / float64(limit)))
	if res.Page > res.Last || res.Page < 1 {
		http.Error(w, "Error 400\n BadRequest \nInvalid 'page' value: 'foo'", http.StatusBadRequest)
		return
	}

	response, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("%s json marshal error: %v", op, err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s write response error: %v", op, err), http.StatusInternalServerError)
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
