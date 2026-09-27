import { Button, Card, DeleteModal } from "@reservations/components";
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
    <section>
      <Card styles="h-auto! border-red-300! p-6! dark:border-red-900!">
        <h2 className="text-xl font-semibold text-red-600 dark:text-red-400">
          Danger zone
        </h2>
        <p className="text-text_color/65 mt-1 mb-6 text-sm">
          These actions affect your business identity or permanently remove it.
        </p>
        <div className="flex flex-col gap-5">
        <DangerZoneItem
          title="Change business name"
          description="Changing the name also changes the URL of your booking page."
          action={
            <MerchantNameModal
              onSubmit={handleNameChange}
              trigger={
                <Button
                  variant="danger"
                  styles="py-2 px-3 w-fit text-nowrap"
                  buttonText="Change name"
                />
              }
            />
          }
        />
        <div className="border-border_color border-t" />
        <DangerZoneItem
          title="Business visibility"
          description="Make this business private or public."
          buttonText="Change visibility"
        />
        <div className="border-border_color border-t" />
        <DangerZoneItem
          title="Ownership transfer"
          description="Transfer this business to another account."
          buttonText="Transfer ownership"
        />
        <div className="border-border_color border-t" />
        <DangerZoneItem
          title="Delete business"
          description="This permanently deletes the business and cannot be undone."
          action={
            <DeleteModal
              onDelete={deletehandler}
              itemName="this business"
              trigger={
                <Button
                  variant="danger"
                  styles="py-2 px-3 w-fit text-nowrap"
                  buttonText="Delete business"
                />
              }
            />
          }
        />
        </div>
      </Card>
    </section>
  );
}
