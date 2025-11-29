import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { useEffect, useRef } from "react";

const accessToken = import.meta.env.VITE_MAPBOX_TOKEN;

export default function MapboxMap({
  styles,
  coordinates,
  minHeight = 400,
  zoom = 15,
}) {
  const mapContainerRef = useRef(null);
  const mapRef = useRef(null);
  const markerRef = useRef(null);

  useEffect(() => {
    if (!accessToken || !coordinates || mapRef.current) return;

    if (mapContainerRef.current) {
      mapboxgl.accessToken = accessToken;
      const map = new mapboxgl.Map({
        container: mapContainerRef.current,
        style: "mapbox://styles/mapbox/streets-v12",
        center: coordinates,
        zoom: zoom,
        interactive: false, // Disable all interactions
      });

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

  return (
    <div
      ref={mapContainerRef}
      className={`${styles} h-full w-full rounded-lg bg-red-400`}
      style={{ minHeight: minHeight }}
    />
  );
}
