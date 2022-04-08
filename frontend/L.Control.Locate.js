L.Control.Locate = L.Control.extend({
	position: null,
	_firstPosition: true,
	coordinates: function () {
		return L.latLng(
			this.position.coords.latitude,
			this.position.coords.longitude
		);
	},
	_geolocationWatchId: null,
	_button: document.createElement("button"),
	_icon: document.createElement("span"),
	// Avoid conflict with _map property of parent class.
	_boundMap: null,
	_marker: null,
	options: {
		iconClass: "fa-solid fa-location-dot fa-2x",
		animationClass: "fa-fade",
		markerOptions: {
			radius: 5,
			color: "steelblue",
			fillOpacity: 1,
		},
		geolocationOptions: {
			enableHighAccuracy: true,
			timeout: 10000,
		},
	},
	initialize: function (options) {
		L.setOptions(this, options);
		this._icon.classList = this.options.iconClass;
		this._button.appendChild(this._icon);
	},
	addTo(map) {
		L.Control.prototype.addTo.call(this, map);
		this.requestPosition();
	},
	onAdd: function (map) {
		this._boundMap = map;
		L.DomEvent.on(this._button, "click", this.onClick, this);
		return this._button;
	},
	onRemove: function () {
		L.DomEvent.off(this._button, "click", this.onClick, this);
		navigator.geolocation.clearWatch(this._geolocationWatchId);
		this._boundMap = null;
	},
	onClick: function () {
		if (this.position != null) this.flyToPosition();
	},
	requestPosition: function () {
		this._icon.classList.add(this.options.animationClass);
		this._geolocationWatchId = navigator.geolocation.watchPosition(
			this._processPosition.bind(this),
			this._handleError.bind(this),
			this.options.geolocationOptions
		);
	},
	flyToPosition: function () {
		this._boundMap.flyTo(this.coordinates());
	},
	_processPosition: function (position) {
		if (this._marker != null) this._marker.remove();
		this.position = position;
		this._marker = L.circleMarker(
			this.coordinates(),
			this.options.markerOptions
		).addTo(this._boundMap);
		if (this._firstPosition) {
			this.flyToPosition();
			this._icon.classList.remove(this.options.animationClass);
			this._firstPosition = false;
		}
	},
	_handleError: function () {
		this.remove();
	},
});

L.control.locate = function (options) {
	return new L.Control.Locate(options);
};
