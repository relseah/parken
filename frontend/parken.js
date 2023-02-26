function initializeMap() {
	let map = L.map("map", {
		center: [49.41, 8.7],
		zoom: 12,
		minZoom: 10,
		maxZoom: 17,
		maxBounds: [
			[48.3, 7.95],
			[49.67, 9.61],
		],
	});
	L.tileLayer("/tiles/{z}/{x}/{y}.png", {
		attribution:
			'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	}).addTo(map);
	return map;
}

let highlightedParkingElement;
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

function redirectToNavigation(id, name) {
	url =
		"https://www.google.com/maps/dir/?api=1&travelmode=driving&destination=" +
		`P${parking.id} ${parking.name}, Heidelberg`;
	window.open(url, "_blank").focus();
}

function prependHeading(heading, value) {
	let div = document.createElement("div");
	let b = document.createElement("b");
	b.textContent = heading;
	div.append(b);
	div.append(document.createElement("br"));
	let lines = value.split("\n");
	for (let i = 0; i < lines.length; i++) {
		div.append(document.createTextNode(lines[i]));
		if (i !== lines.length - 1) div.append(document.createElement("br"));
	}
	return div;
}

function createChargingStationsDiv(parking) {
	let chargingStationsDiv = prependHeading(
		"Ladesäulen",
		parking.chargingStations
	);
	parking.chargingStationsDiv = chargingStationsDiv;
	return chargingStationsDiv;
}

function createInfoDiv(parking) {
	let div = document.createElement("div");
	div.className = "info";
	if (!chargingStationsCheckbox.checked && parking.chargingStations) {
		div.append(
			parking.chargingStationsDiv
				? parking.chargingStationsDiv
				: createChargingStationsDiv(parking)
		);
	}
	let addressDiv = prependHeading("Adresse", formatAddress(parking.address));
	addressDiv.className = "address";
	div.append(addressDiv);
	let openingHoursDiv = prependHeading("Öffnungszeiten", parking.openingHours);
	div.append(openingHoursDiv);
	let pricesDiv = prependHeading("Preise", parking.prices);
	div.append(pricesDiv);
	return div;
}

function toggleInfo(parking) {
	if (!parking.infoDiv) {
		parking.infoDiv = createInfoDiv(parking);
		parking.element.append(parking.infoDiv);
	} else {
		let style = parking.infoDiv.style;
		style.display = style.display === "none" ? "block" : "none";
	}
}

function createNavigationButton(id, name, sizeFactor) {
	let button = document.createElement("button");
	button.className = "right";
	let icon = document.createElement("span");
	icon.className = "fa-solid fa-route";
	if (sizeFactor !== 1) {
		icon.classList.add("fa-" + sizeFactor + "x");
	}
	button.append(icon);
	button.addEventListener("click", () => {
		redirectToNavigation(id, name);
	});
	return button;
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
	li.append(document.createElement("hr"));
	let navigationButton = createNavigationButton(parking.id, parking.name, 2);
	li.append(navigationButton);
	let infoButton = document.createElement("button");
	infoButton.className = "right";
	let infoIcon = document.createElement("span");
	infoIcon.className = "fa-solid fa-2x fa-circle-info";
	infoButton.onclick = () => toggleInfo(parking);
	infoButton.append(infoIcon);
	li.append(infoButton);
	let nameStrong = createNameStrong(parking);
	nameStrong.onclick = () => {
		highlightParkingElement(parking.element, false);
		parking.marker.togglePopup();
		if (highlightedParkingElement) map.panTo(parking.coordinates);
	};
	li.append(nameStrong);
	let occupancyDiv = createOccupancyDiv(parking);
	li.append(occupancyDiv);
	let distanceDiv = document.createElement("div");
	distanceDiv.textContent = "Entfernung: ";
	let distanceSpan = document.createElement("span");
	distanceSpan.textContent = "-";
	distanceDiv.append(distanceSpan);
	li.append(distanceDiv);
	parking.distanceSpan = distanceSpan;
	return li;
}

let markerIcon;
function markParking(parking) {
	if (!markerIcon) {
		markerIcon = L.elementIcon(document.createElement("span"), {
			className: "fa-solid fa-square-parking fa-2x marker",
			// The icon's width is 7/8 of its height.
			iconAnchor: [7 / 16, 0.5],
			anchorUnit: "em",
		});
	}
	let marker = L.marker(parking.coordinates, {
		icon: markerIcon,
	});
	marker.on("click", () => {
		if (highlightedParkingElement !== parking.element)
			highlightParkingElement(parking.element, true);
	});
	marker.on("popupclose", () => {
		if (highlightedParkingElement)
			highlightParkingElement(parking.element, false);
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

let parkingsUl = document.getElementById("parkings");
let displayedParkings;
function displayParkings() {
	for (parking of parkings) {
		markParking(parking);
		parkingsUl.append(parking.element);
	}
	displayedParkings = [...parkings];
}

// Has displayedParkings changed? Requires updating parkingsUl.
function updateParkingsUl() {
	let elements = [];
	for (parking of displayedParkings) {
		elements.push(parking.element);
	}
	parkingsUl.replaceChildren(...elements);
}

function sortParkings() {
	let comparator = sortByDistanceCheckbox.checked
		? (a, b) => {
				if (a.distance === b.distance) return 0;
				return a.distance < b.distance ? -1 : 1;
		  }
		: (a, b) => {
				return a.id < b.id ? -1 : 1;
		  };
	displayedParkings.sort(comparator);
	updateParkingsUl();
}
let sortByDistanceCheckbox = document.getElementById("sort-by-distance");
sortByDistanceCheckbox.addEventListener("change", sortParkings);

function filterParkings() {
	if (chargingStationsCheckbox.checked) {
		for (let i = 0; i < displayedParkings.length; i++) {
			if (!displayedParkings[i].chargingStations) {
				displayedParkings.splice(i, 1);
				i--;
			}
		}
	} else {
		for (parking of parkings) {
			if (!parking.chargingStations) {
				displayedParkings.push(parking);
			}
		}
	}
	sortParkings();
}
let chargingStationsCheckbox = document.getElementById("charging-stations");
chargingStationsCheckbox.addEventListener("change", () => {
	if (!chargingStationsCheckbox.checked) {
		for (parking of displayedParkings) {
			if (parking.infoDiv) {
				parking.infoDiv.prepend(parking.chargingStationsDiv);
			} else {
				parking.chargingStationsDiv.remove();
			}
		}
	}
	filterParkings();
	if (chargingStationsCheckbox.checked) {
		for (parking of displayedParkings) {
			if (parking.chargingStations)
				parking.element.append(
					parking.chargingStationsDiv
						? parking.chargingStationsDiv
						: createChargingStationsDiv(parking)
				);
		}
	}
});

function processPosition(position) {
	position = L.latLng(position.coords.latitude, position.coords.longitude);
	for (let parking of parkings) {
		let distance = Math.round(position.distanceTo(parking.coordinates));
		let displayDistance = distance;
		let unit;
		if (distance > 999) {
			unit = "km";
			displayDistance = distance * 0.001;
		} else unit = "m";
		parking.distanceSpan.textContent =
			displayDistance.toLocaleString("de-DE") + " " + unit;
		parking.distance = distance;
	}
	if (sortByDistanceCheckbox.checked) sortParkings();
	else if (sortByDistanceCheckbox.disabled) {
		// first position
		sortByDistanceCheckbox.disabled = false;
		sortByDistanceCheckbox.checked = true;
		sortParkings();
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

let parkings;
fetch("/api/parkings")
	.then((response) => response.json())
	.then((result) => {
		let updated = Date.parse(result.updated);
		let updatedSpan = document.getElementById("updated");
		let interval;
		function update() {
			let elapsed = (Date.now() - updated) / 60000;
			let difference;
			if (elapsed < 1) {
				difference = "weniger als eine Minute";
			} else if (elapsed < 60) {
				difference = Math.round(elapsed) + " Minuten";
			} else if (elapsed < 1440) {
				difference = Math.round(elapsed / 60) + "Stunden";
			} else {
				difference = "mehr als ein Tag";
				if (interval) clearInterval(interval);
			}
			updatedSpan.textContent = difference;
		}
		update();
		// to-do: Compute exact time to update the span.
		interval = setInterval(update, 60000);
		for (let parking of result.parkings) {
			parking.element = convertToElement(parking);
			parking.coordinates = L.latLng(
				parking.coordinates.latitude,
				parking.coordinates.longitude
			);
		}
		parkings = result.parkings;
		displayParkings();
		chargingStationsCheckbox.disabled = false;
		if (latestPosition) processPosition(latestPosition);
		locateControl.options.onPosition = processPosition;
		latestPosition = null;
	});
