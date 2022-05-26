L.Control.Locate = L.Control.extend({
	// _available appears to cause a naming conflict.
	_locationAvailable: false,
	getAvailable: function () {
		return this._locationAvailable;
	},
	setAvailable: function (available) {
		this._locationAvailable = available;
		if (available) this.onAvailable();
	},
	_isFirstPosition: true,
	getIsFirstPosition: function () {
		return this._isFirstPosition;
	},
	coordinates: function () {
		return L.latLng(
			this.position.coords.latitude,
			this.position.coords.longitude
		);
	},
	_button: document.createElement("button"),
	_icon: document.createElement("span"),
	options: {
		iconClass: "fa-solid fa-location-dot fa-2x",
		animationClass: "fa-fade",
		markerOptions: null,
		geolocationOptions: {
			enableHighAccuracy: true,
		},
	},
	initialize: function (onAvailable, options) {
		this.onAvailable = onAvailable;
		L.setOptions(this, options);
		this._icon.classList = this.options.iconClass;
		this._button.append(this._icon);

		if ("geolocation" in navigator)
			if ("permissions" in navigator)
				navigator.permissions.query({ name: "geolocation" }).then((status) => {
					this.setAvailable(status.state !== "denied");
					status.onchange = () => {
						if (status.state !== "denied") this.setAvailable(true);
					};
				});
			else this.setAvailable(true);
	},
	addTo(map) {
		if (!this.getAvailable()) {
			return;
		}
		L.Control.prototype.addTo.call(this, map);
	},
	onAdd: function (map) {
		// Avoid conflict with _map property of parent class.
		this._boundMap = map;
		L.DomEvent.on(this._button, "click", this.onClick, this);
		return this._button;
	},
	onRemove: function () {
		if (this._marker) this._marker.remove();
		this._stopAnimation();
		L.DomEvent.off(this._button, "click", this.onClick, this);
		if (this._geolocationWatchId) {
			navigator.geolocation.clearWatch(this._geolocationWatchId);
			this._geolocationWatchId = null;
		}
		this._boundMap = null;
		this._isFirstPosition = true;
	},
	onClick: function () {
		if (!this._geolocationWatchId) this.requestPosition();
		else if (this.position) this.panToPosition();
	},
	requestPosition: function () {
		if (!this.getAvailable() || this._geolocationWatchId) {
			return;
		}
		this._icon.classList.add(this.options.animationClass);
		this._appliedAnimationClass = this.options.animationClass;
		this._geolocationWatchId = navigator.geolocation.watchPosition(
			this._processPosition.bind(this),
			this._handleError.bind(this),
			this.options.geolocationOptions
		);
	},
	panToPosition: function () {
		this._boundMap.panTo(this.coordinates());
	},
	_processPosition: function (position) {
		this.position = position;
		if (this._marker) this._marker.remove();
		this._marker = L.circleMarker(
			this.coordinates(),
			this.options.markerOptions
		).addTo(this._boundMap);
		if (this.options.onPosition) {
			this.options.onPosition(position);
		}
		if (this.getIsFirstPosition()) {
			this._stopAnimation();
			this._isFirstPosition = false;
		}
	},
	_handleError: function () {
		this.remove();
		this.setAvailable(false);
	},
	_stopAnimation: function () {
		this._icon.classList.remove(this._appliedAnimationClass);
	},
});

L.control.locate = function (onAvailable, options) {
	return new L.Control.Locate(onAvailable, options);
};
