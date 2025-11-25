import { SearchBox } from "@mapbox/search-js-react";
import { MapPinIcon } from "@reservations/assets";
import { Button, ServerError } from "@reservations/components";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { useCallback, useEffect, useRef, useState } from "react";

const accessToken = import.meta.env.VITE_MAPBOX_TOKEN;

// Main Location Picker Page
export default function LocationPicker({
  isSubmitDone,
  isCompleted,
  redirect,
}) {
  const [isLoading, setIsLoading] = useState(false);
  const [serverError, setServerError] = useState();

  const [selectedLocation, setSelectedLocation] = useState(null);
  const [showMap, setShowMap] = useState(false);
  const [mapInstance, setMapInstance] = useState(null);

  const mapContainerRef = useRef(null);
  const mapRef = useRef(null);
  const markerRef = useRef(null);
  const initialLocationRef = useRef(null);

  // Helper to create location object from Mapbox feature
  const createLocationFromFeature = useCallback(
    (feature, draggedCoordinates = null) => {
      const coordinates = feature.geometry?.coordinates || [0, 0];
      const properties = feature.properties || {};
      const context = properties.context || {};

      // Use routable_points if available, otherwise fall back to geometry coordinates
      const routablePoints =
        properties.coordinates?.routable_points?.[0]?.coordinates;
      const finalGeoPoint = routablePoints || coordinates;

      // Use dragged coordinates for display if provided, otherwise use geometry coordinates
      const displayCoordinates = draggedCoordinates || coordinates;

      let formatted_location = properties.full_address;

      if (context.place?.name && context.address?.name) {
        formatted_location = `${context.place?.name}, ${context.address?.name}`;
      }

      return {
        country: context.country?.name || null,
        city: context.place?.name || null,
        postal_code: context.postcode?.name || null,
        address: context.address?.name || null,
        geo_point: {
          latitude: finalGeoPoint[1],
          longitude: finalGeoPoint[0],
        }, // Routable point for backend
        place_id: properties.mapbox_id || null,
        formatted_location: formatted_location,
        // Keep these for UI/map purposes - use dragged coordinates
        coordinates: displayCoordinates,
      };
    },
    []
  );

  const handleMarkerDragEnd = useCallback(
    async (lngLat) => {
      const draggedCoords = [lngLat.lng, lngLat.lat];

      // Reverse geocode to get address info for dragged location
      try {
        const response = await fetch(
          `https://api.mapbox.com/search/geocode/v6/reverse?longitude=${lngLat.lng}&latitude=${lngLat.lat}&access_token=${accessToken}`
        );
        const data = await response.json();

        if (data.features && data.features.length > 0) {
          const feature = data.features[0];
          // Pass dragged coordinates to keep marker at exact dragged position
          setSelectedLocation(
            createLocationFromFeature(feature, draggedCoords)
          );
        } else {
          // Fallback if reverse geocoding fails
          setSelectedLocation((prev) => ({
            ...prev,
            coordinates: draggedCoords,
            geo_point: {
              latitude: lngLat.lat,
              longitude: lngLat.lng,
            },
            formatted_location: `${lngLat.lat.toFixed(6)}, ${lngLat.lng.toFixed(6)}`,
          }));
        }
      } catch (error) {
        console.error("Reverse geocoding failed:", error);
        // Fallback on error
        setSelectedLocation((prev) => ({
          ...prev,
          coordinates: draggedCoords,
          geo_point: {
            latitude: lngLat.lat,
            longitude: lngLat.lng,
          },
          formatted_location: `${lngLat.lat.toFixed(6)}, ${lngLat.lng.toFixed(6)}`,
        }));
      }
    },
    [createLocationFromFeature]
  );

  // Store initial location when map is first shown
  useEffect(() => {
    if (showMap && !initialLocationRef.current) {
      initialLocationRef.current = selectedLocation;
    }
  }, [showMap, selectedLocation]);

  // Initialize map once when first shown
  useEffect(() => {
    if (!accessToken || !showMap || mapRef.current) return;

    if (mapContainerRef.current) {
      mapboxgl.accessToken = accessToken;
      const map = new mapboxgl.Map({
        container: mapContainerRef.current,
        style: "mapbox://styles/mapbox/streets-v12",
        center: initialLocationRef.current?.coordinates || [0, 0],
        zoom: initialLocationRef.current ? 15 : 2,
      });

      mapRef.current = map;
      setMapInstance(map);

      // Add marker (hidden initially)
      markerRef.current = new mapboxgl.Marker({
        draggable: true,
        color: "#3b82f6",
      });

      // Handle marker drag
      markerRef.current.on("dragend", () => {
        const lngLat = markerRef.current.getLngLat();
        handleMarkerDragEnd(lngLat);
      });

      // Change cursor to pointer when hovering over POI layers
      const poiLayers = [
        "poi-label",
        "transit-label",
        "airport-label",
        "settlement-major-label",
        "settlement-minor-label",
        "settlement-subdivision-label",
      ];

      poiLayers.forEach((layer) => {
        map.on("mouseenter", layer, () => {
          map.getCanvas().style.cursor = "pointer";
        });
        map.on("mouseleave", layer, () => {
          map.getCanvas().style.cursor = "";
        });
      });

      // Handle clicks on POI/named places
      map.on("click", async (e) => {
        const features = map.queryRenderedFeatures(e.point, {
          layers: poiLayers,
        });

        if (features.length > 0) {
          const feature = features[0];
          const { lng, lat } = e.lngLat;
          const clickedCoords = [lng, lat];

          // Try to get full place details from Mapbox API
          try {
            const response = await fetch(
              `https://api.mapbox.com/search/geocode/v6/reverse?longitude=${lng}&latitude=${lat}&access_token=${accessToken}`
            );
            const data = await response.json();

            if (data.features && data.features.length > 0) {
              // Pass clicked coordinates to keep marker at exact clicked position
              setSelectedLocation(
                createLocationFromFeature(data.features[0], clickedCoords)
              );
            } else {
              // Fallback to basic info from map feature
              setSelectedLocation({
                country: null,
                city: null,
                postal_code: null,
                address: null,
                geo_point: {
                  latitude: lat,
                  longitude: lng,
                },
                place_id: null,
                formatted_location:
                  feature.properties.name ||
                  feature.properties.name_en ||
                  `${lat.toFixed(6)}, ${lng.toFixed(6)}`,
                coordinates: clickedCoords,
              });
            }
          } catch (error) {
            console.error("Failed to fetch place details:", error);
            // Fallback to basic info
            setSelectedLocation({
              country: null,
              city: null,
              postal_code: null,
              address: null,
              geo_point: {
                latitude: lat,
                longitude: lng,
              },
              place_id: null,
              formatted_location:
                feature.properties.name ||
                feature.properties.name_en ||
                `${lat.toFixed(6)}, ${lng.toFixed(6)}`,
              coordinates: clickedCoords,
            });
          }
        }
      });
    }

    return () => {
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
        markerRef.current = null;
        setMapInstance(null);
      }
    };
  }, [showMap, handleMarkerDragEnd, createLocationFromFeature]);

  // Update marker position when location changes
  useEffect(() => {
    if (mapRef.current && markerRef.current && selectedLocation) {
      markerRef.current.setLngLat(selectedLocation.coordinates);

      // Add marker to map if not already added
      if (!markerRef.current.getElement().parentNode) {
        markerRef.current.addTo(mapRef.current);
      }

      mapRef.current.flyTo({
        center: selectedLocation.coordinates,
        zoom: 15,
        essential: true,
      });
    }
  }, [selectedLocation]);

  const handleRetrieve = (result) => {
    const feature = result.features[0];
    if (feature) {
      setSelectedLocation(createLocationFromFeature(feature));
      setShowMap(true);
    }
  };

  async function submitHandler() {
    setIsLoading(true);
    try {
      const response = await fetch("/api/v1/merchants/location", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          country: selectedLocation.country,
          city: selectedLocation.city,
          postal_code: selectedLocation.postal_code,
          address: selectedLocation.address,
          geo_point: selectedLocation.geo_point,
          place_id: selectedLocation.place_id,
          formatted_location: selectedLocation.formatted_location,
          is_primary: true,
          is_active: true,
        }),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        redirect();
        setServerError("");
        isCompleted(true);
        isSubmitDone(true);
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="bg-layer_bg h-fit overflow-hidden rounded-lg shadow-lg">
      <div className="px-4 py-6">
        <ServerError error={serverError} />
        <h1 className="text-text_color mb-2 text-3xl font-bold">
          Where is your business located?
        </h1>
        <p className="text-gray-700 dark:text-gray-300">
          Search for an address and adjust the pin on the map for the correct
          location
        </p>
      </div>
      <div className="flex flex-col p-4">
        <div className="mb-4">
          <SearchBox
            accessToken={accessToken}
            map={mapInstance}
            mapboxgl={mapboxgl}
            onRetrieve={handleRetrieve}
            placeholder="Search for an address..."
          />
        </div>

        {/* Selected Location Display */}
        {selectedLocation && (
          <div
            className="border-primary/80 bg-primary/10 mb-4 rounded-lg border
              p-4"
          >
            <div className="flex items-start">
              <MapPinIcon styles="mt-0.5 mr-3 size-5 shrink-0 text-primary" />
              <div className="flex-1">
                <p className="mb-1 font-medium text-gray-900 dark:text-gray-200">
                  Selected Location
                </p>
                <p className="mb-2 text-sm text-gray-700 dark:text-gray-400">
                  {selectedLocation.address}
                </p>
                <p className="text-xs text-gray-500">
                  {selectedLocation.city && `${selectedLocation.city} `}
                  {selectedLocation.postal_code &&
                    `${selectedLocation.postal_code}, `}
                  {selectedLocation.country || "Unknown location"}
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Map */}
        <div
          className="border-border_color overflow-hidden rounded-lg border"
          style={{ height: "400px" }}
        >
          {showMap ? (
            <div
              ref={mapContainerRef}
              className="map-container h-full w-full rounded-lg"
            />
          ) : (
            <div
              className="flex h-full w-full items-center justify-center
                bg-gray-200 p-4 dark:bg-gray-700"
            >
              <div className="text-center">
                <MapPinIcon
                  styles="mx-auto mb-4 size-16 text-gray-400 dark:text-gray-300"
                />
                <p className="font-medium text-gray-600 dark:text-gray-300">
                  Search for an address for the map to appear
                </p>
              </div>
            </div>
          )}
        </div>

        <div className="pt-4">
          <Button
            styles="w-full py-2"
            variant="primary"
            buttonText="Continue"
            disabled={!selectedLocation}
            isLoading={isLoading}
            onClick={submitHandler}
          />
        </div>
      </div>
    </div>
  );
}
