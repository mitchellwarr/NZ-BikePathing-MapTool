package main

import (
	"container/heap"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var weather []weatherInfo

func GetRoutePolyLine(points routePoints) (*node, *node, [][]node, []*node) {
	weather = getWeatherInfo()

	startNodeID := getClosestNode(points.start)
	endNodeID := getClosestNode(points.end)
	fmt.Println(startNodeID, " - ", endNodeID)
	currentNode, closedNodes, err := routePath(startNodeID, endNodeID)
	//current := currentNode
	paths := getPathsFromParents(currentNode)
	for _, el := range paths {
		fmt.Println("---")
		fmt.Println(el)
	}
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("\nDistance: ", currentNode.gcost)
	startNode := getNodeFromID(startNodeID)
	endNode := getNodeFromID(endNodeID)
	//nodes := getNodesFromParents(current)
	return startNode, endNode, paths, closedNodes
	//googleLine := makePolyFromPath(paths)
}

func getNodesFromParents(n *node) []node {
	nodes := make([]node, 0)
	current := n
	for current.parent != nil {
		newNode := node{
			Lat:       current.Lat,
			Lon:       current.Lon,
			Elevation: current.Elevation,
		}
		nodes = append(nodes, newNode)
		current = current.parent
	}
	return nodes
}

func getClosestNode(node node) int64 {

	rows, err := db.Query(
		fmt.Sprintf(
			"SELECT id FROM public.%s ORDER BY the_geom <-> ST_GeometryFromText('POINT(%f %f)') LIMIT 1;",
			DB_NODE,
			node.Lon, node.Lat))
	CheckErr(err)
	defer rows.Close()
	var id int64
	if rows.Next() {
		err = rows.Scan(&id)
		CheckErr(err)
	} else {
		CheckErr(err)
	}

	return id
}

func getNeighbours(n *node) []*node {
	var row1, row2 *sql.Rows
	neighbours := make([]*node, 0)
	ids := make([]int64, 0)

	row1, err := db.Query(
		fmt.Sprintf(
			"SELECT target as neighbour FROM public.%s JOIN public.%s ON public.%s.id = public.%s.source WHERE public.%s.id = %d;",
			DB_NODE, DB_PATH, DB_NODE, DB_PATH, DB_NODE,
			n.id))
	CheckErr(err)
	defer row1.Close()
	for row1.Next() {
		var id int64
		err = row1.Scan(&id)
		CheckErr(err)
		ids = append(ids, id)
	}
	row2, err = db.Query(
		fmt.Sprintf(
			"SELECT source as neighbour FROM public.%s JOIN public.%s ON public.%s.id = public.%s.target WHERE public.%s.id = %d;",
			DB_NODE, DB_PATH, DB_NODE, DB_PATH, DB_NODE,
			n.id))
	CheckErr(err)
	defer row2.Close()
	for row2.Next() {
		var id int64
		err = row2.Scan(&id)
		CheckErr(err)
		ids = append(ids, id)
	}
	for _, id := range ids {
		neighbours = append(neighbours, getNodeFromID(id))
	}

	return neighbours
}

func getNodeFromID(id int64) *node {

	rows, err := db.Query(
		fmt.Sprintf(
			"SELECT st_astext(the_geom) as lonlat, elevation FROM public.%s WHERE id = %d;",
			DB_NODE,
			id))
	CheckErr(err)
	defer rows.Close()
	newNode := node{}
	if rows.Next() {
		var pointLonLat string
		var elevation float64
		err = rows.Scan(&pointLonLat, &elevation)
		CheckErr(err)
		pointLonLat = strings.Trim(pointLonLat, "POINT()")
		addCoordsToNode(&newNode, pointLonLat)
		newNode.id = id
		newNode.Elevation = elevation
	}

	return &newNode
}

func getPathBetweenNodes(n1, n2 *node) string {

	rows, err := db.Query(
		fmt.Sprintf(
			"SELECT st_astext(geom) FROM public.%s WHERE (source = %d OR target = %d) AND (source = %d OR target = %d);",
			DB_PATH,
			n1.id, n1.id, n2.id, n2.id))
	CheckErr(err)
	defer rows.Close()
	var coords string
	if rows.Next() {
		err = rows.Scan(&coords)
		CheckErr(err)
	} else {
		panic("Cannot find path")
	}

	return coords
}

func getPathsFromParents(child *node) [][]node {
	paths := make([][]node, 0)
	for !(child.parent == nil) {
		coords := getPathBetweenNodes(child, child.parent)
		coords = strings.Trim(coords, "MULTILINESTRING(())")
		path := createPathFromCoords(coords)
		paths = append(paths, path)
		child = child.parent
	}
	return paths

}

func routePath(startNodeID, endNodeID int64) (*node, []*node, error) {
	startNode := getNodeFromID(startNodeID)
	endNode := getNodeFromID(endNodeID)
	startNode.calcCost(0, calcHCost(startNode, endNode))
	var current *node

	openNodes := HeapNode{}
	heap.Init(&openNodes)
	heap.Push(&openNodes, startNode)

	closedNodes := make([]*node, 0)

	for {

		if openNodes.Len() == 0 {
			return current, closedNodes, errors.New("Could not find exit")
		}

		current = heap.Pop(&openNodes).(*node)
		closedNodes = append(closedNodes, current)

		fmt.Printf("Open %d Closed %d dist %f\n", len(openNodes), len(closedNodes), calcHCost(current, endNode))

		for _, neighbour := range getNeighbours(current) {
			if !ContainsNode(closedNodes, neighbour) {
				newPath := &node{Lat: neighbour.Lat, Lon: neighbour.Lon, id: neighbour.id, Elevation: neighbour.Elevation}
				newPath.parent = current
				generateCosts(newPath, endNode)
				if (!ContainsNode(openNodes, neighbour)) || newPath.isShorterThan(neighbour) {
					neighbour.parent = current
					generateCosts(neighbour, endNode)
					if !ContainsNode(openNodes, neighbour) {
						heap.Push(&openNodes, neighbour)
					}
				}
			}
		}
		if current.id == endNode.id {
			break
		}
	}
	return current, closedNodes, nil
}

func getLengthOfNodeArray(nodeList []node) float64 {
	length := 0.0
	for i := 1; i < len(nodeList); i++ {
		lat := nodeList[i-1].Lat - nodeList[i].Lat
		lon := nodeList[i-1].Lon - nodeList[i].Lon
		length += math.Sqrt((lat * lat) + (lon * lon))
	}
	return length
}

func createPathFromCoords(coords string) []node {
	coordsMultiList := strings.Split(coords, "),(")
	nodes := make([]node, 0)
	for _, coordsSeg := range coordsMultiList {
		coordsList := strings.Split(coordsSeg, ",")
		for i, el := range coordsList {
			nodes = append(nodes, node{})
			addCoordsToNode(&nodes[i], el)
		}
	}
	return nodes
}

func addCoordsToNode(n *node, coords string) {
	lonlat := strings.Split(coords, " ")
	n.Lon = parseFloat(lonlat[0])
	n.Lat = parseFloat(lonlat[1])
}

func generateCosts(n, endNode *node) {
	g := n.parent.gcost
	pathCost := getPathCost(n.parent, n)
	n.calcCost(g+pathCost, calcHCost(n, endNode))
}

func calcHCost(node1, node2 *node) float64 {
	x := node1.Lat - node2.Lat
	y := node1.Lon - node2.Lon
	var h = math.Sqrt((x * x) + (y * y))
	return h * HCOSTER
}

func getPathCost(n1, n2 *node) float64 {
	coords := getPathBetweenNodes(n1, n2)
	coords = strings.Trim(coords, "MULTILINESTRING(())")
	path := createPathFromCoords(coords)
	return getCostOfNodeArray(n1, n2, path)
}

func getCostOfNodeArray(n, n2 *node, nodeList []node) float64 {
	length := 0.0
	time := getTimeForDist(n.gcost)
	var elevation float64
	if nodeList[0].Lat == n.Lat && nodeList[0].Lon == n.Lon {
		elevation = n2.Elevation - n.Elevation
		for i := 1; i < len(nodeList); i++ {
			x := (nodeList[i].Lat - nodeList[i-1].Lat) * 110.574
			y := (nodeList[i].Lon - nodeList[i-1].Lon) * 111.320 * math.Cos(nodeList[i-1].Lat*(math.Pi/180))
			dist := math.Sqrt((x * x) + (y * y))
			angle := math.Atan2(y, x)
			weatherSlice := getWeatherFromTime(time, weather)
			length += applyTerrainMod(dist, angle, weatherSlice)
			time = getTimeForDist(n.gcost + length)
		}
	} else {
		elevation = n.Elevation - n2.Elevation
		for i := len(nodeList) - 1; i > 0; i-- {
			x := (nodeList[i-1].Lat - nodeList[i].Lat) * 110.574
			y := (nodeList[i-1].Lon - nodeList[i].Lon) * 111.320 * math.Cos(nodeList[i-1].Lat*(math.Pi/180))
			dist := math.Sqrt((x * x) + (y * y))
			angle := math.Atan2(y, x)
			weatherSlice := getWeatherFromTime(time, weather)
			length += applyTerrainMod(dist, angle, weatherSlice)
			time = getTimeForDist(n.gcost + length)
		}
	}
	length = applyElevationMod(length, elevation)
	return length
}

func getTimeForDist(d float64) time.Time {
	tsec := PREV_TIME * math.Pow((d/PREV_DIST), (1+PERCENT_SLOW_DOWN))
	return time.Now().Add(time.Second * time.Duration(tsec))
}

func applyTerrainMod(dist, angle float64, weather weatherInfo) float64 {
	newDist := dist
	if USE_WIND {
		windDeg := (weather.Wind.Deg * (-1)) + 90
		angle = angle + (math.Pi / 2)
		angle = angle * (180 / math.Pi)
		x := weather.Wind.Speed*math.Cos(windDeg) + VELOCITY*math.Cos(angle)
		y := weather.Wind.Speed*math.Sin(windDeg) + VELOCITY*math.Sin(angle)
		length := math.Sqrt((x * x) + (y * y))
		newDist = dist * (length / VELOCITY)
	}
	return newDist
}

func applyElevationMod(dist, elevation float64) float64 {
	grade := 1.0
	if USE_ELEVATION {
		grade = ((elevation / (dist * 1000)) * 2) + 1
	}
	return dist * grade
}
