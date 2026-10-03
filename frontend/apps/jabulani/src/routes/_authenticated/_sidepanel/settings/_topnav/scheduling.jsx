import { Loading, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { queryOptions, useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import SchedulingSettings from "./-components/SchedulingSettings";
import SettingsPageHeader from "./-components/SettingsPageHeader";
import SettingsSaveAction from "./-components/SettingsSaveAction";
import SettingsSection from "./-components/SettingsSection";

async function fetchSchedulingSettings(merchantId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/settings/scheduling`
  );
  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

function schedulingSettingsQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "scheduling-settings"],
    queryFn: () => fetchSchedulingSettings(merchantId),
  });
}

async function updateSchedulingSettings(merchantId, settings) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/settings/scheduling`,
    {
      method: "PATCH",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify(settings),
    }
  );

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    const result = await response.json();
    throw result.error;
  }

  return settings;
}

const approvalOptions = [
  {
    value: "auto",
    name: "Automatic",
    description: "Confirm every booking immediately.",
  },
  {
    value: "manual",
    name: "Manual",
    description: "Require approval for every booking request.",
  },
  {
    value: "manual_for_new",
    name: "Manual for new customers",
    description: "Only require approval for first-time customers.",
  },
];

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_topnav/scheduling"
)({
  component: SchedulingPage,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(
      schedulingSettingsQueryOptions(merchantId)
    );
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function SchedulingPage() {
  const { merchantId } = useAuth();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { showToast } = useToast();
  const { data: savedSettings, isLoading } = useQuery(
    schedulingSettingsQueryOptions(merchantId)
  );
  const [changes, setChanges] = useState({});
  const settings = { ...savedSettings, ...changes };

  const updateMutation = useMutation({
    mutationFn: () => updateSchedulingSettings(merchantId, settings),
    onSuccess: (updatedSettings) => {
      queryClient.setQueryData(
        schedulingSettingsQueryOptions(merchantId).queryKey,
        updatedSettings
      );
      setChanges({});
      showToast({
        message: "Scheduling settings updated successfully",
        variant: "success",
      });
    },
  });

  function handleChange({ name, value }) {
    setChanges((current) => ({ ...current, [name]: value }));
  }

  if (isLoading) return <Loading />;

  return (
    <div className="flex flex-col gap-12 pb-28 md:pb-8">
      <SettingsPageHeader
        title="Scheduling"
        description="Set the default booking rules used by your services. Individual services can override these defaults."
        action={
          <SettingsSaveAction
            buttonText="Save"
            onClick={() => updateMutation.mutate()}
            disabled={Object.keys(changes).length === 0}
            isLoading={updateMutation.isPending}
            error={updateMutation.error}
          />
        }
      />

      <SchedulingSettings settings={settings} onChange={handleChange} />

      <SettingsSection
        title="Booking approval"
        description="Choose when a booking request becomes confirmed."
      >
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          {approvalOptions.map((option) => {
            const active = settings.approval_policy === option.value;
            return (
              <button
                key={option.value}
                type="button"
                aria-pressed={active}
                onClick={() =>
                  handleChange({
                    name: "approval_policy",
                    value: option.value,
                  })
                }
                className={`rounded-xl border p-4 text-left transition-colors ${
                  active
                    ? "border-primary bg-primary/10"
                    : `border-border_color hover:border-primary/50
                      hover:bg-hvr_gray`
                }`}
              >
                <span className="text-text_color block text-sm font-semibold">
                  {option.name}
                </span>
                <span className="text-text_color/65 mt-1 block text-xs leading-5">
                  {option.description}
                </span>
              </button>
            );
          })}
        </div>
      </SettingsSection>

    </div>
  );
}
