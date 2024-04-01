package main

import (
	"compress/gzip"
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed templates/*.html
var t embed.FS
var templates, err = template.ParseFS(t, "templates/*.html")

var db, _ = sql.Open("sqlite3", "db.db")

func main() {
	if err != nil {
		fmt.Print(err)
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", home)
	mux.HandleFunc("GET /filter", filter)

	http.ListenAndServe(":8080", mux)
}

func gzipTemplate(w http.ResponseWriter, file string, data interface{}) {
	w.Header().Set("content-encoding", "gzip")
	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close()
	templates.ExecuteTemplate(gzipWriter, file, data)
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	results := selectAll()
	gzipTemplate(w, "index.html", results)
}

func filter(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	search2 := r.URL.Query().Get("search2")
	column := r.URL.Query().Get("column")
	order := r.URL.Query().Get("order")
	empty := r.URL.Query().Get("empty")
	minPriceStr := r.URL.Query().Get("minprice")
	maxPriceStr := r.URL.Query().Get("maxprice")
	if minPriceStr == "" {
		minPriceStr = "0"
	}
	if maxPriceStr == "" {
		maxPriceStr = "9999"
	}
	minPrice, _ := strconv.ParseUint(minPriceStr, 10, 64)
	maxPrice, _ := strconv.ParseUint(maxPriceStr, 10, 64)

	results := filterAll("%"+search+"%", "%"+search2+"%", column, order, empty, uint16(minPrice), uint16(maxPrice))
	gzipTemplate(w, "data", results)
}

type Laptop struct {
	ID            uint8
	Title         string
	Link          string
	Price         uint16
	Description   string
	Searchtext    string
	Family        string
	Model         string
	Cores         uint8
	Threads       uint8
	Clockspeed    float32
	L3cache       uint8
	Gpucores      uint8
	Gpuclockspeed float32
	Tdp           uint8
}

func selectAll() []*Laptop {
	results := make([]*Laptop, 0, 500)
	stmt, _ := db.Prepare(`
	select laptops.title, laptops.link, laptops.price, laptops.searchtext, cpus.family, cpus.model, cpus.cores, cpus.threads, cpus.clockspeed, cpus.gpucores, cpus.gpuclockspeed
	from laptops 
	join cpus on laptops.model = cpus.model
	order by price
	limit 500
	`)
	rows, _ := stmt.Query()
	for rows.Next() {
		laptop := new(Laptop)
		rows.Scan(&laptop.Title, &laptop.Link, &laptop.Price, &laptop.Searchtext, &laptop.Family, &laptop.Model, &laptop.Cores, &laptop.Threads, &laptop.Clockspeed, &laptop.Gpucores, &laptop.Gpuclockspeed)
		results = append(results, laptop)
	}
	return results
}

func filterAll(search, search2, column, order, empty string, minprice, maxprice uint16) []*Laptop {
	query := fmt.Sprintf(`select laptops.id, laptops.title, laptops.link, laptops.price, laptops.searchtext, cpus.family, cpus.model, cpus.cores, cpus.threads, cpus.clockspeed, cpus.gpucores, cpus.gpuclockspeed
	from laptops 
	join cpus on laptops.model = cpus.model
	where laptops.searchtext like ?
	and laptops.searchtext like ?
	and laptops.price >= ?
	and laptops.price <= ?
	and cpus.model != ?
	order by %s %s
	limit 500
	`, column, order)
	results := make([]*Laptop, 0, 500)
	stmt, _ := db.Prepare(query)
	rows, _ := stmt.Query(search, search2, minprice, maxprice, empty)
	for rows.Next() {
		laptop := new(Laptop)
		rows.Scan(&laptop.ID, &laptop.Title, &laptop.Link, &laptop.Price, &laptop.Searchtext, &laptop.Family, &laptop.Model, &laptop.Cores, &laptop.Threads, &laptop.Clockspeed, &laptop.Gpucores, &laptop.Gpuclockspeed)
		results = append(results, laptop)
	}
	return results
}
