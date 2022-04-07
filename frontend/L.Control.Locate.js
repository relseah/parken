L.Control.Locate = L.Control.extend({
    position: null,
    coordinates: function() {
        return L.latLng(this.position.coords.latitude, this.position.coords.longitude);
    },
    _button: document.createElement('button'),
    _icon: document.createElement('span'),
    _map: null,
    _marker: null,
    options: {
        iconClass: 'fa-solid fa-location-dot fa-2x',
        animationClass: 'fa-fade',
        markerOptions: {
            radius: 5,
            color: 'steelblue',
            fillOpacity: 1
        }
    },
    initialize: function(options) {
        L.setOptions(this, options);
        this._icon.classList = this.options.iconClass;
        this._button.appendChild(this._icon);
    },
    onAdd: function(map) {
        this._map = map;
        L.DomEvent.on(this._button, 'click', this.requestPosition, this);
        return this._button;
    },
    onRemove: function() {
        L.DomEvent.off(_button, 'click', this.requestPosition, this);
        this._map = null;
    },
    requestPosition: function() {
        this.startAnimation();
        navigator.geolocation.getCurrentPosition(this._processPosition.bind(this), null, {enableHighAccuracy: true});
    },
    _processPosition: function(position)  {
        if (this._marker != null) {
            this._marker.remove();
        }
        this.position = position;
        this._marker = L.circleMarker(this.coordinates(), this.options.markerOptions).addTo(map);
        map.flyTo(this.coordinates());
        this.stopAnimation();
    },
    startAnimation: function() {
        if (!this._isAnimated) {
            this._icon.classList.add(this.options.animationClass);
            this._isAnimated = true;
        }
    },
    stopAnimation: function() {
        if(this._isAnimated) {
            this._icon.classList.remove(this.options.animationClass)
            this._isAnimated = false;
        }
    }
});

L.control.locate = function(options) {
    return new L.Control.Locate(options);
};
