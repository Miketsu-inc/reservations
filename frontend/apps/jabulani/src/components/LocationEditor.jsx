import { Location01Icon } from "@hugeicons/core-free-icons";
import { SearchBox } from "@mapbox/search-js-react";
import { Button, Icon } from "@reservations/components";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { useCallback, useEffect, useRef, useState } from "react";

const accessToken = import.meta.env.VITE_MAPBOX_TOKEN;
const fallbackCenter = [0, 20];

function coordinatesFromLocation(location) {
  const longitude = location?.geo_point?.longitude;
  const latitude = location?.geo_point?.latitude;

  if (!Number.isFinite(longitude) || !Number.isFinite(latitude)) return null;

  return [longitude, latitude];
}

export function isLocationComplete(location) {
  const coordinates = coordinatesFromLocation(location);

  return Boolean(
    location?.formatted_location?.trim() &&
    coordinates &&
    coordinates[0] >= -180 &&
    coordinates[0] <= 180 &&
    coordinates[1] >= -90 &&
    coordinates[1] <= 90
  );
}

export default function LocationEditor({ value, onChange }) {
  const [address, setAddress] = useState(value?.formatted_location ?? "");
  const [showMap, setShowMap] = useState(
    Boolean(coordinatesFromLocation(value))
  );
  const [searchBoxKey, setSearchBoxKey] = useState(0);
  const [mapInstance, setMapInstance] = useState(null);
  const [runtimeMapError, setRuntimeMapError] = useState("");
  const [mapSupportError] = useState(() =>
    !accessToken
      ? "The map is unavailable because its access token is missing."
      : !mapboxgl.supported()
        ? "The map could not start. Try enabling hardware acceleration in your browser."
        : ""
  );

  const mapContainerRef = useRef(null);
  const mapRef = useRef(null);
  const markerRef = useRef(null);
  const locationRef = useRef(value);
  const addressRef = useRef(address);
  const onChangeRef = useRef(onChange);
  const reverseRequestRef = useRef(null);
  const initialCoordinatesRef = useRef(coordinatesFromLocation(value));

  const mapError = mapSupportError || runtimeMapError;
  const longitude = value?.geo_point?.longitude;
  const latitude = value?.geo_point?.latitude;

  useEffect(() => {
    locationRef.current = value;
  }, [value]);

  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  const emitLocation = useCallback((location) => {
    locationRef.current = location;
    onChangeRef.current?.(location);
  }, []);

  const locationFromFeature = useCallback(
    (feature, formattedLocation, exactCoordinates) => {
      const properties = feature?.properties ?? {};
      const context = properties.context ?? {};
      const coordinates = exactCoordinates ?? feature?.geometry?.coordinates;

      if (!Array.isArray(coordinates) || coordinates.length < 2) return null;

      return {
        country: context.country?.name ?? null,
        city: context.place?.name ?? context.locality?.name ?? null,
        postal_code: context.postcode?.name ?? null,
        address:
          context.address?.name ??
          properties.address ??
          properties.name ??
          null,
        geo_point: {
          latitude: coordinates[1],
          longitude: coordinates[0],
        },
        place_id: properties.mapbox_id ?? null,
        formatted_location: formattedLocation,
      };
    },
    []
  );

  const updatePin = useCallback(
    async ({ lng, lat }) => {
      const coordinates = [lng, lat];
      const formattedLocation = addressRef.current.trim();
      const current = locationRef.current ?? {};
      const exactLocation = {
        ...current,
        geo_point: { latitude: lat, longitude: lng },
        formatted_location: formattedLocation,
      };

      emitLocation(exactLocation);

      reverseRequestRef.current?.abort();
      const controller = new AbortController();
      reverseRequestRef.current = controller;

      try {
        const url = new URL("https://api.mapbox.com/search/geocode/v6/reverse");
        url.searchParams.set("longitude", lng);
        url.searchParams.set("latitude", lat);
        url.searchParams.set("access_token", accessToken);

        const response = await fetch(url, { signal: controller.signal });
        if (!response.ok) return;

        const result = await response.json();
        const feature = result.features?.[0];
        if (!feature) return;

        const currentCoordinates = coordinatesFromLocation(locationRef.current);
        if (
          currentCoordinates?.[0] !== lng ||
          currentCoordinates?.[1] !== lat
        ) {
          return;
        }

        const enrichedLocation = locationFromFeature(
          feature,
          addressRef.current.trim(),
          coordinates
        );
        if (enrichedLocation) emitLocation(enrichedLocation);
      } catch (error) {
        if (error.name !== "AbortError") {
          console.error("Reverse geocoding failed:", error);
        }
      }
    },
    [emitLocation, locationFromFeature]
  );

  useEffect(() => {
    return () => reverseRequestRef.current?.abort();
  }, []);

  useEffect(() => {
    if (!showMap || !mapContainerRef.current || mapRef.current) return;

    if (mapSupportError) return;

    try {
      mapboxgl.accessToken = accessToken;
      const initialCoordinates = initialCoordinatesRef.current;
      const map = new mapboxgl.Map({
        container: mapContainerRef.current,
        style: "mapbox://styles/mapbox/streets-v12",
        center: initialCoordinates ?? fallbackCenter,
        zoom: initialCoordinates ? 16 : 1.5,
      });

      mapRef.current = map;
      setMapInstance(map);

      const marker = new mapboxgl.Marker({
        draggable: true,
        color: "#3b82f6",
      });
      markerRef.current = marker;

      marker.on("dragend", () => updatePin(marker.getLngLat()));
      map.on("click", ({ lngLat }) => updatePin(lngLat));
      map.on("webglcontextlost", () => {
        setRuntimeMapError(
          "The map stopped rendering. Try reloading after enabling hardware acceleration."
        );
      });
    } catch (error) {
      console.error("Map initialization failed:", error);
      queueMicrotask(() =>
        setRuntimeMapError(
          "The map could not start. Try enabling hardware acceleration in your browser."
        )
      );
    }

    return () => {
      mapRef.current?.remove();
      mapRef.current = null;
      markerRef.current = null;
    };
  }, [mapSupportError, showMap, updatePin]);

  useEffect(() => {
    if (
      !mapInstance ||
      !markerRef.current ||
      !Number.isFinite(longitude) ||
      !Number.isFinite(latitude)
    ) {
      return;
    }

    const coordinates = [longitude, latitude];

    markerRef.current.setLngLat(coordinates);
    if (!markerRef.current.getElement().parentNode) {
      markerRef.current.addTo(mapInstance);
    }

    mapInstance.flyTo({ center: coordinates, zoom: 16 });
  }, [latitude, longitude, mapInstance]);

  function handleAddressChange(nextAddress) {
    addressRef.current = nextAddress;
    setAddress(nextAddress);

    if (locationRef.current) {
      emitLocation({
        ...locationRef.current,
        formatted_location: nextAddress,
      });
    }
  }

  function handleRetrieve(result) {
    const feature = result.features?.[0];
    if (!feature) return;

    const formattedLocation = addressRef.current.trim();
    const location = locationFromFeature(feature, formattedLocation);
    if (!location) return;

    setShowMap(true);
    setSearchBoxKey((current) => current + 1);
    emitLocation(location);
  }

  function handleManualPin() {
    setShowMap(true);

    navigator.geolocation?.getCurrentPosition(
      ({ coords }) =>
        updatePin({ lng: coords.longitude, lat: coords.latitude }),
      () => {},
      { enableHighAccuracy: false, timeout: 5000, maximumAge: 300000 }
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div>
        <label className="text-text_color mb-1 block text-sm font-medium">
          Customer-facing address
        </label>
        <SearchBox
          key={searchBoxKey}
          accessToken={accessToken}
          map={mapInstance}
          mapboxgl={mapboxgl}
          marker={false}
          componentOptions={{ flyTo: false }}
          value={address}
          onChange={handleAddressChange}
          onRetrieve={handleRetrieve}
          placeholder="Type your address..."
        />
        <p className="text-text_color/60 mt-2 text-xs leading-5">
          This exact text will be shown to customers. Choose an autocomplete
          result to position the pin, then edit the text further if needed.
        </p>
      </div>

      {!showMap && address.trim() && (
        <Button
          type="button"
          variant="tertiary"
          styles="w-fit px-3 py-2"
          buttonText="Set pin manually"
          onClick={handleManualPin}
        />
      )}

      <div
        className="border-border_color relative h-72 overflow-hidden rounded-lg
          border sm:h-96"
      >
        {showMap ? (
          <>
            <div ref={mapContainerRef} className="h-full w-full" />
            {mapError && (
              <div
                role="alert"
                className="absolute inset-0 flex items-center justify-center
                  bg-gray-100 p-6 text-center text-sm text-red-700
                  dark:bg-gray-900 dark:text-red-400"
              >
                {mapError}
              </div>
            )}
          </>
        ) : (
          <div
            className="flex h-full items-center justify-center bg-gray-100 p-6
              dark:bg-gray-800"
          >
            <div className="text-center">
              <Icon
                icon={Location01Icon}
                styles="mx-auto mb-3 size-14 text-gray-400"
              />
              <p className="text-text_color font-medium">
                Choose an address to position the pin
              </p>
              <p className="text-text_color/60 mt-1 text-sm">
                If there is no matching result, type the address and set the pin
                manually.
              </p>
            </div>
          </div>
        )}
      </div>

      {showMap && !mapError && (
        <p className="text-text_color/65 text-sm">
          Drag the pin or click anywhere on the map to mark the exact entrance.
        </p>
      )}
    </div>
  );
}
