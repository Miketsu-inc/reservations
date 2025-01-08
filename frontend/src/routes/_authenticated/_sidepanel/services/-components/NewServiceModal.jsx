import Button from "@components/Button";
import Input from "@components/Input";
import ServerError from "@components/ServerError";
import XIcon from "@icons/XIcon";
import { useClickOutside } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { useEffect, useRef, useState } from "react";
import TableColorPicker from "./TableColorPicker";

// html inputs always return strings regardless of the input type
const defaultServiceData = {
  name: "",
  description: "",
  colorPicker: "#2334b8",
  duration: "",
  price: "",
  cost: "",
};

export default function NewServiceModal({ isOpen, onClose, onSuccess }) {
  const modalRef = useRef();
  const [serverError, setServerError] = useState();
  const [serviceData, setServiceData] = useState(defaultServiceData);
  useClickOutside(modalRef, onClose);

  useEffect(() => {
    isOpen ? modalRef.current.showModal() : modalRef.current.close();
  }, [isOpen]);

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    try {
      const response = await fetch("/api/v1/merchants/services", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          name: serviceData.name,
          description: serviceData.description,
          color: serviceData.colorPicker,
          duration: parseInt(serviceData.duration),
          price: parseInt(serviceData.price),
          cost: parseInt(serviceData.cost) || 0,
        }),
      });

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        setServerError();
        setServiceData(defaultServiceData);
        onSuccess();
        onClose();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  function onChangeHandler(e) {
    let name = "";
    let value = "";

    if (e.target) {
      name = e.target.name;
      value = e.target.value;
    } else {
      name = e.name;
      value = e.value;
    }

    setServiceData((prev) => ({
      ...prev,
      [name]: value,
    }));
  }

  return (
    <dialog
      className="w-full rounded-lg bg-layer_bg text-text_color shadow-md shadow-layer_bg
        transition-all backdrop:bg-black backdrop:bg-opacity-35 md:w-fit"
      ref={modalRef}
    >
      <form id="newServiceForm" onSubmit={submitHandler} className="m-2 mx-3">
        <div className="flex flex-col">
          <div className="my-1 flex flex-row items-center justify-between">
            <p className="text-lg md:text-xl">New service</p>
            <XIcon
              styles="h-8 w-8 md:h-9 md:w-9 fill-text_color cursor-pointer"
              onClick={onClose}
            />
          </div>
          <ServerError error={serverError} />
          <hr className="py-2 md:py-3" />
        </div>
        <div className="flex flex-col gap-6 pb-1 md:flex-row md:gap-8">
          <div className="flex flex-col gap-4 md:w-72">
            <div className="flex flex-col gap-1">
              <label htmlFor="name">Name</label>
              <Input
                styles="p-2"
                id="name"
                name="name"
                type="text"
                hasError={false}
                placeholder="Serivce name"
                value={serviceData.name}
                inputData={onChangeHandler}
              />
            </div>
            <div className="flex h-8 flex-row items-center gap-6">
              <label htmlFor="colorPicker">Color</label>
              <TableColorPicker
                value={serviceData.colorPicker}
                onChange={onChangeHandler}
              />
            </div>
            <div className="flex flex-col gap-1">
              <label htmlFor="description">Description</label>
              <textarea
                id="description"
                name="description"
                placeholder="About this service..."
                className="max-h-20 min-h-20 w-full rounded-lg border border-gray-300 bg-bg_color p-2
                  text-sm outline-none focus:border-primary md:max-h-32 md:min-h-32"
                value={serviceData.description}
                onChange={onChangeHandler}
              />
            </div>
          </div>
          <div className="flex flex-col gap-4 md:w-[25rem] md:justify-between md:pt-7">
            <div className="flex flex-col gap-2 md:gap-4">
              <div className="flex flex-col gap-1 md:flex-row md:items-center md:justify-between md:gap-10">
                <label htmlFor="duration">Duration (min)</label>
                <Input
                  styles="p-2 md:w-64"
                  id="duration"
                  name="duration"
                  type="number"
                  placeholder="0"
                  min={1}
                  max={1440}
                  errorText="Duration must be between 1 and 1440"
                  value={serviceData.duration}
                  inputData={onChangeHandler}
                />
              </div>
              <div className="flex flex-col gap-1 md:flex-row md:items-center md:justify-between md:gap-10">
                <label htmlFor="price">Price (HUF)</label>
                <Input
                  styles="p-2 md:w-64"
                  id="price"
                  name="price"
                  type="number"
                  placeholder="0"
                  min={0}
                  max={10000000}
                  errorText="Price must be between 1 and 1,000,000"
                  value={serviceData.price}
                  inputData={onChangeHandler}
                />
              </div>
              <div className="flex flex-col gap-1 md:flex-row md:items-center md:justify-between md:gap-10">
                <label htmlFor="cost">Cost (HUF)</label>
                <Input
                  styles="p-2 md:w-64"
                  id="cost"
                  name="cost"
                  type="number"
                  placeholder="0"
                  min={0}
                  max={10000000}
                  required={false}
                  errorText="Cost must be between 0 and 1,000,000"
                  value={serviceData.cost}
                  inputData={onChangeHandler}
                />
              </div>
            </div>
            <div className="text-right">
              <Button
                type="submit"
                name="add service"
                styles="py-2 px-4"
                buttonText="Add"
              />
            </div>
          </div>
        </div>
      </form>
    </dialog>
  );
}
