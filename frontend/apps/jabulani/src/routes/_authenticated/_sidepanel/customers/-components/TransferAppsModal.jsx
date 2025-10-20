import { TransferIcon } from "@reservations/assets";
import { Button, ComboBox, Modal } from "@reservations/components";
import { useState } from "react";

export default function TransferAppsModal({ data, isOpen, onClose, onSubmit }) {
  const [showError, setShowError] = useState(false);
  const [isCOmboBoxOpen, setIsComboBoxOpen] = useState(false);
  const [toCustomer, setToCustomer] = useState("");
  const fromCustomer = data?.customers.find((c) => c.id === data.from);

  // filter out dummy customers and the one being transferred from
  const filteredCustomers = data?.customers.filter(
    (customer) => customer.id !== data.from && customer.is_dummy === false
  );

  const options = filteredCustomers?.map((customer) => ({
    value: customer.id,
    label: `${customer.first_name} ${customer.last_name}`,
  }));

  function submitHandler(e) {
    e.preventDefault();

    if (!toCustomer) {
      setShowError("Please select a customer!");
      return;
    }

    setShowError("");
    onSubmit({
      from: data.from,
      to: toCustomer,
    });
    onClose();
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => {
        setShowError("");
        onClose();
      }}
      disableFocusTrap={true}
      suspendCloseOnClickOutside={isCOmboBoxOpen}
    >
      <form onSubmit={submitHandler} className="m-3 sm:w-md">
        <p className="pb-6 text-xl">Transfer bookings</p>
        <div className="flex items-center justify-center gap-6 py-2 sm:px-4">
          <p className="w-fit text-lg font-semibold sm:text-nowrap">
            {fromCustomer?.first_name + " " + fromCustomer?.last_name}
          </p>
          <TransferIcon styles="w-7 h-7" />
          <ComboBox
            options={options}
            value={toCustomer}
            placeholder="Search customers"
            emptyText={
              filteredCustomers?.length === 0
                ? "You have no customer to transfer to"
                : ""
            }
            onSelect={(option) => setToCustomer(option.value)}
            styles="w-fit"
            maxVisibleItems={5}
            onOpenChange={(open) => setIsComboBoxOpen(open)}
          />
        </div>
        <p
          className={`${showError ? "visible" : "invisible"} text-center
            text-red-500`}
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
            styles="py-2 px-3"
            buttonText="Cancel"
            type="button"
            onClick={() => {
              setShowError("");
              setToCustomer("");
              onClose();
            }}
          />
          <Button
            variant="danger"
            name="transfer"
            styles="py-2 px-3"
            buttonText="Transfer"
            type="submit"
          />
        </div>
      </form>
    </Modal>
  );
}
