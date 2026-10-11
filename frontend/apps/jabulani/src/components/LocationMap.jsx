import { useTheme } from "@reservations/lib";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { useEffect, useEffectEvent, useRef, useState } from "react";
import {
  coordinatesFromLocation,
  locationFromMapboxFeature,
} from "./locationUtils.js";

const accessToken = import.meta.env.VITE_MAPBOX_TOKEN;
const fallbackCenter = [0, 20];

function mapStyle(isDarkTheme) {
  return isDarkTheme
    ? "mapbox://styles/mapbox/dark-v11"
    : "mapbox://styles/mapbox/streets-v12";
}

export default function LocationMap({ value, onChange, compact = false }) {
  const { isDarkTheme } = useTheme();
  const initialCoordinates = coordinatesFromLocation(value);
  const [runtimeMapError, setRuntimeMapError] = useState("");
  const mapContainerRef = useRef(null);
  const mapRef = useRef(null);
  const markerRef = useRef(null);
  const reverseRequestRef = useRef(null);
  const shouldReverseGeocodeRef = useRef(!value?.provider_place_id);
  const initialCoordinatesRef = useRef(initialCoordinates);
  const currentStyleRef = useRef(mapStyle(isDarkTheme));
  const lastCoordinatesRef = useRef(initialCoordinates);

  const mapSupportError = !accessToken
    ? "The map is unavailable because its access token is missing."
    : !mapboxgl.supported()
      ? "The map could not start. Try enabling hardware acceleration in your browser."
      : "";
  const mapError = mapSupportError || runtimeMapError;
  const longitude = value?.geo_point?.longitude;
  const latitude = value?.geo_point?.latitude;

  const updatePin = useEffectEvent(async ({ lng, lat }) => {
    const location = {
      ...(value ?? {}),
      geo_point: { latitude: lat, longitude: lng },
    };

    onChange?.(location);

    if (!shouldReverseGeocodeRef.current) return;

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

      const enrichedLocation = locationFromMapboxFeature(
        feature,
        location.formatted_location,
        location,
        [lng, lat]
      );
      if (enrichedLocation) onChange?.(enrichedLocation);
    } catch (error) {
      if (error.name !== "AbortError") {
        console.error("Reverse geocoding failed:", error);
      }
    }
  });

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current || mapSupportError) return;

    try {
      mapboxgl.accessToken = accessToken;
      const initialCoordinates = initialCoordinatesRef.current;
      const map = new mapboxgl.Map({
        container: mapContainerRef.current,
        style: currentStyleRef.current,
        center: initialCoordinates ?? fallbackCenter,
        zoom: initialCoordinates ? 16 : 1.5,
      });

      mapRef.current = map;

      const marker = new mapboxgl.Marker({
        draggable: true,
        color: "#3b82f6",
      });
      markerRef.current = marker;

      if (initialCoordinates) {
        marker.setLngLat(initialCoordinates).addTo(map);
      }

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
      reverseRequestRef.current?.abort();
      mapRef.current?.remove();
      mapRef.current = null;
      markerRef.current = null;
    };
  }, [mapSupportError]);

  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;

    const nextStyle = mapStyle(isDarkTheme);
    if (nextStyle === currentStyleRef.current) return;

    currentStyleRef.current = nextStyle;
    map.setStyle(nextStyle);
  }, [isDarkTheme]);

  useEffect(() => {
    const map = mapRef.current;
    if (
      !map ||
      !markerRef.current ||
      !Number.isFinite(longitude) ||
      !Number.isFinite(latitude)
    ) {
      return;
    }

    const coordinates = [longitude, latitude];
    if (
      lastCoordinatesRef.current?.[0] === longitude &&
      lastCoordinatesRef.current?.[1] === latitude
    ) {
      return;
    }

    lastCoordinatesRef.current = coordinates;
    markerRef.current.setLngLat(coordinates);
    if (!markerRef.current.getElement().parentNode) {
      markerRef.current.addTo(map);
    }

    map.flyTo({ center: coordinates, zoom: 16 });
  }, [latitude, longitude]);

  return (
    <>
      <div
        className={`border-border_color relative overflow-hidden rounded-lg
          border ${compact ? "h-52 lg:h-72" : "h-72 sm:h-96"}`}
      >
        <div ref={mapContainerRef} className="h-full w-full" />
        {mapError && (
          <div
            role="alert"
            className="absolute inset-0 flex items-center justify-center
              bg-gray-100 p-6 text-center text-sm text-red-700 dark:bg-gray-900
              dark:text-red-400"
          >
            {mapError}
          </div>
        )}
      </div>

      {!mapError && (
        <p className="text-text_color/65 text-sm">
          {compact && (
            <span className="lg:hidden">
              Drag the pin or tap the map to mark the exact entrance.
            </span>
          )}
          <span className={compact ? "hidden lg:inline" : undefined}>
            Drag the pin or click anywhere on the map to mark the exact
            entrance.
          </span>
        </p>
      )}
    </>
  );
}
