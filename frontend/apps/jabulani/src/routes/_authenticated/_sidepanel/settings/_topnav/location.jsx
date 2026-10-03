import { Location01Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  Card,
  Icon,
  Loading,
  ServerError,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
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
  const { data: settings, isLoading } = useQuery(
    locationSettingsQueryOptions(merchantId)
  );

  if (isLoading) return <Loading />;

  return (
    <div className="flex flex-col gap-12 pb-8">
      <SettingsPageHeader
        title="Location"
        description="Review the primary location shown to customers and attached to bookings."
      />
      <Card styles="h-auto! max-w-2xl p-6!">
        <h2 className="text-text_color text-xl font-semibold">
          Primary location
        </h2>
        <p className="text-text_color/65 mt-1 mb-8 text-sm leading-6">
          Location editing is not available yet. Your current location is shown
          below.
        </p>
        <div className="flex flex-col gap-6">
          <div className="flex min-w-0 items-center gap-4">
            <div
              className="bg-primary/10 text-primary shrink-0 rounded-full p-3"
            >
              <Icon icon={Location01Icon} styles="size-6" />
            </div>
            <p className="text-text_color min-w-0 flex-1 font-medium">
              {settings.formatted_location || "No formatted address available"}
            </p>
          </div>
          <dl className="grid gap-5 sm:grid-cols-2 sm:pl-16">
            <Detail label="Address" value={settings.address} />
            <Detail label="City" value={settings.city} />
            <Detail label="Postal code" value={settings.postal_code} />
            <Detail label="Country" value={settings.country} />
          </dl>
          <div className="border-border_color flex border-t pt-5 sm:justify-end">
            <Button
              variant="tertiary"
              styles="w-full px-3 py-2 sm:w-auto"
              buttonText="Change location"
            />
          </div>
        </div>
      </Card>
    </div>
  );
}
