export function coordinatesFromLocation(location) {
  const longitude = location?.geo_point?.longitude;
  const latitude = location?.geo_point?.latitude;

  if (!Number.isFinite(longitude) || !Number.isFinite(latitude)) return null;

  return [longitude, latitude];
}

export function locationFromMapboxFeature(
  feature,
  formattedLocation,
  currentLocation,
  exactCoordinates
) {
  const properties = feature?.properties ?? {};
  const context = properties.context ?? {};
  const coordinates = exactCoordinates ?? feature?.geometry?.coordinates;
  const providerPlaceId = properties.mapbox_id ?? null;

  if (!Array.isArray(coordinates) || coordinates.length < 2) return null;

  return {
    ...(currentLocation ?? {}),
    country: context.country?.name ?? null,
    country_code: context.country?.country_code ?? null,
    region: context.region?.name ?? null,
    city: context.place?.name ?? context.locality?.name ?? null,
    postal_code: context.postcode?.name ?? null,
    address:
      context.address?.name ?? properties.address ?? properties.name ?? null,
    geo_point: {
      latitude: coordinates[1],
      longitude: coordinates[0],
    },
    geocoding_provider: providerPlaceId ? "mapbox" : null,
    provider_place_id: providerPlaceId,
    formatted_location: formattedLocation,
  };
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
