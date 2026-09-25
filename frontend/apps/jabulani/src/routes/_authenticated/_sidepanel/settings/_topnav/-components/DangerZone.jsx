import { Button, DeleteModal } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  invalidateLocalStorageAuth,
  meQueryOptions,
  useToast,
} from "@reservations/lib";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import DangerZoneItem from "./DangerZoneItem";
import MerchantNameModal from "./MerchantNameModal";
import SectionHeader from "./SectionHeader";

export default function DangerZone() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { showToast } = useToast();
  const { merchantId } = useAuth();

  async function deletehandler() {
    try {
      const response = await fetch(`/api/v1/merchants/${merchantId}`, {
        method: "DELETE",
      });
      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
        return;
      }

      localStorage.removeItem("activeMerchantId");
      await queryClient.invalidateQueries(meQueryOptions());
      showToast({
        message: "Merchant deleted successfully",
        variant: "success",
      });
      router.navigate({ to: "/" });
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    }
  }

  async function handleNameChange(newName) {
    try {
      const response = await fetch(`/api/v1/merchants/${merchantId}/name`, {
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
        return false;
      }

      await queryClient.invalidateQueries(meQueryOptions());
      showToast({
        message: "Name changed successfully",
        variant: "success",
      });
      return true;
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
      return false;
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <SectionHeader styles="text-red-600" title="Danger zone" />
      <DangerZoneItem
        title="Change Merchant Name"
        description="By changing the name the URL of your page will change as well."
        action={
          <MerchantNameModal
            onSubmit={handleNameChange}
            trigger={
              <Button
                variant="danger"
                styles="py-1 px-2 w-fit"
                buttonText="Change name"
              />
            }
          />
        }
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
        action={
          <DeleteModal
            onDelete={deletehandler}
            itemName="this merchant"
            trigger={
              <Button
                variant="danger"
                styles="py-1 px-2 w-fit"
                buttonText="Delete your merchant"
              />
            }
          />
        }
      />
    </div>
  );
}
