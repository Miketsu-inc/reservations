import {
  Call02Icon,
  Edit03Icon,
  Location01Icon,
  Mail01Icon,
} from "@hugeicons/core-free-icons";
import {
  Button,
  Card,
  CloseButton,
  Icon,
  Input,
  Loading,
  ResponsiveDialog,
  ResponsiveDialogClose,
  ResponsiveDialogContent,
  ServerError,
} from "@reservations/components";
import { isLocationComplete } from "@reservations/jabulani/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  invalidateLocalStorageAuth,
  useToast,
  useWindowSize,
} from "@reservations/lib";
import { queryOptions, useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { lazy, Suspense, useState } from "react";
import SettingsPageHeader from "./-components/SettingsPageHeader";

const LocationEditor = lazy(
  () => import("@reservations/jabulani/components/LocationEditor.jsx")
);

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
    name: location.name.trim(),
    contact_email: location.contact_email?.trim() || null,
    phone_number: location.phone_number?.trim() || null,
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
    name: location.name?.trim() ?? "",
    contactEmail: location.contact_email?.trim() || null,
    phoneNumber: location.phone_number?.trim() || null,
    country: location.country ?? null,
    countryCode: location.country_code ?? null,
    region: location.region ?? null,
    city: location.city ?? null,
    postalCode: location.postal_code ?? null,
    address: location.address ?? null,
    latitude: location.geo_point?.latitude ?? null,
    longitude: location.geo_point?.longitude ?? null,
    geocodingProvider: location.geocoding_provider ?? null,
    providerPlaceId: location.provider_place_id ?? null,
    formattedLocation: location.formatted_location?.trim() ?? "",
  });
}

