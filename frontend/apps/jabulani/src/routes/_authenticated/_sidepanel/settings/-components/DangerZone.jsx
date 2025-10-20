import { DeleteModal } from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { useRouter } from "@tanstack/react-router";
import { useState } from "react";
import DangerZoneItem from "./DangerZoneItem";
import MerchantNameModal from "./MerchantNameModal";
import SectionHeader from "./SectionHeader";

export default function DangerZone() {
  const router = useRouter();
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isMerchantNameModalOpen, setisMerchantNameModalOpen] = useState(false);
  const { showToast } = useToast();

  async function deletehandler() {
    const response = await fetch("/api/v1/merchants", {
      method: "DELETE",
    });
    if (!response.ok) {
      invalidateLocalStorageAuth(response.status);
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      router.navigate({
        from: "/settings/merchant",
        to: "/login",
      });
      showToast({
        message: "Merchant deleted successfully",
        variant: "success",
      });
    }
  }

  async function handleNameChange(newName) {
    const response = await fetch("/api/v1/merchants/name", {
      method: "PATCH",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({ name: newName }),
    });

    if (!response.ok) {
      invalidateLocalStorageAuth(response.status);
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      showToast({
        message: "Name changed successfully",
        variant: "success",
      });
    }
  }

  return (
    <>
      <MerchantNameModal
        isOpen={isMerchantNameModalOpen}
        onClose={() => setisMerchantNameModalOpen(false)}
        onSubmit={(name) => handleNameChange(name)}
      />
      <DeleteModal
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        onDelete={deletehandler}
        itemName="this merchant"
      />
      <div className="flex flex-col gap-4">
        <SectionHeader styles="text-red-600" title="Danger zone" />
        <DangerZoneItem
          title="Change Merchant Name"
          description="By changing the name the URL of your page will change as well."
          buttonText="Change name"
          onClick={() => setisMerchantNameModalOpen(true)}
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
          onClick={() => setIsDeleteModalOpen(true)}
        />
      </div>
    </>
  );
}
