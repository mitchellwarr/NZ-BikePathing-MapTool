

angular.module('bmap', [])
.controller('bmapController', function TrigController($scope, $http) {
    $scope.getRoute = function() {
        var start = $scope.loc.start;
        var end = $scope.loc.end;
        if(start.lat != null && end.lat != null && !$scope.findingRoute){
            console.log('/getRoute/'+start.lat+'/'+start.lng+'/'+end.lat+"/"+end.lng);
            $http.get('/getRoute/'+start.lat+'/'+start.lng+'/'+end.lat+"/"+end.lng)
            .then(function(response) {
                data = response.data;
                $scope.loc.start.lat = data.startLat;
                $scope.loc.start.lng = data.startLon;
                $scope.loc.end.lat = data.endLat;
                $scope.loc.end.lng = data.endLon;
                
                if($scope.bikepath.length != 0){
                    for(var i = 0; i < $scope.bikepath.length; i++){
                        $scope.bikepath[i].setMap(null);
                    }
                }
                $scope.bikepath = [];
                var paths = data.paths;
                for (var i = 0; i < paths.length; i++) {
                    $scope.bikepath.push(new google.maps.Polyline({
                        path: paths[i],
                        geodesic: true,
                        strokeColor: '#ff2121',
                        strokeOpacity: 1.0,
                        strokeWeight: 4
                    }));
                    $scope.bikepath[i].setMap($scope.map);
                }

                console.log(paths);
                if($scope.heatmap != null) {
                    $scope.heatmap.setMap(null);
                }
                var heatMapData = [];
                for(var d = 0; d < data.nodes.length; d++) {
                    heatMapData.push({
                        location: new google.maps.LatLng(data.nodes[d].lat, data.nodes[d].lng),
                        weight: data.nodes[d].elevation*data.nodes[d].elevation*data.nodes[d].elevation,
                        maxIntensity: 80*80*80
                    });
                }
                $scope.heatmap = new google.maps.visualization.HeatmapLayer({
                    data: heatMapData,
                    radius: 10
                });
                $scope.setHeatMap();
                $scope.redraw();
            });
        }
    };

    $scope.setHeatMap = function() {
        if($scope.heatmap != null) {
            if($scope.showHeatMap) {
                $scope.heatmap.setMap($scope.map);
            } else {
                $scope.heatmap.setMap(null);
            }
        }
    }

    $scope.initMap = function() {
        $scope.map = new google.maps.Map(document.getElementById('googleMap'), {
            center: {lat: -39.650394, lng: 176.781005},
            zoom: 10
        });

        google.maps.event.addListener($scope.map, "click", function (event) {
            $scope.setPoint( event.latLng.lat(), event.latLng.lng() );
            $scope.redraw();
            console.log("Markers: ");
            console.log($scope.markers);
            $scope.$apply();
        });

        google.maps.event.addListener($scope.map, "mousemove", function (event) {
            $scope.setMousePoint( event.latLng.lat(), event.latLng.lng() );
            $scope.$apply();
        });
    };

    $scope.setMousePoint = function(lat, lng) {
        $scope.loc.mouse = {lat: lat, lng: lng};
    };

    $scope.setPoint = function(lat, lng) {
        switch($scope.current) {
            case 1:
                $scope.loc.start = {lat: lat, lng: lng};
                break;
            case 2:
                $scope.loc.end = {lat: lat, lng: lng};
                break;
        }
        $scope.current = 0;
    };

    $scope.redraw = function() {
        function setMapOnAll(map) {
            for (var i = 0; i < $scope.markers.length; i++) {
                $scope.markers[i].setMap(map);
            }
        }
        setMapOnAll(null);
        $scope.markers = [];

        if($scope.loc.start.lat != null) {
            var marker = new google.maps.Marker({
                position: $scope.loc.start,
                map: $scope.map
            });
            $scope.markers.push(marker);
        }
        if($scope.loc.end.lat != null) {
            var marker = new google.maps.Marker({
                position: $scope.loc.end,
                map: $scope.map
            });
            $scope.markers.push(marker);
        }
        setMapOnAll($scope.map);
    };

    $scope.formatPoint = function(lat, lng) {
        var lenstr = "0000000000"
        return (lat+lenstr).slice(0, lenstr.length)+", "+(lng+lenstr).slice(0, lenstr.length);
    }

    $scope.selectStart = function() {
        $scope.current = 1;
    };

    $scope.selectEnd = function() {
        $scope.current = 2;
    };

    $scope.getSettings = function() {
        $http.get('/getSettings')
            .then(function(response) {
                data = response.data;
                $scope.use_wind = data.wind;
                $scope.use_ele = data.elevation;
                $scope.wind_deg = data.deg;
                console.log(data);
            });
    };

    $scope.setSettings = function() {
        var p1, p2;
        if($scope.use_wind) {
            p1 = "true";
        } else {
            p1 = "false";
        }
        if($scope.use_ele) {
            p2 = "true";
        } else {
            p2 = "false";
        }
        $http.get('/setSettings/'+p1+'/'+p2+'/'+$scope.wind_deg).then(function(){
            $scope.getSettings();
        });
    };

    $scope.preset = function(type, withSettings) {
        if(type == 1) {
            $scope.loc.start.lat = -39.481769;
            $scope.loc.start.lng = 176.897694;
            $scope.loc.end.lat = -39.492425;
            $scope.loc.end.lng = 176.909476;
            $scope.wind_deg = 360;
            $scope.redraw();
            $scope.use_wind = false;
            $scope.use_ele = withSettings;
            $scope.setSettings();
        }
        if(type == 2) {
            $scope.loc.start.lat = -39.481769;
            $scope.loc.start.lng = 176.897694;
            $scope.loc.end.lat = -39.526779;
            $scope.loc.end.lng = 176.860347;
            $scope.wind_deg = 360;
            $scope.redraw();
            $scope.use_wind = withSettings;
            $scope.use_ele = false;
            $scope.setSettings();
        }
        if(type == 3) {
            $scope.loc.start.lat = -39.481769;
            $scope.loc.start.lng = 176.897694;
            $scope.loc.end.lat = -39.622149;
            $scope.loc.end.lng = 176.782565;
            $scope.wind_deg = 270;
            $scope.redraw();
            $scope.use_wind = withSettings;
            $scope.use_ele = false;
            $scope.setSettings();
        }
        if(type == 4) {
            $scope.loc.start.lat = -39.481651;
            $scope.loc.start.lng = 176.911135;
            $scope.loc.end.lat = -39.525868;
            $scope.loc.end.lng = 176.853611;
            $scope.wind_deg = 0;
            $scope.redraw();
            $scope.use_wind = withSettings;
            $scope.use_ele = false;
            $scope.setSettings();
        }
    }

    $scope.map;
    $scope.loc = {
        start: {
            lat: null,
            lng: null
        },
        end: {
            lat: null,
            lng: null
        },
        mouse: {
            lat: null,
            lng: null
        }
    };
    $scope.showHeatMap = true;
    $scope.$watch('showHeatMap', function(){
        $scope.setHeatMap();
    })
    $scope.use_wind = true;
    $scope.use_ele = true;
    $scope.wind_deg = 360;
    $scope.markers = [];
    $scope.bikepath = [];
    $scope.current = -1;
    $scope.initMap();
    $scope.setSettings();
});