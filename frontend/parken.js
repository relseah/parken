function initializeMap() {
    let map = L.map('map').setView([49.40, 8.67], 13);
    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href=\"https://www.openstreetmap.org/copyright\">OpenStreetMap</a> contributors'
    }).addTo(map);    
    return map
}

function processPosition(map, position) {
    L.marker([position.coords.latitude, position.coords.longitude]).addTo(map);
}

let map = initializeMap();
if ('geolocation' in navigator) {
    navigator.geolocation.watchPosition(position => processPosition(map, position));
}
