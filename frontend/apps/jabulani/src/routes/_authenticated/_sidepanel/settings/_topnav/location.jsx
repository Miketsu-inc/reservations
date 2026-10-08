import { Location01Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  Card,
  Icon,
  Loading,
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
  const response = await fetch(`/api/v1/merchants/${merchantId}/locations`, {
    method: "PATCH",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
    body: JSON.stringify(updatedLocation),
  });

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    const result = await response.json();
    throw result.error;
  }

  return updatedLocation;
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

function Detail({ label, value }) {
  if (!value) return null;

  return (
    <div>
      <dt
        className="text-text_color/60 text-xs font-medium tracking-wide
          uppercase"
      >
        {label}
      </dt>
      <dd className="text-text_color mt-1 text-sm">{value}</dd>
    </div>
  );
}

function LocationPage() {
  const { merchantId } = useAuth();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { showToast } = useToast();
  const { data: settings, isLoading } = useQuery(
    locationSettingsQueryOptions(merchantId)
  );
  const [isEditing, setIsEditing] = useState(false);
  const [draftLocation, setDraftLocation] = useState(null);

  const updateMutation = useMutation({
    mutationFn: (location) => updateLocationSettings(merchantId, location),
    onSuccess: (updatedLocation) => {
      queryClient.setQueryData(
        locationSettingsQueryOptions(merchantId).queryKey,
        updatedLocation
      );
      setDraftLocation(updatedLocation);
      setIsEditing(false);
      showToast({
        message: "Location updated successfully",
        variant: "success",
      });
    },
  });

  if (isLoading) return <Loading />;

  function startEditing() {
    setDraftLocation(settings);
    setIsEditing(true);
  }

  function cancelEditing() {
    setDraftLocation(settings);
    setIsEditing(false);
    updateMutation.reset();
  }

  return (
    <div className="flex flex-col gap-12 pb-8">
      <SettingsPageHeader
        title="Location"
        description="Manage the primary address shown to customers and used for new bookings."
      />
      <Card styles="h-auto! max-w-3xl p-6!">
        <h2 className="text-text_color text-xl font-semibold">
          Primary location
        </h2>
        <p className="text-text_color/65 mt-1 mb-8 text-sm leading-6">
          The address text and exact map pin can be updated independently.
        </p>

        {isEditing ? (
          <div className="flex flex-col gap-6">
            <LocationEditor value={draftLocation} onChange={setDraftLocation} />
            <ServerError error={updateMutation.error?.message} />
            <div
              className="border-border_color flex flex-col-reverse gap-3
                border-t pt-5 sm:flex-row sm:justify-end"
            >
              <Button
                type="button"
                variant="tertiary"
                styles="px-4 py-2"
                buttonText="Cancel"
                disabled={updateMutation.isPending}
                onClick={cancelEditing}
              />
              <Button
                type="button"
                styles="px-4 py-2"
                buttonText="Save location"
                disabled={!isLocationComplete(draftLocation)}
                isLoading={updateMutation.isPending}
                onClick={() => updateMutation.mutate(draftLocation)}
              />
            </div>
          </div>
        ) : (
          <div className="flex flex-col gap-6">
            <div className="flex min-w-0 items-center gap-4">
              <div
                className="bg-primary/10 text-primary shrink-0 rounded-full p-3"
              >
                <Icon icon={Location01Icon} styles="size-6" />
              </div>
              <p className="text-text_color min-w-0 flex-1 font-medium">
                {settings.formatted_location ||
                  "No formatted address available"}
              </p>
            </div>
            <dl className="grid gap-5 sm:grid-cols-2 sm:pl-16">
              <Detail label="Address" value={settings.address} />
              <Detail label="City" value={settings.city} />
              <Detail label="Postal code" value={settings.postal_code} />
              <Detail label="Country" value={settings.country} />
            </dl>
            <div
              className="border-border_color flex border-t pt-5 sm:justify-end"
            >
              <Button
                type="button"
                variant="tertiary"
                styles="w-full px-3 py-2 sm:w-auto"
                buttonText="Change location"
                onClick={startEditing}
              />
            </div>
          </div>
        )}
      </Card>
    </div>
  );
}
