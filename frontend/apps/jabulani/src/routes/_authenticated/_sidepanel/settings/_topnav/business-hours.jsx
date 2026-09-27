import { Loading, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  businessHoursSettingsQueryOptions,
  preferencesQueryOptions,
  updateBusinessHoursSettings,
  useToast,
} from "@reservations/lib";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import BusinessHours from "./-components/BusinessHours";
import SettingsPageHeader from "./-components/SettingsPageHeader";
import SettingsSaveAction from "./-components/SettingsSaveAction";
import SettingsSection from "./-components/SettingsSection";

const daysOfWeek = {
  0: "Sunday",
  1: "Monday",
  2: "Tuesday",
  3: "Wednesday",
  4: "Thursday",
  5: "Friday",
  6: "Saturday",
};

const emptyBusinessHours = {
  0: [],
  1: [],
  2: [],
  3: [],
  4: [],
  5: [],
  6: [],
};

function validateBusinessHours(hours) {
  for (const day in hours) {
    const sortedPeriods = [...hours[day]].sort((a, b) =>
      a.start_time.localeCompare(b.start_time)
    );

    for (let index = 0; index < sortedPeriods.length; index += 1) {
      const { start_time: startTime, end_time: endTime } = sortedPeriods[index];

      if (startTime >= endTime) {
        return `${daysOfWeek[day]} has an invalid time range.`;
      }

      if (index > 0 && sortedPeriods[index - 1].end_time > startTime) {
        return `${daysOfWeek[day]} has overlapping opening hours.`;
      }
    }
  }

  return "";
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_topnav/business-hours"
)({
  component: BusinessHoursPage,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId, employeeId },
    },
  }) => {
    await Promise.all([
      queryClient.ensureQueryData(
        businessHoursSettingsQueryOptions(merchantId)
      ),
      queryClient.ensureQueryData(
        preferencesQueryOptions(merchantId, employeeId)
      ),
    ]);
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function BusinessHoursPage() {
  const { merchantId, employeeId } = useAuth();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { showToast } = useToast();
  const { data: savedBusinessHours, isLoading } = useQuery(
    businessHoursSettingsQueryOptions(merchantId)
  );
  const { data: preferences } = useQuery(
    preferencesQueryOptions(merchantId, employeeId)
  );
  const [businessHoursChanges, setBusinessHoursChanges] = useState(null);
  const businessHours =
    businessHoursChanges ?? savedBusinessHours ?? emptyBusinessHours;

  const errorMessage = businessHours
    ? validateBusinessHours(businessHours)
    : "";
  const hasUnsavedChanges = businessHoursChanges !== null;

  const updateMutation = useMutation({
    mutationFn: () => updateBusinessHoursSettings(merchantId, businessHours),
    onSuccess: (updatedBusinessHours) => {
      queryClient.setQueryData(
        businessHoursSettingsQueryOptions(merchantId).queryKey,
        updatedBusinessHours
      );
      queryClient.invalidateQueries({
        queryKey: [merchantId, "normalized-business-hours"],
      });
      setBusinessHoursChanges(null);
      showToast({
        message: "Business hours updated successfully",
        variant: "success",
      });
    },
  });

  if (isLoading) return <Loading />;

  return (
    <div className="flex flex-col gap-12 pb-28 md:pb-8">
      <SettingsPageHeader
        title="Business hours"
        description="Set the regular hours customers can expect your business to be open."
        action={
          <SettingsSaveAction
            buttonText="Save"
            onClick={() => updateMutation.mutate()}
            disabled={!hasUnsavedChanges || Boolean(errorMessage)}
            isLoading={updateMutation.isPending}
            error={updateMutation.error}
          />
        }
      />
      <SettingsSection
        title="Weekly opening hours"
        description="Add a second time range when your day includes a break. Availability can still vary by team member and service."
      >
        <BusinessHours
          data={businessHours}
          setBusinessHours={(updater) =>
            setBusinessHoursChanges((current) =>
              updater(
                current ?? savedBusinessHours ?? emptyBusinessHours
              )
            )
          }
          preferences={preferences}
        />
        {errorMessage && (
          <p className="mt-4 text-sm text-red-600 dark:text-red-400">
            {errorMessage}
          </p>
        )}
      </SettingsSection>
    </div>
  );
}
