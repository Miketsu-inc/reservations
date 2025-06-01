import Button from "@components/Button";
import CloseButton from "@components/CloseButton";
import Input from "@components/Input";
import Modal from "@components/Modal";
import { useEffect, useState } from "react";

// html inputs always return strings regardless of the input type
// null id indicates that this customer shall be added as new
const defaultCustomerData = {
  id: null,
  first_name: "",
  last_name: "",
  email: "",
  phone_number: "",
};

export default function CustomerModal({ data, isOpen, onClose, onSubmit }) {
  const [customerData, setCustomerData] = useState(defaultCustomerData);

  useEffect(() => {
    setCustomerData(data || defaultCustomerData);
  }, [data]);

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    let didChange = false;

    // if received data is empty and checkValidity passed
    // onSubmit can get triggered
    if (data) {
      for (var key in customerData) {
        if (customerData[key] !== data[key]) {
          didChange = true;
        }
      }
    } else {
      didChange = true;
    }

    if (didChange) {
      onSubmit({
        id: customerData.id,
        first_name: customerData.first_name,
        last_name: customerData.last_name,
        email: customerData.email,
        phone_number: customerData.phone_number,
      });
    }

    setCustomerData(defaultCustomerData);
    onClose();
  }

  function onChangeHandler(e) {
    let name = e.name;
    let value = e.value;

    setCustomerData((prev) => ({
      ...prev,
      [name]: value,
    }));
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <form
        className="m-2 mx-3 sm:w-72"
        id="CustomerForm"
        onSubmit={submitHandler}
      >
        <div className="flex flex-col">
          <div className="my-1 flex flex-row items-center justify-between">
            <p className="text-xl">Customer</p>
            <CloseButton onClick={onClose} />
          </div>
          <hr className="py-3" />
        </div>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-1">
              <label htmlFor="first_name">First name</label>
              <Input
                styles="p-2"
                id="first_name"
                name="first_name"
                type="text"
                hasError={false}
                placeholder="First name"
                value={customerData.first_name}
                inputData={onChangeHandler}
              />
            </div>
            <div className="flex flex-col gap-1">
              <label htmlFor="last_name">Last name</label>
              <Input
                styles="p-2"
                id="last_name"
                name="last_name"
                type="text"
                hasError={false}
                placeholder="Last name"
                value={customerData.last_name}
                inputData={onChangeHandler}
              />
            </div>
            <div className="flex flex-col gap-1">
              <label htmlFor="email">Email</label>
              <Input
                styles="p-2"
                id="email"
                name="email"
                type="email"
                hasError={false}
                placeholder="customer@gmail.com"
                required={false}
                value={customerData.email}
                inputData={onChangeHandler}
              />
            </div>
            <div className="flex flex-col gap-1">
              <label htmlFor="phone_number">Phone number</label>
              <Input
                styles="p-2"
                id="phone_number"
                name="phone_number"
                type="tel"
                hasError={false}
                placeholder="+36201234567"
                required={false}
                value={customerData.phone_number}
                inputData={onChangeHandler}
              />
            </div>
          </div>
          <div className="text-right">
            <Button
              variant="primary"
              type="submit"
              name="add service"
              styles="py-2 px-4"
              buttonText="Submit"
            />
          </div>
        </div>
      </form>
    </Modal>
  );
}
