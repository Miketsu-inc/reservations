import Button from "@components/Button";
import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import DangerZoneItem from "../-components/DangerZoneItem";
import ImageUploader from "../-components/ImageUploader";
import SectionHeader from "../-components/SectionHeader";
import TextArea from "../-components/TextArea";

async function fetchMerchantData() {
  const response = await fetch(`/api/v1/merchants/settings-info`, {
    method: "GET",
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalSotrageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function hasChanges(changedData, originalData) {
  return JSON.stringify(changedData) !== JSON.stringify(originalData);
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
};

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_pages/merchant"
)({
  component: MerchantPage,
  loader: async () => {
    const merchantData = await fetchMerchantData();

    return {
      crumb: "Merchant",
      ...merchantData,
    };
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function MerchantPage() {
  const [merchantInfo, setMerchantInfo] = useState(defaultMerchantInfo);
  const [originalData, setOriginalData] = useState(defaultMerchantInfo);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [serverError, setServerError] = useState("");
  const loaderData = Route.useLoaderData();
  const { showToast } = useToast();

  useEffect(() => {
    if (loaderData) {
      const shortLocation =
        loaderData.address +
        ", " +
        loaderData.city +
        " " +
        loaderData.postal_code;

      const merchantData = {
        merchant_name: loaderData.merchant_name,
        location_id: loaderData.location_id,
        contact_email: loaderData.contact_email,
        short_location: shortLocation,
        introduction: loaderData.introduction,
        announcement: loaderData.announcement,
        about_us: loaderData.about_us,
        parking_info: loaderData.parking_info,
        payment_info: loaderData.payment_info,
      };
      setMerchantInfo(merchantData);
      setOriginalData(merchantData);
    }
  }, [loaderData]);

  function handleInputData(data) {
    setMerchantInfo((prevFormData) => ({
      ...prevFormData,
      [data.name]: data.value,
    }));
  }

  // needed because inside the handleInputData it couldn!t detect copy pasting
  useEffect(() => {
    setHasUnsavedChanges(hasChanges(merchantInfo, originalData));
  }, [merchantInfo, originalData]);

  async function updateButtonHandler() {
    if (!hasUnsavedChanges) return;

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
        <TextArea
          styles="max-h-96 min-h-28 md:w-2/3 min-w-min md:min-w-52 md:max-w-2xl text-text_color/90"
          id="intorudtion"
          placeholder="Introduce your company to the clients"
          name="introduction"
          description="The introduction will appear on the top of the booking page"
          label="Introduction"
          value={merchantInfo.introduction}
          sendInputData={handleInputData}
        />
        <TextArea
          styles="text-text_color/90 max-h-96 min-h-28 md:w-2/3 min-w-52 md:max-w-2xl"
          id="announcement"
          placeholder=""
          name="announcement"
          description="I dont even know what the hell is that to be honest but here i guess
            we will explain to you."
          label="Announcement"
          value={merchantInfo.announcement}
          sendInputData={handleInputData}
        />
        <TextArea
          styles="text-text_color/90 max-h-24 min-h-16 md:w-2/3 min-w-52 md:max-w-2xl resize-none"
          id="payment_info"
          placeholder=""
          name="payment_info"
          description="Tell your clients what they can use to pay you."
          label="Payment Info"
          value={merchantInfo.payment_info}
          sendInputData={handleInputData}
        />
        <TextArea
          styles="text-text_color/90 max-h-24 min-h-16 md:w-2/3 min-w-52 md:max-w-2xl resize-none"
          id="parking_info"
          placeholder=""
          name="parking_info"
          description="Share your client if there is any parking opportunity near your office"
          label="Parking Info"
          value={merchantInfo.parking_info}
          sendInputData={handleInputData}
        />
        <TextArea
          styles="max-h-96 min-h-28 md:w-2/3 min-w-min md:min-w-52 md:max-w-2xl text-text_color/90"
          id="about_us"
          placeholder="Tell about your company to the clients"
          name="about_us"
          description=""
          label="About Us"
          value={merchantInfo.about_us}
          sendInputData={handleInputData}
        />

        <div className="my-2 font-semibold">Currency (I have no idea yet)</div>
        <div className="my-2 font-semibold">Working hours</div>

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
        <div className="mt-6 flex flex-col items-center justify-between gap-10 md:flex-row">
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
