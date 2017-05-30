

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
                        strokeColor: '#FF0000',
                        strokeOpacity: 1.0,
                        strokeWeight: 2
                    }));
                    $scope.bikepath[i].setMap($scope.map);
                }

                console.log(paths);
                $scope.redraw();
            });
        }
    };

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
    $scope.markers = [];
    $scope.bikepath = [];
    $scope.current = -1;
    $scope.initMap();
});