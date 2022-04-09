function initializeMap() {
	let map = L.map("map").setView([49.41032, 8.69707], 13);
	L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
		attribution:
			'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	}).addTo(map);
	return map;
}

function markParkings(parkings) {
	let factor = 2;
	let icon = L.elementIcon(document.createElement("span"), {
		className: `fa-solid fa-square-parking fa-${factor}x parking`,
		size: [(7 / 8) * factor, factor],
		sizeUnit: "em",
	});
	for (const parking of parkings) {
		L.marker([parking.coordinates.latitude, parking.coordinates.longitude], {
			icon: icon,
		}).addTo(map);
	}
}

function processPosition(position) {}

var map = initializeMap();
var locateControl = L.control.locate(
	function () {
		this.addTo(map);
		this.requestPosition();
	},
	{ position: "bottomright", onPosition: processPosition }
);
var parkings;
fetch("api")
	.then((response) => response.json())
	.then((data) => {
		parkings = data;
		markParkings(parkings);
	});
