import { UserSwitchIcon } from "@hugeicons/core-free-icons";
import {
  Button,
  ComboBox,
  Icon,
  ResponsiveDialog,
  ServerError,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  customersQueryOptions,
  invalidateLocalStorageAuth,
  useToast,
} from "@reservations/lib";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

async function transferBookings(merchantId, fromCustomerId, toCustomerId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/customers/transfer`,
    {
      method: "PUT",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({
        from_customer_id: fromCustomerId,
        to_customer_id: toCustomerId,
      }),
    }
  );

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    const result = await response.json().catch(() => null);
    throw new Error(result?.error?.message || "Could not transfer bookings");
  }
}

export default function TransferAppsModal({ fromCustomerId, isOpen, onClose }) {
  const [showValidationError, setShowValidationError] = useState(false);
  const [isComboBoxOpen, setIsComboBoxOpen] = useState(false);
  const [toCustomerId, setToCustomerId] = useState("");
  const { merchantId } = useAuth();
  const { showToast } = useToast();
  const queryClient = useQueryClient();

  const customersQuery = useQuery({
    ...customersQueryOptions(merchantId),
    enabled: isOpen && Boolean(merchantId),
  });

  const transferMutation = useMutation({
    mutationFn: () =>
      transferBookings(merchantId, fromCustomerId, toCustomerId),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries(customersQueryOptions(merchantId)),
        queryClient.invalidateQueries({
          queryKey: [merchantId, "customer-info"],
        }),
      ]);
      showToast({
        message: "Bookings transferred successfully",
        variant: "success",
      });
      handleClose();
    },
  });

  const customers = customersQuery.data || [];
  const fromCustomer = customers.find(
    (customer) => customer.id === fromCustomerId
  );

  // Dummy customers cannot receive transferred bookings.
  const filteredCustomers = customers.filter(
    (customer) => customer.id !== fromCustomerId && !customer.is_dummy
  );

  const options = filteredCustomers.map((customer) => ({
    value: customer.id,
    label: `${customer.first_name} ${customer.last_name}`,
  }));

  function handleClose() {
    setShowValidationError(false);
    setToCustomerId("");
    transferMutation.reset();
    onClose();
  }

  function submitHandler(e) {
    e.preventDefault();

    if (!toCustomerId) {
      setShowValidationError(true);
      return;
    }

    setShowValidationError(false);
    transferMutation.mutate();
  }

  return (
    <ResponsiveDialog
      isOpen={isOpen}
      onClose={handleClose}
      disableFocusTrap={true}
      suspendCloseOnClickOutside={isComboBoxOpen || transferMutation.isPending}
    >
      <form onSubmit={submitHandler} className="m-3 sm:w-md">
        <p className="pb-6 text-xl">Transfer bookings</p>
        <ServerError
          styles="mb-4"
          error={
            customersQuery.error?.message || transferMutation.error?.message
          }
        />
        <div className="flex items-center justify-center gap-6 py-2 sm:px-4">
          <p className="w-fit text-lg font-semibold sm:text-nowrap">
            {fromCustomer
              ? `${fromCustomer.first_name} ${fromCustomer.last_name}`
              : customersQuery.isError
                ? "Customer unavailable"
                : "Loading customer..."}
          </p>
          <Icon icon={UserSwitchIcon} styles="size-7" />
          <ComboBox
            options={options}
            value={toCustomerId}
            placeholder="Search customers"
            emptyText={
              customersQuery.isLoading
                ? "Loading customers..."
                : filteredCustomers.length === 0
                  ? "You have no customer to transfer to"
                  : ""
            }
            onSelect={(option) => {
              setToCustomerId(option.value);
              setShowValidationError(false);
            }}
            styles="w-fit"
            maxVisibleItems={5}
            onOpenChange={setIsComboBoxOpen}
            disabled={customersQuery.isLoading || customersQuery.isError}
          />
        </div>
        <p
          className={`${showValidationError ? "visible" : "invisible"}
            text-center text-red-500`}
        >
          Please select a customer!
        </p>
        <div className="flex justify-center py-3">
          <div className="py-4 text-center">
            <p className="text-gray-700 dark:text-gray-300">
              You are about to transfer all past and future bookings (booked
              until now) to another customer.
              <br />
              This is a permanent action which cannot be reverted!
            </p>
          </div>
        </div>
        <div className="flex flex-row items-center justify-end gap-4">
          <Button
            variant="tertiary"
            name="cancel"
            styles="py-2 px-3 hidden lg:block"
            buttonText="Cancel"
            type="button"
            onClick={handleClose}
            disabled={transferMutation.isPending}
          />
          <Button
            variant="danger"
            name="transfer"
            styles="py-2 px-3 w-full lg:w-auto"
            buttonText="Transfer"
            type="submit"
            isLoading={transferMutation.isPending}
            disabled={
              customersQuery.isLoading ||
              customersQuery.isError ||
              filteredCustomers.length === 0
            }
          />
        </div>
      </form>
    </ResponsiveDialog>
  );
}
