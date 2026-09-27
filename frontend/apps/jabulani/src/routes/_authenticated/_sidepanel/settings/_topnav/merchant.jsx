import { Loading, ServerError, Textarea } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  businessProfileSettingsQueryOptions,
  updateBusinessProfileSettings,
  useToast,
} from "@reservations/lib";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import DangerZone from "./-components/DangerZone";
import ImageUploader from "./-components/ImageUploader";
import SettingsPageHeader from "./-components/SettingsPageHeader";
import SettingsSaveAction from "./-components/SettingsSaveAction";
import SettingsSection from "./-components/SettingsSection";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_topnav/merchant"
)({
  component: MerchantPage,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(
      businessProfileSettingsQueryOptions(merchantId)
    );
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function MerchantPage() {
  const { merchantId, role } = useAuth();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { showToast } = useToast();
  const { data: profileSettings, isLoading } = useQuery(
    businessProfileSettingsQueryOptions(merchantId)
  );
  const [changes, setChanges] = useState({});

  const merchantInfo = { ...profileSettings, ...changes };
  const hasUnsavedChanges = Object.keys(changes).length > 0;

  const updateMutation = useMutation({
    mutationFn: () => updateBusinessProfileSettings(merchantId, merchantInfo),
    onSuccess: (updatedSettings) => {
      queryClient.setQueryData(
        businessProfileSettingsQueryOptions(merchantId).queryKey,
        updatedSettings
      );
      setChanges({});
      showToast({
        message: "Business profile updated successfully",
        variant: "success",
      });
    },
  });

  function handleInputData({ name, value }) {
    setChanges((current) => ({ ...current, [name]: value }));
  }

  if (isLoading) return <Loading />;

  return (
    <div className="flex flex-col gap-12 pb-28 md:pb-8">
      <SettingsPageHeader
        title="Business profile"
        description="Manage the information customers see on your public booking page."
        action={
          <SettingsSaveAction
            buttonText="Save"
            onClick={() => updateMutation.mutate()}
            disabled={!hasUnsavedChanges}
            isLoading={updateMutation.isPending}
            error={updateMutation.error}
          />
        }
      />

      <SettingsSection
        title="About your business"
        description="Use clear, customer-facing language. All fields are optional."
      >
        <div className="grid gap-5">
          <Textarea
            styles="p-3 max-h-96 min-h-28 w-full"
            id="introduction"
            placeholder="A short introduction to your business"
            name="introduction"
            required={false}
            labelText="Introduction"
            value={merchantInfo.introduction}
            inputData={handleInputData}
          />
          <Textarea
            styles="p-3 max-h-72 min-h-24 w-full"
            id="announcement"
            placeholder="Share an important update with customers"
            name="announcement"
            required={false}
            labelText="Announcement"
            value={merchantInfo.announcement}
            inputData={handleInputData}
          />
          <Textarea
            styles="p-3 max-h-96 min-h-28 w-full"
            id="about_us"
            placeholder="Tell customers more about your team and services"
            name="about_us"
            required={false}
            labelText="About us"
            value={merchantInfo.about_us}
            inputData={handleInputData}
          />
        </div>
      </SettingsSection>

      <SettingsSection
        title="Visit and payment information"
        description="Help customers prepare before they arrive."
      >
        <div className="grid gap-5 md:grid-cols-2">
          <Textarea
            styles="p-3 max-h-48 min-h-24 w-full"
            id="payment_info"
            placeholder="Accepted payment methods or payment instructions"
            name="payment_info"
            required={false}
            labelText="Payment information"
            value={merchantInfo.payment_info}
            inputData={handleInputData}
          />
          <Textarea
            styles="p-3 max-h-48 min-h-24 w-full"
            id="parking_info"
            placeholder="Parking, entrance, or arrival instructions"
            name="parking_info"
            required={false}
            labelText="Arrival and parking information"
            value={merchantInfo.parking_info}
            inputData={handleInputData}
          />
        </div>
      </SettingsSection>

      <SettingsSection
        title="Media"
        description="Add a profile image and gallery images for your public booking page. Uploads are preview-only until media storage is connected."
      >
        <div className="flex flex-col items-center gap-8 md:flex-row">
          <ImageUploader
            text="Profile picture"
            styles="h-48! w-48! shrink-0 rounded-3xl"
            imageStyles="overflow-hidden rounded-3xl object-cover!"
          />
          <div className="grid w-full gap-4 sm:grid-cols-2">
            <ImageUploader text="Gallery image 1" styles="h-48! rounded-lg" />
            <ImageUploader text="Gallery image 2" styles="h-48! rounded-lg" />
          </div>
        </div>
      </SettingsSection>

      {role === "owner" && <DangerZone />}
    </div>
  );
}
