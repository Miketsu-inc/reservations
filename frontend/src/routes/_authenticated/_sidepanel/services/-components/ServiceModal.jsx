import Button from "@components/Button";
import Input from "@components/Input";
import Modal from "@components/Modal";
import XIcon from "@icons/XIcon";
import { useEffect, useState } from "react";

// html inputs always return strings regardless of the input type
// null id indicates that this service shall be added as new
const defaultServiceData = {
  id: null,
  name: "",
  description: "",
  color: "#2334b8",
  duration: "",
  price: "",
  cost: "",
};

export default function ServiceModal({ data, isOpen, onClose, onSubmit }) {
  const [serviceData, setServiceData] = useState(defaultServiceData);

  useEffect(() => {
    setServiceData(data || defaultServiceData);
  }, [data]);

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    let didChange = false;
    for (var key in serviceData) {
      if (serviceData[key] !== data[key]) {
        didChange = true;
      }
    }

    if (didChange) {
      onSubmit({
        id: serviceData.id,
        name: serviceData.name,
        description: serviceData.description,
        color: serviceData.color,
        duration: parseInt(serviceData.duration),
        price: parseInt(serviceData.price),
        cost: parseInt(serviceData.cost) || 0,
      });
    } else {
      onSubmit();
    }

    onClose();
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
    <Modal isOpen={isOpen} onClose={onClose}>
      <form id="newServiceForm" onSubmit={submitHandler} className="m-2 mx-3">
        <div className="flex flex-col">
          <div className="my-1 flex flex-row items-center justify-between">
            <p className="text-lg md:text-xl">New service</p>
            <XIcon
              styles="h-8 w-8 md:h-9 md:w-9 fill-text_color cursor-pointer"
              onClick={onClose}
            />
          </div>
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
              <label htmlFor="color">Color</label>
              <input
                id="color"
                className="h-full cursor-pointer bg-transparent"
                name="color"
                type="color"
                value={serviceData.color}
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
                buttonText="Submit"
              />
            </div>
          </div>
        </div>
      </form>
    </Modal>
  );
}
