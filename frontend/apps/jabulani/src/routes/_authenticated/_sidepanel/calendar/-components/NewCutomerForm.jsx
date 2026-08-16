import {
  Button,
  CloseButton,
  Input,
  ResponsiveDialog,
} from "@reservations/components";
import { useWindowSize } from "@reservations/lib";
import { useState } from "react";

export default function NewCustomerOverlay({ onSave, onClose, isOpen }) {
  const [isPhoneCountrySelectOpen, setIsPhoneCountrySelectOpen] =
    useState(false);

  return (
    <ResponsiveDialog
      isOpen={isOpen}
      onClose={onClose}
      disableFocusTrap={true}
      suspendCloseOnClickOutside={isPhoneCountrySelectOpen}
    >
      <NewCustomerForm
        onSave={onSave}
        onClose={onClose}
        setSelectOpen={setIsPhoneCountrySelectOpen}
      />
    </ResponsiveDialog>
  );
}

const defaultCustomerData = {
  first_name: "",
  last_name: "",
  email: "",
  phone_number: "",
};

function NewCustomerForm({ onSave, onClose, setSelectOpen }) {
  const { isWindowSmall } = useWindowSize();
  const [customerData, setCustomerData] = useState(defaultCustomerData);

  function updateCustomerData(data) {
    setCustomerData((prev) => ({ ...prev, ...data }));
  }

  function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const newCustomer = {
      ...customerData,
      customer_id:
        Date.now().toString(36) +
        "_" +
        Math.random().toString(36).substring(2, 9), // crypto.randomUUID() would be better, but works only on localhost or https
      isNewCustomer: true,
    };

    onSave(newCustomer);
  }

  return (
    <form
      onSubmit={submitHandler}
      className="flex h-full w-full flex-col gap-5 p-4"
    >
      <div className="flex justify-between">
        <p className="text-lg font-medium">Create Customer</p>{" "}
        {!isWindowSmall && <CloseButton onClick={onClose} />}
      </div>
      <div className="flex w-full flex-col gap-3">
        <Input
          id="FirstName"
          name="FirstName"
          type="text"
          labelText="First Name"
          placeholder="Travis"
          value={customerData.first_name}
          inputData={(data) => updateCustomerData({ first_name: data.value })}
        />
        <Input
          id="LastName"
          name="LastName"
          type="text"
          labelText="Last Name"
          placeholder="Scott"
          value={customerData.last_name}
          inputData={(data) => updateCustomerData({ last_name: data.value })}
        />
        <Input
          id="Email"
          name="Email"
          type="email"
          labelText="Email"
          placeholder="example@gmail.com"
          required={false}
          value={customerData.email}
          inputData={(data) => updateCustomerData({ email: data.value })}
        />
        <Input
          id="PhoneNumber"
          name="PhoneNumber"
          type="tel"
          labelText="Phone Number"
          placeholder="+36 20 678 2012"
          required={false}
          value={customerData.phone_number}
          inputData={(data) => updateCustomerData({ phone_number: data.value })}
          onOpenChange={(open) => setSelectOpen(open)}
        />
      </div>
      <Button
        styles="py-2 px-4 mt-2"
        variant="primary"
        type="submit"
        buttonText="Add Customer"
      />
    </form>
  );
}
