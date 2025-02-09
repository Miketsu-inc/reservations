import Button from "@components/Button";
import Modal from "@components/Modal";
import TransferIcon from "@icons/TransferIcon";
import { useState } from "react";

export default function TransferAppsModal({ data, isOpen, onClose, onSubmit }) {
  const [showError, setShowError] = useState(false);

  // filter out dummy customers and the one being transferred from
  const filteredCustomers = data?.customers.filter(
    (customer) =>
      customer.id !== data.customers[data.fromIndex].id &&
      customer.is_dummy === false
  );

  function submitHandler(e) {
    e.preventDefault();

    const to = e.target.toCustomer.value;

    if (to === "default") {
      setShowError("Please select a customer!");
      return;
    }

    setShowError("");
    onSubmit({
      from: data.customers[data.fromIndex].id,
      to: to,
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
    >
      <form onSubmit={submitHandler} className="m-3 md:m-4 md:w-md">
        <p className="pb-6 text-xl">Transfer appointments</p>
        <div className="flex items-center justify-center gap-2 py-2">
          <p className="text-lg font-semibold">
            {data?.customers[data.fromIndex].first_name +
              " " +
              data?.customers[data.fromIndex].last_name}
          </p>
          <TransferIcon styles="w-7 h-7" />
          <select
            id="toCustomer"
            className="bg-bg_color text-text_color font-semibold"
            defaultValue="default"
          >
            <option value="default" disabled={true} hidden={true}>
              A customer
            </option>
            {filteredCustomers?.map((customer) => (
              <option value={customer.id} key={customer.id}>
                {customer.first_name + " " + customer.last_name}
              </option>
            ))}
          </select>
        </div>
        <p
          className={`${showError ? "visible" : "invisible"} text-center text-red-500`}
        >
          Please select a customer!
        </p>
        <div className="flex justify-center py-3">
          <div className="py-4 text-center">
            <p className="text-gray-700 dark:text-gray-300">
              You are about to transfer all past and future appointments (booked
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
