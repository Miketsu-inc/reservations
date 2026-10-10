import { Edit03Icon, Location01Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  Card,
  CloseButton,
  Icon,
  Loading,
  ResponsiveDialog,
  ResponsiveDialogClose,
  ResponsiveDialogContent,
  ServerError,
} from "@reservations/components";
import {
  isLocationComplete,
  LocationEditor,
} from "@reservations/jabulani/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { queryOptions, useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import SettingsPageHeader from "./-components/SettingsPageHeader";

async function fetchLocationSettings(merchantId) {
  const response = await fetch(`/api/v1/merchants/${merchantId}/locations`);
  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

function locationSettingsQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "location-settings"],
    queryFn: () => fetchLocationSettings(merchantId),
  });
}

async function updateLocationSettings(merchantId, location) {
  const updatedLocation = {
    ...location,
    formatted_location: location.formatted_location.trim(),
  };
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/locations/${location.id}`,
    {
      method: "PATCH",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify(updatedLocation),
    }
  );

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    const result = await response.json();
    throw result.error;
  }

  return updatedLocation;
}

function locationSignature(location) {
  if (!location) return "";

  return JSON.stringify({
    country: location.country ?? null,
    city: location.city ?? null,
    postalCode: location.postal_code ?? null,
    address: location.address ?? null,
    latitude: location.geo_point?.latitude ?? null,
    longitude: location.geo_point?.longitude ?? null,
    placeId: location.place_id ?? null,
    formattedLocation: location.formatted_location?.trim() ?? "",
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_topnav/location"
)({
  component: LocationPage,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(locationSettingsQueryOptions(merchantId));
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function LocationCard({ location, onEdit }) {
  const secondaryAddress = [
    location.postal_code,
    location.city,
    location.country,
  ]
    .filter(Boolean)
    .join(", ");

  return (
    <Card styles="h-auto! flex flex-col justify-between gap-5 p-5!">
      <div className="flex min-w-0 items-start gap-4">
        <div className="bg-primary/10 text-primary shrink-0 rounded-full p-3">
          <Icon icon={Location01Icon} styles="size-6" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="mb-2 flex flex-wrap items-center gap-2">
            <h2 className="text-text_color font-semibold">
              {location.name ??
                (location.is_primary ? "Primary location" : "Location")}
            </h2>
            {location.is_primary && (
              <span
                className="bg-primary/10 text-primary rounded-full px-2 py-0.5
                  text-xs font-medium"
              >
                Primary
              </span>
            )}
            {!location.is_active && (
              <span
                className="rounded-full bg-gray-200 px-2 py-0.5 text-xs
                  font-medium text-gray-600 dark:bg-gray-700 dark:text-gray-300"
              >
                Inactive
              </span>
            )}
          </div>
          <p className="text-text_color text-sm leading-6 font-medium">
            {location.formatted_location}
          </p>
          {secondaryAddress && (
            <p className="text-text_color/55 mt-1 text-xs leading-5">
              {secondaryAddress}
            </p>
          )}
        </div>
      </div>

      <div className="border-border_color flex justify-end border-t pt-4">
        <Button
          type="button"
          variant="tertiary"
          styles="px-3 py-2"
          buttonText="Edit location"
          onClick={() => onEdit(location)}
        >
          <Icon icon={Edit03Icon} styles="mr-2 size-4" />
        </Button>
      </div>
    </Card>
  );
}

function LocationPage() {
  const { merchantId } = useAuth();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { showToast } = useToast();
  const { data: locations, isLoading } = useQuery(
    locationSettingsQueryOptions(merchantId)
  );
  const [editingLocation, setEditingLocation] = useState(null);
  const [draftLocation, setDraftLocation] = useState(null);

  const updateMutation = useMutation({
    mutationFn: (location) => updateLocationSettings(merchantId, location),
    onSuccess: (updatedLocation) => {
      queryClient.setQueryData(
        locationSettingsQueryOptions(merchantId).queryKey,
        (currentLocations) =>
          currentLocations.map((location) =>
            location.id === updatedLocation.id ? updatedLocation : location
          )
      );
      setEditingLocation(null);
      setDraftLocation(null);
      showToast({
        message: "Location updated successfully",
        variant: "success",
      });
    },
  });

  if (isLoading) return <Loading />;

  function openEditor(location) {
    setEditingLocation(location);
    setDraftLocation(location);
    updateMutation.reset();
  }

  function closeEditor() {
    if (updateMutation.isPending) return;

    setEditingLocation(null);
    setDraftLocation(null);
    updateMutation.reset();
  }

  const hasChanges =
    locationSignature(draftLocation) !== locationSignature(editingLocation);

  return (
    <div className="flex flex-col gap-6 pb-8">
      <SettingsPageHeader
        title="Locations"
        description="Manage the places customers can visit for their bookings."
      />

      {locations.length > 0 ? (
        <div className="grid gap-4 lg:grid-cols-2">
          {locations.map((location) => (
            <LocationCard
              key={location.id}
              location={location}
              onEdit={openEditor}
            />
          ))}
        </div>
      ) : (
        <div
          className="border-border_color bg-layer_bg flex flex-col items-center
            rounded-xl border px-4 py-12 text-center shadow-sm"
        >
          <div className="bg-primary/10 text-primary mb-4 rounded-full p-3">
            <Icon icon={Location01Icon} styles="size-6" />
          </div>
          <h2 className="text-text_color text-lg font-semibold">
            No locations yet
          </h2>
          <p className="text-text_color/65 mt-1 max-w-sm text-sm leading-6">
            Add a business location before accepting in-person bookings.
          </p>
        </div>
      )}

      <ResponsiveDialog
        open={Boolean(editingLocation)}
        onOpenChange={(open) => {
          if (!open) closeEditor();
        }}
      >
        <ResponsiveDialogContent
          styles="lg:max-h-[90vh] lg:max-w-3xl lg:overflow-y-auto"
          popUpStyles="max-h-[calc(95dvh+3rem)]!"
        >
          {editingLocation && (
            <div className="flex w-full flex-col gap-4 p-4 sm:gap-5 sm:p-6">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <h2 className="text-text_color text-xl font-semibold">
                    Edit location
                  </h2>
                  <p
                    className="text-text_color/65 mt-1 hidden text-sm leading-6
                      lg:block"
                  >
                    Update the customer-facing address and exact map pin.
                  </p>
                </div>
                <ResponsiveDialogClose asChild>
                  <CloseButton styles="hidden lg:block" />
                </ResponsiveDialogClose>
              </div>

              <LocationEditor
                key={editingLocation.id}
                value={draftLocation}
                compact
                onChange={(location) => {
                  setDraftLocation(location);
                  updateMutation.reset();
                }}
              />

              <ServerError error={updateMutation.error?.message} />

              <div
                className="border-border_color flex justify-end gap-3 border-t
                  pt-4 lg:pt-5"
              >
                <ResponsiveDialogClose asChild>
                  <Button
                    type="button"
                    variant="tertiary"
                    styles="hidden px-4 py-2 lg:flex"
                    buttonText="Cancel"
                    disabled={updateMutation.isPending}
                  />
                </ResponsiveDialogClose>
                <Button
                  type="button"
                  styles="w-full px-4 py-2 lg:w-auto"
                  buttonText="Save location"
                  disabled={!hasChanges || !isLocationComplete(draftLocation)}
                  isLoading={updateMutation.isPending}
                  onClick={() => updateMutation.mutate(draftLocation)}
                />
              </div>
            </div>
          )}
        </ResponsiveDialogContent>
      </ResponsiveDialog>
    </div>
  );
}
