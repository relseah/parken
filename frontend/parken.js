function initializeMap() {
	let map = L.map("map").setView([49.41032, 8.69707], 12);
	L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png ", {
		attribution:
			'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	}).addTo(map);
	return map;
}

function highlightParkingElement(element, scroll) {
	if (highlightedParkingElement) {
		highlightedParkingElement.classList.remove("parking-highlighted");
	}
	if (highlightedParkingElement === element) {
		highlightedParkingElement = null;
	} else {
		element.classList.add("parking-highlighted");
		if (scroll) {
			element.scrollIntoView();
		}
		highlightedParkingElement = element;
	}
}

let highlightedParkingElement;
function markParkings() {
	let factor = 2;
	let icon = L.elementIcon(document.createElement("span"), {
		className: `fa-solid fa-square-parking fa-${factor}x marker`,
		// The icon's width is 7/8 of its height.
		iconAnchor: [7 / 16, 0.5],
		anchorUnit: "em",
	});
	for (let parking of parkings) {
		let marker = L.marker(parking.coordinates, {
			icon: icon,
		});
		marker.on("click", () => {
			highlightParkingElement(parking.element, true);
		});
		marker.bindPopup(() => {
			let popupDiv = document.createElement("div");
			let nameStrong = createNameStrong(parking);
			popupDiv.append(nameStrong);
			let navigationButton = createNavigationButton(parking.address, 1);
			popupDiv.append(navigationButton);
			let occupancyDiv = createOccupancyDiv(parking);
			popupDiv.append(occupancyDiv);
			popupDiv.className = "popup";
			return popupDiv;
		});
		marker.addTo(map);
		parking.marker = marker;
	}
}

function redirectToNavigation(parking) {
	url =
		"https://www.google.com/maps/dir/?api=1&travelmode=driving&destination=" +
		`P${parking.id} ${parking.name}, Heidelberg`;
	window.open(url, "_blank").focus();
}

function createNavigationButton(parking, sizeFactor) {
	let navigationButton = document.createElement("button");
	navigationButton.className = "navigation";
	let navigationIcon = document.createElement("span");
	navigationIcon.className = "fa-solid fa-route";
	if (sizeFactor !== 1) {
		navigationIcon.classList.add("fa-" + sizeFactor + "x");
	}
	navigationButton.append(navigationIcon);
	navigationButton.addEventListener("click", () => {
		redirectToNavigation(parking);
	});
	return navigationButton;
}

function createNameStrong(parking) {
	let nameStrong = document.createElement("strong");
	nameStrong.textContent = `P${parking.id} ${parking.name}`;
	nameStrong.className = "name";
	return nameStrong;
}

function createOccupancyDiv(parking) {
	let occupancyDiv = document.createElement("div");
	occupancyDiv.textContent = "Belegung: ";
	let occupancySpan = document.createElement("span");
	occupancySpan.textContent = `${parking.spots}/${parking.capacity}`;
	let color;
	if (parking.spots < 10) color = "red";
	else if (parking.spots < 20) color = "orange";
	else color = "green";
	occupancySpan.className = "occupancy-" + color;
	occupancyDiv.append(occupancySpan);
	return occupancyDiv;
}

function formatAddress(address) {
	return `${address.street}${
		address.houseNumber ? " " + address.houseNumber : ""
	}, ${address.postalCode} ${address.town}`;
}

function convertToElement(parking) {
	let li = document.createElement("li");
	let navigationButton = createNavigationButton(parking, 2);
	li.append(navigationButton);
	let nameStrong = createNameStrong(parking);
	nameStrong.onclick = () => {
		highlightParkingElement(parking.element, false);
		parking.marker.togglePopup();
	};
	li.append(nameStrong);
	let addressDiv = document.createElement("div");
	addressDiv.textContent = formatAddress(parking.address);
	addressDiv.className = "address";
	li.append(addressDiv);
	let occupancyDiv = createOccupancyDiv(parking);
	li.append(occupancyDiv);
	let distanceDiv = document.createElement("div");
	distanceDiv.textContent = "Entfernung: ";
	let distanceSpan = document.createElement("span");
	distanceSpan.textContent = "-";
	distanceDiv.append(distanceSpan);
	li.append(distanceDiv);
	parking.distanceSpan = distanceSpan;
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

function updateParkingsList() {
	sortButton.disabled = true;
	for (parking of parkings) {
		parkingsElement.append(parking.element);
	}
	displayedParkings = [...parkings];
}

function processPosition(position) {
	position = L.latLng(position.coords.latitude, position.coords.longitude);
	for (let parking of parkings) {
		let distance = Math.round(position.distanceTo(parking.coordinates));
		let displayDistance, unit;
		if (distance > 999) {
			unit = "km";
			displayDistance = distance * 0.001;
		} else unit = "m";
		parking.distanceSpan.textContent =
			distance.toLocaleString("de-DE") + " " + unit;
		parking.distance = distance;
	}
	sortButton.disabled = sortParkings();
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

let sortButton = document.getElementById("sort");
sortButton.onclick = updateParkingsList;

let parkings;
fetch("/api/parkings")
	.then((response) => response.json())
	.then((data) => {
		for (let parking of data) {
			parking.element = convertToElement(parking);
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
