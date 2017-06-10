package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	DB_USER            = "postgres"
	DB_PASSWORD        = "postgres"
	DB_NAME            = "nz_network"
	DB_NODE            = "nz_roads_fix_vertices_pgr"
	DB_PATH            = "nz_roads_fix"
	HCOSTER            = 1.0
	PREV_TIME          = 714 + ((0 + (0 * 60)) * 60) //sec + (min + (hour * 60)) * 60);
	PREV_DIST          = 5
	VEL_M              = (PREV_DIST * 1000)
	VEL_T              = PREV_TIME
	VELOCITY           = VEL_M / VEL_T
	PERCENT_SLOW_DOWN  = 0.06
	USE_TEST_WIND_DATA = true
)

var USE_WIND bool = false
var TEST_WIND_DEG float64 = 270
var USE_ELEVATION bool = true

var db *sql.DB = OpenDatabase()

type routePoints struct {
	start node
	end   node
}

type Response struct {
	StartLat float64  `json:"startLat"`
	StartLon float64  `json:"startLon"`
	EndLat   float64  `json:"endLat"`
	EndLon   float64  `json:"endLon"`
	Paths    [][]node `json:"paths"`
	Nodes    []*node  `json:"nodes"`
}

// error response contains everything we need to use http.Error
type handlerError struct {
	Error   error
	Message string
	Code    int
}

// a custom type that we can use for handling errors and formatting responses
type handler func(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError)

// attach the standard ServeHTTP method to our handler so the http library can call it
func (fn handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// here we could do some prep work before calling the handler if we wanted to

	// call the actual handler
	response, err := fn(w, r)

	// check for errors
	if err != nil {
		log.Printf("ERROR: %v\n", err.Error)
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Message), err.Code)
		return
	}
	if response == nil {
		log.Printf("ERROR: response from method is nil\n")
		http.Error(w, "Internal server error. Check the logs.", http.StatusInternalServerError)
		return
	}

	// turn the response into JSON
	bytes, e := json.Marshal(response)
	if e != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}

	// send the response and log
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	log.Printf("%s %s %s %d", r.RemoteAddr, r.Method, r.URL, 200)
}

func getRoute(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	// mux.Vars grabs variables from the path
	startlat := parseFloat(mux.Vars(r)["startlat"])
	startlon := parseFloat(mux.Vars(r)["startlon"])
	endlat := parseFloat(mux.Vars(r)["endlat"])
	endlon := parseFloat(mux.Vars(r)["endlon"])
	points := routePoints{
		start: node{
			Lat: startlat,
			Lon: startlon,
		},
		end: node{
			Lat: endlat,
			Lon: endlon,
		},
	}
	fmt.Printf("Route to: %f, %f -> %f, %f\n", points.start.Lat, points.start.Lon, points.end.Lat, points.end.Lon)

	start, end, paths, nodes := GetRoutePolyLine(points)
	response := Response{
		StartLat: start.Lat,
		StartLon: start.Lon,
		EndLat:   end.Lat,
		EndLon:   end.Lon,
		Paths:    paths,
		Nodes:    nodes,
	}
	fmt.Println(response)
	return response, nil
}

func getSettings(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	type res struct {
		Wind      bool    `json:"wind"`
		Elevation bool    `json:"elevation"`
		Deg       float64 `json:"deg"`
	}
	response := res{
		Wind:      USE_WIND,
		Elevation: USE_ELEVATION,
		Deg:       TEST_WIND_DEG,
	}
	fmt.Println(response)
	return response, nil
}

func setSettings(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	// mux.Vars grabs variables from the path
	USE_WIND = mux.Vars(r)["wind"] == "true"
	TEST_WIND_DEG = parseFloat(mux.Vars(r)["deg"])
	USE_ELEVATION = mux.Vars(r)["elevation"] == "true"
	type res struct{}
	return res{}, nil
}

func main() {
	fmt.Println("\n\n\n\n---\nRunning Server\n---\n\n")
	// getNetwork()

	// command line flags
	port := flag.Int("port", 8080, "port to serve on")
	dir := flag.String("directory", "./app/web", "directory of web files")
	flag.Parse()

	// handle all requests by serving a file of the same name
	fs := http.Dir(*dir)
	fileHandler := http.FileServer(fs)

	// setup routes
	router := mux.NewRouter()
	router.Handle("/", http.RedirectHandler("/my-app/", 302))
	router.Handle("/getRoute/{startlat}/{startlon}/{endlat}/{endlon}", handler(getRoute)).Methods("GET")
	router.Handle("/getSettings", handler(getSettings)).Methods("GET")
	router.Handle("/setSettings/{wind}/{elevation}/{deg}", handler(setSettings)).Methods("GET")
	router.PathPrefix("/my-app/").Handler(http.StripPrefix("/my-app", fileHandler))
	http.Handle("/", router)

	log.Printf("Running on port %d\n", *port)

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	// this call blocks -- the progam runs here forever
	err := http.ListenAndServe(addr, nil)
	fmt.Println(err.Error())
	panic(http.ListenAndServe(":8080", http.FileServer(http.Dir("./app/web"))))
}

func parseFloat(num string) float64 {
	number, err := strconv.ParseFloat(num, 64)
	CheckErr(err)
	return number
}

func OpenDatabase() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	CheckErr(err)
	return db
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
