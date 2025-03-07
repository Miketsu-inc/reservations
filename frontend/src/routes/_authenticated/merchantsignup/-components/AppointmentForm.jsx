import Button from "@components/Button";
import Input from "@components/Input";
import { useState } from "react";

const defaultFormData = {
  name: "",
  duration: "",
  price: "",
};

export default function AppointmentForm({ sendInputData }) {
  const [formData, setFormData] = useState(defaultFormData);
  const [isEmpty, setIsEmpty] = useState(false);

  function handleSubmit(e) {
    e.preventDefault();
    const form = e.target;
    if (!form.checkValidity()) {
      setIsEmpty(true);
      return;
    }

    sendInputData({
      name: formData.name,
      duration: formData.duration,
      price: formData.price,
    });
  }

  function handleInputData(data) {
    setIsEmpty(false);
    setFormData((prevFormData) => ({
      ...prevFormData,
      [data.name]: data.value,
    }));
  }

  return (
    <>
      <form
        noValidate
        className="mt-4 flex w-full flex-col items-center justify-center gap-4 *:w-full"
        onSubmit={handleSubmit}
      >
        <Input
          type="text"
          styles="p-2"
          placeholder="Nail polish"
          pattern=".{0,255}"
          name="name"
          id="appointment_type"
          errorText="Inputs must be 256 character or less!"
          labelText="Appointment type"
          inputData={handleInputData}
          hasError={isEmpty}
        />
        <Input
          type="text"
          styles="p-2"
          placeholder="25"
          pattern="^[0-9]{0,255}$"
          name="duration"
          id="duration"
          errorText="The input should be numbers and less than 256 characters!"
          labelText="Duration (minutes)"
          inputData={handleInputData}
          hasError={isEmpty}
        />
        <Input
          type="text"
          styles="p-2"
          placeholder="3300"
          pattern="^[0-9]{0,255}$"
          name="price"
          id="price"
          errorText="Price should be only numbers!"
          labelText="Price"
          inputData={handleInputData}
          hasError={isEmpty}
        />

        <Button
          styles="p-2 w-5/6 mt-10 font-semibold focus-visible:outline-1 bg-primary
            hover:bg-hvr_primary text-white"
          name="add service"
          type="submit"
          buttonText="Add Service"
        />
      </form>
    </>
  );
}
