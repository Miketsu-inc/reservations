import {
  ImageCropDialog,
  Loading,
  ServerError,
  Textarea,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  deleteImage,
  invalidateLocalStorageAuth,
  uploadImage,
  useToast,
} from "@reservations/lib";
import { queryOptions, useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import DangerZone from "./-components/DangerZone";
import LogoUploader from "./-components/LogoUploader";
import SettingsPageHeader from "./-components/SettingsPageHeader";
import SettingsSaveAction from "./-components/SettingsSaveAction";
import SettingsSection from "./-components/SettingsSection";

async function fetchMerchantProfileSettings(merchantId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/settings/merchant-profile`
  );
  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

function merchantProfileSettingsQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "merchant-profile-settings"],
    queryFn: () => fetchMerchantProfileSettings(merchantId),
  });
}

async function updateMerchantProfileSettings(merchantId, settings) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/settings/merchant-profile`,
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
      merchantProfileSettingsQueryOptions(merchantId)
    );
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function MerchantPage() {
  const { merchantId, role } = useAuth();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { showToast } = useToast();
  const { data: profileSettings, isLoading } = useQuery(
    merchantProfileSettingsQueryOptions(merchantId)
  );
  const [changes, setChanges] = useState({});

  const merchantInfo = { ...profileSettings, ...changes };
  const hasUnsavedChanges = Object.keys(changes).length > 0;

  const updateMutation = useMutation({
    mutationFn: () => updateMerchantProfileSettings(merchantId, merchantInfo),
    onSuccess: (updatedSettings) => {
      queryClient.setQueryData(
        merchantProfileSettingsQueryOptions(merchantId).queryKey,
        updatedSettings
      );
      setChanges({});
      showToast({
        message: "Merchant profile updated successfully",
        variant: "success",
      });
    },
  });

  const [logoFile, setLogoFile] = useState(null);
  const [logoPreviewUrl, setLogoPreviewUrl] = useState(null);
  const [isLogoUploading, setIsLogoUploading] = useState(false);

  const canEditLogo = role === "owner" || role === "admin";
  const logoUrl = `/api/v1/merchants/${merchantId}/media/logo`;

  function handleInputData({ name, value }) {
    setChanges((current) => ({ ...current, [name]: value }));
  }

  function selectLogo(file) {
    setLogoFile(file);
    setLogoPreviewUrl(URL.createObjectURL(file));
  }

  function closeLogoCropDialog() {
    if (logoPreviewUrl) URL.revokeObjectURL(logoPreviewUrl);
    setLogoFile(null);
    setLogoPreviewUrl(null);
  }

  async function refreshLogo() {
    // only the logo is refetched, unsaved text changes are kept
    await queryClient.invalidateQueries(
      merchantProfileSettingsQueryOptions(merchantId)
    );
  }

  async function saveLogo(crop) {
    setIsLogoUploading(true);

    try {
      await uploadImage(logoUrl, logoFile, crop);
      await refreshLogo();
      closeLogoCropDialog();
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    } finally {
      setIsLogoUploading(false);
    }
  }

  async function removeLogo() {
    try {
      await deleteImage(logoUrl);
      await refreshLogo();
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    }
  }

  if (isLoading) return <Loading />;

  return (
    <div className="flex flex-col gap-12 pb-28 md:pb-8">
      <SettingsPageHeader
        title="Merchant profile"
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

      {canEditLogo && (
        <SettingsSection
          title="Logo"
          description={
            <>
              Shown on your booking page and in link previews. Photos of your
              business are managed in the{" "}
              <Link
                from={Route.fullPath}
                to="/settings/location"
                className="text-fg-primary font-medium hover:underline"
              >
                location settings
              </Link>
              .
            </>
          }
        >
          <LogoUploader
            logoUrl={profileSettings.logo_url}
            isUploading={isLogoUploading}
            onFileSelected={selectLogo}
            onRemove={removeLogo}
          />
          <ImageCropDialog
            key={logoPreviewUrl}
            open={logoPreviewUrl !== null}
            onOpenChange={(open) =>
              !open && !isLogoUploading && closeLogoCropDialog()
            }
            imageSrc={logoPreviewUrl}
            title="Crop logo"
            cropShape="rect"
            isSaving={isLogoUploading}
            onSave={saveLogo}
          />
        </SettingsSection>
      )}

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

      {role === "owner" && <DangerZone />}
    </div>
  );
}
