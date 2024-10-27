import { useState } from "react";
import Button from "../../components/Button";
import Input from "../../components/Input";

const defaultFormData = {
  appointment_type: "",
  duration: "",
  price: "",
};

export default function AppointmentForm({ sendInputData, isCompleted }) {
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
      appointment_type: formData.appointment_type,
      duration: formData.duration,
      price: formData.price,
    });
    isCompleted(true);
  }

  function handleInputData(data) {
    setFormData((prevFormData) => ({
      ...prevFormData,
      [data.name]: data.value,
    }));
  }

  return (
    <>
      <h1 className="mt-10 text-xl font-bold">Appointemnt Info</h1>
      <p className="px-6 text-center">
        Give info about an appointment, which your client will apply for
      </p>
      <form
        noValidate
        className="mt-10 flex w-full flex-col items-center justify-center px-8 *:w-full sm:px-10"
        onSubmit={handleSubmit}
      >
        <Input
          type="text"
          styles=""
          placeholder="Nail polish"
          pattern=".{0,255}"
          name="appointment_type"
          id="appointment_type"
          errorText="Inputs must be 256 character or less!"
          labelText="Appointment type"
          inputData={handleInputData}
          hasError={isEmpty}
        />
        <Input
          type="text"
          styles=""
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
          styles=""
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
          onCLick={""}
          styles="p-2 w-5/6 mt-10 hover:bg-hvr_primary"
          name=""
          type="submit"
          buttonText="Continue"
        />
      </form>
    </>
  );
}
