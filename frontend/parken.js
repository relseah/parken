function initializeMap() {
	let map = L.map("map").setView([49.41032, 8.69707], 13);
	L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
		attribution:
			'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	}).addTo(map);
	return map;
}

function initializeLocateControl(map) {
	L.control.locate({ position: "bottomright" }).addTo(map);
}

let map = initializeMap();
if ("geolocation" in navigator)
	if ("permissions" in navigator)
		navigator.permissions.query({ name: "geolocation" }).then((status) => {
			if (status.state != "denied") {
				initializeLocateControl(map);
			}
			status.onchange = function () {
				if (this.state == "prompt") initializeLocateControl(map);
			};
		});
	else initializeLocateControl();
