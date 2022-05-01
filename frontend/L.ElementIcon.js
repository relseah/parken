L.ElementIcon = L.Icon.extend({
	options: {
		sizeUnit: "px",
		anchorUnit: "px",
	},
	initialize: function (element, options) {
		this.element = element;
		L.setOptions(this, options);
	},
	createIcon: function (oldIcon) {
		this.element.classList.add("leaflet-marker-icon");
		if (this.options.className) {
			for (let className of this.options.className.split(" "))
				this.element.classList.add(className);
		}

		let size = this.options.size;
		if (typeof size === "number") {
			size = L.point(size, size);
		} else if (size) {
			// The factory function will return the argument if an instance of Point was passed.
			size = L.point(size);
		}
		let anchor;
		let defaultAnchor = false;
		if (this.options.iconAnchor) {
			anchor = L.point(this.options.iconAnchor);
		} else if (size) {
			anchor = size.divideBy(2);
			defaultAnchor = true;
		}

		if (anchor) {
			let anchorUnit = defaultAnchor
				? this.options.sizeUnit
				: this.options.anchorUnit;
			this.element.style.marginLeft = -anchor.x + anchorUnit;
			this.element.style.marginTop = -anchor.y + anchorUnit;
		}
		if (size) {
			this.element.style.width = size.x + this.options.sizeUnit;
			this.element.style.height = size.y + this.options.sizeUnit;
		}

		if (oldIcon && oldIcon.isEqualNode(this.element)) {
			return oldIcon;
		}
		return this.element.cloneNode(true);
	},
	createShadow: function () {
		return null;
	},
});

L.elementIcon = function elementIcon(element, options) {
	return new L.ElementIcon(element, options);
};
