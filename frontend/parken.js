function initializeMap() {
    let map = L.map('map').setView([49.41032, 8.69707], 13);
    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href=\"https://www.openstreetmap.org/copyright\">OpenStreetMap</a> contributors'
    }).addTo(map);
    return map
}

let map = initializeMap();
if ('geolocation' in navigator) {
    let locateControl = L.control.locate({position: 'bottomright'}).addTo(map);
    locateControl.requestPosition();
}
