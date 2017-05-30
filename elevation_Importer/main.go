package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"strconv"

	_ "github.com/lib/pq"
	"googlemaps.github.io/maps"
)

const (
	DB_USER        = "postgres"
	DB_PASSWORD    = "postgres"
	DB_NAME        = "nz_network"
	DB_GET_NODES   = "SELECT id, st_astext(the_geom), the_geom <-> ST_GeometryFromText('POINT(176.90 -39.48)') as distance FROM nz_roads_fix_vertices_pgr WHERE elevation IS NULL ORDER BY distance LIMIT 500"
	DB_UPDATE_NODE = "UPDATE nz_roads_fix_vertices_pgr SET elevation=$1 WHERE id=$2"
	ELEVATION_KEY  = "AIzaSyBEbuMob4gb5Nonp-wkOsPOaDEnLcNr8HU"
	URL            = "https://maps.googleapis.com/maps/api/elevation/json?locations=enc:%s&key=%s"
)

type dbWorkingInfo struct {
	points []maps.LatLng
	ids    []int64
	height []float64
	mean   float64
	url    string
}

type elevationJSON struct {
	Result []struct {
		Elevation  float64 `json:"elevation"`
		Resolution float64 `json:"resolution"`
	} `json:"results"`
}

var db *sql.DB = OpenDatabase()

func main() {
	for i := 0; i < 1; i++ {
		points := getNextNodes()
		points = getHeightOfPoints(points)
		fmt.Println("\n--\nMean", points.mean)
		fmt.Println(addElevations(points))
	}
}

func addElevations(points dbWorkingInfo) string {
	Tx, err := db.Begin()
	CheckErr(err)
	defer Tx.Rollback()

	stmt, err := Tx.Prepare(DB_UPDATE_NODE)
	CheckErr(err)

	var affect int64
	for i := range points.height {
		res, err := stmt.Exec(points.height[i], points.ids[i])
		CheckErr(err)

		rows, err := res.RowsAffected()
		CheckErr(err)

		affect += rows
	}

	err = Tx.Commit()
	CheckErr(err)

	return fmt.Sprintf("%d rows changed", affect)
}

func getNextNodes() dbWorkingInfo {
	rows, err := db.Query(DB_GET_NODES)
	CheckErr(err)
	defer rows.Close()

	points := dbWorkingInfo{
		points: make([]maps.LatLng, 0),
		ids:    make([]int64, 0),
	}

	for rows.Next() {
		var id int64
		var dist float64
		var pointLonLat string
		err = rows.Scan(&id, &pointLonLat, &dist)
		CheckErr(err)
		points.points = append(points.points,
			getCoordsFromGeom(strings.Trim(pointLonLat, "POINT()")))
		points.ids = append(points.ids, id)
	}

	return points
}

func getCoordsFromGeom(coords string) maps.LatLng {
	lonlat := strings.Split(coords, " ")
	lon := parseFloat(lonlat[0])
	lat := parseFloat(lonlat[1])
	return maps.LatLng{Lat: lat, Lng: lon}
}

func getHeightOfPoints(points dbWorkingInfo) dbWorkingInfo {
	encodedPoly := maps.Encode(points.points)
	url := fmt.Sprintf(URL, encodedPoly, ELEVATION_KEY)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	CheckErr(err)

	// For control over HTTP client headers,
	// redirect policy, and other settings,
	// create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	resp, err := client.Do(req)
	CheckErr(err)

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	// Fill the record with the data from the JSON
	var record elevationJSON
	if err := json.Unmarshal([]byte(string(body)), &record); err != nil {
		log.Fatal(err)
	}

	points.height = make([]float64, 0)
	for _, r := range record.Result {
		points.height = append(points.height, r.Elevation)
		points.mean += r.Elevation
	}
	points.mean = float64(int(points.mean) / len(record.Result))
	points.url = url

	return points
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
