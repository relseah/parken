function initializeMap() {
	let map = L.map("map").setView([49.41032, 8.69707], 13);
	L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
		attribution:
			'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	}).addTo(map);
	return map;
}

let highlightedParkingElement;
function markParkings() {
	let factor = 2;
	let icon = L.elementIcon(document.createElement("span"), {
		className: `fa-solid fa-square-parking fa-${factor}x marker-parking`,
		// The icon's width is 7/8 of its height.
		iconAnchor: [7 / 16, 0.5],
		anchorUnit: "em",
	});
	for (let parking of parkings) {
		let marker = L.marker(parking.coordinates, {
			icon: icon,
		}).addTo(map);
		marker.on("click", () => {
			if (highlightedParkingElement) {
				highlightedParkingElement.classList.remove("parking-highlighted");
			}
			if (highlightedParkingElement === parking.element) {
				highlightedParkingElement = null;
			} else {
				parking.element.classList.add("parking-highlighted");
				parking.element.scrollIntoView();
				highlightedParkingElement = parking.element;
			}
		});
	}
}

function redirectToNavigation(address) {
	url =
		"https://www.google.com/maps/dir/?api=1&travelmode=driving&destination=" +
		formatAddress(address);
	window.open(url, "_blank").focus();
}

function formatAddress(address) {
	return `${address.street}${
		address.houseNumber ? " " + address.houseNumber : ""
	}, ${address.postalCode} ${address.town}`;
}

function toElement(parking) {
	let li = document.createElement("li");
	let navigationButton = document.createElement("button");
	navigationButton.className = "navigation";
	let navigationIcon = document.createElement("span");
	navigationIcon.className = "fa-solid fa-route fa-2x";
	navigationButton.append(navigationIcon);
	navigationButton.addEventListener("click", () => {
		redirectToNavigation(parking.address);
	});
	li.append(navigationButton);
	let nameStrong = document.createElement("strong");
	nameStrong.textContent = `P${parking.id} ${parking.name}`;
	nameStrong.className = "parking-name";
	nameStrong.onclick = () => {
		map.panTo(parking.coordinates);
	};
	li.append(nameStrong);
	let addressDiv = document.createElement("div");
	addressDiv.textContent = formatAddress(parking.address);
	addressDiv.className = "parking-address";
	li.append(addressDiv);
	let capacityDiv = document.createElement("div");
	capacityDiv.textContent = "Kapazität: " + parking.capacity;
	li.append(capacityDiv);
	let distanceDiv = document.createElement("div");
	distanceDiv.textContent = "Entfernung: -";
	distanceDiv.className = "parking-distance";
	li.append(distanceDiv);
	parking.distanceDiv = distanceDiv;
	li.append(document.createElement("hr"));
	return li;
}

let parkingsElement;
function listParkings() {
	parkingsElement = document.getElementById("parkings");
	for (parking of parkings) {
		parkingsElement.append(parking.element);
	}
	displayedParkings = [...parkings];
}

let displayedParkings;
function sortParkings() {
	parkings.sort((a, b) => {
		if (a.distance === b.distance) return 0;
		return a.distance < b.distance ? -1 : 1;
	});
	let identical = true;
	for (let i = 0; identical && i < parkings.length; i++) {
		if (parkings[i] !== displayedParkings[i]) {
			identical = false;
		}
	}
	return identical;
}

function updateParkingsElement() {
	refreshButton.disabled = true;
	for (parking of parkings) {
		parkingsElement.append(parking.element);
	}
	displayedParkings = [...parkings];
}

function processPosition(position) {
	position = L.latLng(position.coords.latitude, position.coords.longitude);
	for (let parking of parkings) {
		let distance = Math.round(position.distanceTo(parking.coordinates));
		let unit;
		if (distance > 999) {
			unit = "km";
			distance *= 0.001;
		} else unit = "m";
		parking.distanceDiv.textContent = `Entfernung: ${distance.toLocaleString(
			"de-DE"
		)} ${unit}`;
		parking.distance = distance;
	}
	let identical = sortParkings();
	if (identical) refreshButton.disabled = true;
	else {
		if (locateControl.getIsFirstPosition()) updateParkingsElement();
		else refreshButton.disabled = false;
	}
}

let map = initializeMap();
let latestPosition;
let locateControl = L.control.locate(
	function () {
		this.addTo(map);
		this.requestPosition();
	},
	{
		position: "bottomright",
		markerOptions: {
			radius: 5,
			fillOpacity: 1,
			color: "#005a8c",
			onPosition: (position) => (latestPosition = position),
		},
	}
);

let refreshButton = document.getElementById("refresh");
refreshButton.onclick = updateParkingsElement;

let parkings;
fetch("/api")
	.then((response) => response.json())
	.then((data) => {
		for (let parking of data) {
			parking.element = toElement(parking);
			parking.coordinates = L.latLng(
				parking.coordinates.latitude,
				parking.coordinates.longitude
			);
		}
		parkings = data;
		markParkings();
		listParkings();
		if (latestPosition) processPosition(latestPosition);
		locateControl.options.onPosition = processPosition;
		latestPosition = null;
	});
