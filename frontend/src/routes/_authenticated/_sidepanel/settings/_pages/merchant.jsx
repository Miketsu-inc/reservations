import Button from "@components/Button";
import ServerError from "@components/ServerError";
import Textarea from "@components/Textarea";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { preferencesQueryOptions } from "@lib/queries";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import BusinessHours from "../-components/BusinessHours";
import DangerZoneItem from "../-components/DangerZoneItem";
import ImageUploader from "../-components/ImageUploader";
import SectionHeader from "../-components/SectionHeader";

const daysOfWeek = {
  1: "Monday",
  2: "Tuesday",
  3: "Wednesday",
  4: "Thursday",
  5: "Friday",
  6: "Saturday",
  0: "Sunday",
};

async function fetchMerchantData() {
  const response = await fetch(`/api/v1/merchants/settings-info`, {
    method: "GET",
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function hasChanges(changedData, originalData) {
  return JSON.stringify(changedData) !== JSON.stringify(originalData);
}

function validateBusinessHours(hours) {
  for (const day in hours) {
    const periods = hours[day];

    if (periods.length === 0) continue;

    // Sort time periods by start time
    const sortedPeriods = [...periods].sort((a, b) =>
      a.start_time.localeCompare(b.start_time)
    );

    for (let i = 0; i < sortedPeriods.length; i++) {
      const { start_time, end_time } = sortedPeriods[i];

      if (start_time >= end_time) {
        return `Invalid time range on ${daysOfWeek[day]}. Start time must be before end time.`;
      }

      if (i > 0 && sortedPeriods[i - 1].end_time > start_time) {
        return `Overlapping hours on ${daysOfWeek[day]}. PLease make sure to correct it`;
      }
    }
  }
  return "";
}

const defaultMerchantInfo = {
  merchant_name: "",
  location_id: 0,
  contact_email: "",
  short_location: "",
  introduction: "",
  announcement: "",
  about_us: "",
  parking_info: "",
  payment_info: "",
  business_hours: { 1: [], 2: [], 3: [], 4: [], 5: [], 6: [], 0: [] },
};

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_pages/merchant"
)({
  component: MerchantPage,
  loader: async () => {
    const merchantData = await fetchMerchantData();

    return {
      ...merchantData,
    };
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function MerchantPage() {
  const loaderData = Route.useLoaderData();
  const initialMerchantInfo = loaderData
    ? {
        merchant_name: loaderData.merchant_name,
        location_id: loaderData.location_id,
        contact_email: loaderData.contact_email,
        short_location: `${loaderData.address}, ${loaderData.city} ${loaderData.postal_code}`,
        introduction: loaderData.introduction,
        announcement: loaderData.announcement,
        about_us: loaderData.about_us,
        parking_info: loaderData.parking_info,
        payment_info: loaderData.payment_info,
        business_hours: loaderData.business_hours,
      }
    : defaultMerchantInfo;

  const [merchantInfo, setMerchantInfo] = useState(initialMerchantInfo);
  const [originalData, setOriginalData] = useState(initialMerchantInfo);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [serverError, setServerError] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const { showToast } = useToast();

  const { data: preferences } = useQuery(preferencesQueryOptions());

  function handleInputData(data) {
    setMerchantInfo((prevFormData) => ({
      ...prevFormData,
      [data.name]: data.value,
    }));
  }

  useEffect(() => {
    const validationError = validateBusinessHours(merchantInfo.business_hours);
    setErrorMessage(validationError);
  }, [merchantInfo.business_hours]);

  // needed because inside the handleInputData it couldn't detect copy pasting
  useEffect(() => {
    setHasUnsavedChanges(hasChanges(merchantInfo, originalData));
  }, [merchantInfo, originalData]);

  async function updateButtonHandler() {
    if (errorMessage || !hasUnsavedChanges) {
      return;
    }

    try {
      const response = await fetch("/api/v1/merchants/reservation-fields", {
        method: "PATCH",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          introduction: merchantInfo.introduction,
          announcement: merchantInfo.announcement,
          about_us: merchantInfo.about_us,
          payment_info: merchantInfo.payment_info,
          parking_info: merchantInfo.parking_info,
          business_hours: merchantInfo.business_hours,
        }),
      });

      if (!response.ok) {
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        setOriginalData(merchantInfo);
        setServerError("");
        showToast({
          message: "Merchant updated successfully!",
          variant: "success",
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  const handleImageUpload = (file) => {
    // Process the uploaded file
    console.log(file);
  };

  return (
    <div className="flex w-full flex-col gap-6">
      <div className="flex w-full flex-col gap-6">
        <SectionHeader title="General info" styles="" />
        <Textarea
          styles="p-2 max-h-96 min-h-28 md:w-2/3 min-w-min md:min-w-52
            md:max-w-2xl"
          id="intorudtion"
          placeholder="Introduce your company to the clients"
          name="introduction"
          required={false}
          labelText="Introduction"
          value={merchantInfo.introduction}
          inputData={handleInputData}
        />
        <Textarea
          styles="p-2 max-h-96 min-h-28 md:w-2/3 min-w-52 md:max-w-2xl"
          id="announcement"
          placeholder=""
          name="announcement"
          required={false}
          labelText="Announcement"
          value={merchantInfo.announcement}
          inputData={handleInputData}
        />
        <Textarea
          styles="p-2 max-h-24 min-h-16 md:w-2/3 min-w-52 md:max-w-2xl"
          id="payment_info"
          placeholder=""
          required={false}
          name="payment_info"
          labelText="Payment Info"
          value={merchantInfo.payment_info}
          inputData={handleInputData}
        />
        <Textarea
          styles="p-2 max-h-24 min-h-16 md:w-2/3 min-w-52 md:max-w-2xl"
          id="parking_info"
          placeholder=""
          required={false}
          name="parking_info"
          labelText="Parking Info"
          value={merchantInfo.parking_info}
          inputData={handleInputData}
        />
        <Textarea
          styles="p-2 max-h-96 min-h-28 md:w-2/3 min-w-min md:min-w-52
            md:max-w-2xl"
          id="about_us"
          placeholder="Tell about your company to the clients"
          name="about_us"
          required={false}
          labelText="About Us"
          value={merchantInfo.about_us}
          inputData={handleInputData}
        />

        <div className="my-2 font-semibold">Currency</div>
        <div className="my-2 flex flex-col gap-3">
          <div className="font-semibold">Business hours</div>
          <p className="text-text_color/70">
            Set your business hours to let your customers know when you're
            available.
          </p>
          <BusinessHours
            data={merchantInfo.business_hours}
            setBusinessHours={(updater) => {
              const updatedHours = updater(merchantInfo.business_hours);
              setMerchantInfo((prev) => ({
                ...prev,
                business_hours: updatedHours,
              }));
            }}
            preferences={preferences}
          />
          {errorMessage && <span className="text-red-500">{errorMessage}</span>}
        </div>

        <div className="flex flex-col gap-3">
          <span className="text-text_color/70 text-sm md:w-2/3">
            All of the fields on this page are optional and can be deleted at
            any time, and by filling them out, you're giving us consent to share
            this data wherever your user profile appears.
          </span>
          <Button
            styles="w-min px-2 text-nowrap py-1"
            variant="primary"
            buttonText="Update fields"
            type="button"
            onClick={updateButtonHandler}
            disabled={!hasUnsavedChanges}
          />
          <ServerError error={serverError} styles="mt-2" />
        </div>
      </div>
      <div className="flex flex-col gap-3">
        <SectionHeader title="Images" styles="" />
        <p className="text-text_color/70">
          Upload images to enhance your reservation details. Your profile
          picture will be displayed on your account while additional images can
          be used to showcase the relevan visuals for your booking. Make sure to
          upload clear and appropiate images to provide the best experiance for
          your customers.
        </p>
        <div
          className="mt-6 flex flex-col items-center justify-between gap-10
            md:flex-row"
        >
          <ImageUploader
            onImageUpload={handleImageUpload}
            text="Upload profile picture"
            styles="rounded-3xl h-48 w-48"
            imageStyles="object-fill overflow-hidden rounded-3xl"
          />

          <div className="flex h-full w-full items-center justify-center gap-4">
            <ImageUploader
              onImageUpload={handleImageUpload}
              text="Upload image 1"
              styles="rounded-lg"
              imageStyles="p-2"
            />
            <ImageUploader
              onImageUpload={handleImageUpload}
              text="Upload image 2"
              styles="rounded-lg"
              imageStyles="p-2"
            />
          </div>
        </div>
      </div>
      <div className="flex flex-col gap-4">
        <SectionHeader title="Change location" styles="" />
        <span>Current location: {merchantInfo.short_location}</span>
        <Button
          variant="tertiary"
          styles="w-min text-nowrap px-2 py-1"
          buttonText="Change location"
        />
      </div>
      <div className="flex flex-col gap-4">
        <SectionHeader styles="text-red-600" title="Danger zone" />
        <DangerZoneItem
          title="Change Merchant Name"
          description="By changing the name the URL of your page will change as well."
          buttonText="Change name"
        />
        <DangerZoneItem
          title="Change Visibility"
          description="Make this merchant private or public."
          buttonText="Change visibility"
        />
        <DangerZoneItem
          title="Transfer Ownership"
          description="Transfer this merchant to another account."
          buttonText="Transfer ownership"
        />

        <DangerZoneItem
          title="Delete Merchant"
          description="Once you delete your Merchant there is no going back! Please be certain."
          buttonText="Delete your merchant"
        />
      </div>
    </div>
  );
}
