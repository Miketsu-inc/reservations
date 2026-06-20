import {
  Button,
  CloseButton,
  Drawer,
  DrawerContent,
  Input,
  Modal,
} from "@reservations/components";
import { useState } from "react";

export default function NewCustomerOverlay({
  onSave,
  onClose,
  isWindowSmall,
  isOpen,
}) {
  {
    return isWindowSmall ? (
      <Drawer
        open={isOpen}
        onOpenChange={(open) => {
          if (!open) onClose();
        }}
        styles="p-0!"
      >
        <DrawerContent styles="h-full" popUpStyles="">
          <NewCustomerForm onSave={onSave} isWindowSmall={isWindowSmall} />
        </DrawerContent>
      </Drawer>
    ) : (
      <Modal
        isOpen={isOpen}
        styles="p-5"
        onClose={onClose}
        disableFocusTrap={true}
      >
        <NewCustomerForm
          onSave={onSave}
          isWindowSmall={isWindowSmall}
          onClose={onClose}
        />
      </Modal>
    );
  }
}

const defaultCustomerData = {
  first_name: "",
  last_name: "",
  email: "",
  phone_number: "",
};

function NewCustomerForm({ onSave, isWindowSmall, onClose }) {
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
      className="flex h-full flex-col gap-5 sm:w-80"
    >
      <div className="flex justify-between">
        <p className="text-lg font-medium">Create Customer</p>{" "}
        {!isWindowSmall && <CloseButton onClick={onClose} />}
      </div>
      <div className="flex w-full flex-col gap-3">
        <Input
          styles="p-2"
          id="FirstName"
          name="FirstName"
          type="text"
          labelText="First Name"
          placeholder="Travis"
          value={customerData.first_name}
          inputData={(data) => updateCustomerData({ first_name: data.value })}
        />
        <Input
          styles="p-2"
          id="LastName"
          name="LastName"
          type="text"
          labelText="Last Name"
          placeholder="Scott"
          value={customerData.last_name}
          inputData={(data) => updateCustomerData({ last_name: data.value })}
        />
        <Input
          styles="p-2"
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
