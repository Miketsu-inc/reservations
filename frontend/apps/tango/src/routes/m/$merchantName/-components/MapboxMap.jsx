import { useTheme } from "@reservations/lib";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { useEffect, useRef } from "react";

const accessToken = import.meta.env.VITE_MAPBOX_TOKEN;

function mapStyle(isDarkTheme) {
  return isDarkTheme
    ? "mapbox://styles/mapbox/dark-v11"
    : "mapbox://styles/mapbox/streets-v12";
}

export default function MapboxMap({
  styles,
  coordinates,
  minHeight = 400,
  zoom = 15,
}) {
  const { isDarkTheme } = useTheme();
  const mapContainerRef = useRef(null);
  const mapRef = useRef(null);
  const markerRef = useRef(null);
  const currentStyleRef = useRef(mapStyle(isDarkTheme));

  useEffect(() => {
    if (!accessToken || !coordinates || mapRef.current) return;

    if (mapContainerRef.current) {
      mapboxgl.accessToken = accessToken;

      let map = undefined;

      try {
        map = new mapboxgl.Map({
          container: mapContainerRef.current,
          style: currentStyleRef.current,
          center: coordinates,
          zoom: zoom,
          interactive: false, // Disable all interactions
        });
      } catch (error) {
        console.log(error);
        return;
      }

      mapRef.current = map;

      // Add non-draggable marker
      markerRef.current = new mapboxgl.Marker({
        draggable: false,
        color: "#3b82f6",
      })
        .setLngLat(coordinates)
        .addTo(map);
    }

    return () => {
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
        markerRef.current = null;
      }
    };
  }, [coordinates, zoom]);

  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;

    const nextStyle = mapStyle(isDarkTheme);
    if (nextStyle === currentStyleRef.current) return;

    currentStyleRef.current = nextStyle;
    map.setStyle(nextStyle);
  }, [isDarkTheme]);

  return (
    <div
      ref={mapContainerRef}
      className={`${styles} bg-bg-sunken h-full w-full rounded-lg`}
      style={{ minHeight: minHeight }}
    />
  );
}
