import 'leaflet/dist/leaflet.css'
import './parken.css'

import L from 'leaflet'

var map = L.map('map').setView([49.40, 8.67], 13);
L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href=\"https://www.openstreetmap.org/copyright\">OpenStreetMap</a> contributors'
}).addTo(map);