function isValidEmail(email) {
  return !email || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function areLocationDetailsValid(location) {
  return Boolean(
    location?.name?.trim() &&
    location.name.trim().length <= 100 &&
    isValidEmail(location.contact_email?.trim()) &&
    (location.phone_number?.trim().length ?? 0) <= 30
  );
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
  return (
    <Card styles="h-auto! p-4! sm:p-5!">
      <div className="flex min-w-0 items-start justify-between gap-3">
        <div className="min-w-0 flex-1 pt-1">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-fg truncate font-semibold">{location.name}</h2>
            <span
              className={`${
                location.is_active
                  ? "bg-success-subtle text-fg-success"
                  : "bg-bg-sunken text-fg-muted "
                } rounded-full px-2 py-0.5 text-xs font-medium`}
            >
              {location.is_active ? "Active" : "Inactive"}
            </span>
          </div>

          <div className="mt-3 flex min-w-0 items-start gap-2">
            <Icon
              icon={Location01Icon}
              styles="text-fg-primary mt-0.5 size-4 shrink-0"
            />
            <p className="text-fg text-sm leading-6 font-medium">
              {location.formatted_location}
            </p>
          </div>
        </div>

        <Button
          type="button"
          variant="tertiary"
          styles="shrink-0 px-2.5 py-2 text-sm sm:px-3
            [&>span:last-child]:hidden sm:[&>span:last-child]:inline"
          buttonText="Edit"
          aria-label={`Edit ${location.name}`}
          onClick={() => onEdit(location)}
        >
          <Icon icon={Edit03Icon} styles="size-4 sm:mr-2" />
        </Button>
      </div>

      {(location.contact_email || location.phone_number) && (
        <div
          className="border-border text-fg-muted mt-4 flex min-w-0 flex-wrap
            items-center gap-x-5 gap-y-2 border-t pt-3 text-xs"
        >
          {location.contact_email && (
            <div className="flex min-w-0 items-center gap-2">
              <Icon icon={Mail01Icon} styles="size-4 shrink-0" />
              <span className="truncate">{location.contact_email}</span>
            </div>
          )}
          {location.phone_number && (
            <div className="flex min-w-0 items-center gap-2">
              <Icon icon={Call02Icon} styles="size-4 shrink-0" />
              <span className="truncate">{location.phone_number}</span>
            </div>
          )}
        </div>
      )}
    </Card>
  );
}

function LocationPage() {
  const { merchantId } = useAuth();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { showToast } = useToast();
  const { isWindowSmall } = useWindowSize();
  const { data: locations, isLoading } = useQuery(
    locationSettingsQueryOptions(merchantId)
  );
  const [editingLocation, setEditingLocation] = useState(null);
  const [draftLocation, setDraftLocation] = useState(null);
  const [editorSection, setEditorSection] = useState("details");

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
    setEditorSection("details");
    updateMutation.reset();
  }

  function closeEditor() {
    if (updateMutation.isPending) return;

    setEditingLocation(null);
    setDraftLocation(null);
    updateMutation.reset();
  }

  function handleLocationDetailChange({ name, value }) {
    setDraftLocation((current) => ({ ...current, [name]: value }));
    updateMutation.reset();
  }

  const hasChanges =
    locationSignature(draftLocation) !== locationSignature(editingLocation);
  const invalidEmail = !isValidEmail(draftLocation?.contact_email?.trim());

  return (
    <div className="flex flex-col gap-6 pb-8">
      <SettingsPageHeader
        title="Locations"
        description="Manage the address and contact details for each place customers can visit."
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
          className="border-border bg-bg-raised flex flex-col items-center
            rounded-xl border px-4 py-12 text-center shadow-sm"
        >
          <div
            className="bg-primary-subtle text-fg-primary mb-4 rounded-full p-3"
          >
            <Icon icon={Location01Icon} styles="size-6" />
          </div>
          <h2 className="text-fg text-lg font-semibold">No locations yet</h2>
          <p className="text-fg-muted mt-1 max-w-sm text-sm leading-6">
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
        <ResponsiveDialogContent styles="lg:max-h-[90vh] lg:w-full lg:max-w-5xl">
          {editingLocation && (
            <div
              className="flex h-full min-h-0 w-full flex-col gap-4 p-4 sm:p-6
                lg:h-auto lg:gap-5"
            >
              <div className="flex items-start justify-between gap-4">
                <h2 className="text-fg text-xl font-semibold">Edit location</h2>
                <ResponsiveDialogClose asChild>
                  <CloseButton styles="hidden lg:block" />
                </ResponsiveDialogClose>
              </div>

              <div
                className="grid grid-cols-2 rounded-lg p-1 lg:hidden"
                role="tablist"
                aria-label="Location editor sections"
              >
                {[
                  ["details", "Details"],
                  ["address", "Address & pin"],
                ].map(([value, label]) => (
                  <button
                    key={value}
                    type="button"
                    role="tab"
                    aria-selected={editorSection === value}
                    className={`${editorSection === value ? "bg-bg-raised text-fg shadow-sm" : "text-fg-muted"}
                    rounded-md px-3 py-2 text-sm font-medium`}
                    onClick={() => setEditorSection(value)}
                  >
                    {label}
                  </button>
                ))}
              </div>

              <div
                className="min-h-0 flex-1 pt-4 lg:grid
                  lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.35fr)] lg:gap-6
                  lg:pt-0"
              >
                {(!isWindowSmall || editorSection === "details") && (
                  <section
                    className="lg:border-border flex min-h-0 flex-col gap-4
                      lg:border-r lg:pr-6"
                    role="tabpanel"
                  >
                    <div className="hidden lg:block">
                      <h3 className="text-fg font-semibold">
                        Location details
                      </h3>
                    </div>

                    <Input
                      type="text"
                      id="location_name"
                      name="name"
                      labelText="Name"
                      placeholder="Main studio"
                      maxLength={100}
                      value={draftLocation.name ?? ""}
                      inputData={handleLocationDetailChange}
                    />
                    <Input
                      type="email"
                      id="location_contact_email"
                      name="contact_email"
                      labelText="Contact email"
                      placeholder="location@example.com"
                      maxLength={320}
                      required={false}
                      value={draftLocation.contact_email ?? ""}
                      inputData={handleLocationDetailChange}
                    />
                    {invalidEmail && (
                      <p className="text-fg-danger -mt-2 text-xs">
                        Enter a valid email address.
                      </p>
                    )}
                    <Input
                      type="tel"
                      id="location_phone_number"
                      name="phone_number"
                      labelText="Phone number"
                      placeholder="Phone number"
                      maxLength={30}
                      required={false}
                      value={draftLocation.phone_number ?? ""}
                      inputData={handleLocationDetailChange}
                    />
                    <p className="text-fg-muted text-xs leading-5">
                      Leave the email blank to use the business contact email.
                    </p>
                  </section>
                )}

                {(!isWindowSmall || editorSection === "address") && (
                  <section
                    className="flex min-h-0 flex-col gap-4"
                    role="tabpanel"
                  >
                    <div className="hidden lg:block">
                      <h3 className="text-fg font-semibold">Address</h3>
                    </div>
                    <Suspense
                      fallback={
                        <div
                          className="bg-bg-sunken h-72 animate-pulse rounded-lg"
                        />
                      }
                    >
                      <LocationEditor
                        key={editingLocation.id}
                        value={draftLocation}
                        compact
                        onChange={(location) => {
                          setDraftLocation(location);
                          updateMutation.reset();
                        }}
                      />
                    </Suspense>
                  </section>
                )}
              </div>

              <ServerError error={updateMutation.error?.message} />

              <div
                className="border-border flex justify-end gap-3 border-t pt-4
                  lg:pt-5"
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
                  disabled={
                    !hasChanges ||
                    !areLocationDetailsValid(draftLocation) ||
                    !isLocationComplete(draftLocation)
                  }
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
