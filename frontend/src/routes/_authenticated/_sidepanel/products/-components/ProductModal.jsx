import Button from "@components/Button";
import CloseButton from "@components/CloseButton";
import Input from "@components/Input";
import Modal from "@components/Modal";
import { useEffect, useState } from "react";
import MultipleSelect from "./MultipleSelect";

const defaultProductData = {
  id: null,
  name: "",
  description: "",
  price: "",
  duration: "",
  stock_quantity: "",
  usage_per_unit: "",
  service_ids: [],
};

export default function ProductModal({
  data,
  isOpen,
  onClose,
  onSubmit,
  serviceData,
}) {
  const [productData, setProductData] = useState(defaultProductData);
  const [submitted, setSubmitted] = useState(false);

  useEffect(() => {
    setProductData(data || defaultProductData);
  }, [data]);

  const selectData = serviceData?.map((service) => ({
    value: service.Id,
    label: service.Name,
    icon: (
      <span
        className="h-4 w-4 shrink-0 rounded-full"
        style={{ backgroundColor: service.Color }}
      ></span>
    ),
  }));

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    let didChange = false;

    // if received data is empty and checkValidity passed
    // onSubmit can get triggered
    if (data) {
      for (var key in productData) {
        if (productData[key] !== data[key]) {
          didChange = true;
        }
      }
    } else {
      didChange = true;
    }
    if (didChange) {
      onSubmit({
        id: productData.id,
        name: productData.name,
        description: productData.description,
        price: parseInt(productData.price),
        stock_quantity: parseInt(productData.stock_quantity),
        usage_per_unit: parseInt(productData.usage_per_unit),
        service_ids: productData.service_ids.map((id) => parseInt(id)),
      });
    }

    setProductData(defaultProductData);
    setSubmitted(true);
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

    setProductData((prev) => ({
      ...prev,
      [name]: value,
    }));
  }

  function handleAddition(newService) {
    setProductData((prev) => ({
      ...prev,
      service_ids: [...prev.service_ids, newService.value],
    }));
  }

  function handleDeletion(serviceToRemove) {
    setProductData((prev) => ({
      ...prev,
      service_ids: prev.service_ids.filter(
        (id) => id !== serviceToRemove.value
      ),
    }));
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <form id="ProductForm" onSubmit={submitHandler} className="m-2 mx-3">
        <div className="flex flex-col">
          <div className="my-1 flex flex-row items-center justify-between">
            <p className="text-lg md:text-xl">Product</p>
            <CloseButton onClick={onClose} />
          </div>
          <hr className="py-2 md:py-3" />
        </div>
        <div className="flex flex-col gap-3 pb-1 md:flex-row md:gap-8">
          <div className="flex flex-col gap-4 md:w-80">
            <Input
              styles="p-2"
              labelText="Product Name"
              id="name"
              name="name"
              type="text"
              hasError={false}
              placeholder="Product name"
              value={productData.name}
              inputData={onChangeHandler}
            />
            <Input
              styles="p-2"
              id="price"
              labelText="Price (HUF)"
              name="price"
              type="number"
              placeholder="0"
              min={0}
              max={10000000}
              errorText="Price must be between 1 and 1,000,000"
              value={productData.price}
              inputData={onChangeHandler}
            />
            <div className="flex flex-col gap-1">
              <label htmlFor="description">Description</label>
              <textarea
                id="description"
                name="description"
                placeholder="About this product..."
                className="bg-bg_color focus:border-primary max-h-20 min-h-16 w-full rounded-lg border
                  border-gray-300 p-2 text-sm outline-hidden md:max-h-32 md:min-h-32"
                value={productData.description}
                onChange={onChangeHandler}
              />
            </div>
          </div>
          <div className="flex flex-col gap-4 md:w-96 md:justify-between">
            <div className="flex flex-col gap-2 md:gap-4">
              <Input
                styles="p-2 md:w-80"
                id="stock_quntity"
                labelText="In Stock Quantity"
                name="stock_quantity"
                type="number"
                placeholder="0"
                min={0}
                max={1000000}
                errorText="Qunatity must be between 1 and 1,000,000"
                value={productData.stock_quantity}
                inputData={onChangeHandler}
              />
              <Input
                styles="p-2 md:w-80"
                id="usage_per_unit"
                labelText="Usage Per Unit"
                name="usage_per_unit"
                type="number"
                placeholder="0"
                min={0}
                max={10000000}
                errorText="This field must be between 1 and 1,000,000"
                value={productData.usage_per_unit}
                inputData={onChangeHandler}
              />
              <div className="flex flex-col gap-1">
                <label className="">Connect the product to services</label>
                <MultipleSelect
                  options={selectData || []}
                  initialItems={productData.service_ids}
                  onAddition={handleAddition}
                  onDeletion={handleDeletion}
                  styles="w-full"
                  setDefault={submitted}
                />
              </div>
            </div>
            <div className="text-right">
              <Button
                variant="primary"
                type="submit"
                name="add product"
                styles="md:py-2 py-1 px-4"
                buttonText="Submit"
              />
            </div>
          </div>
        </div>
      </form>
    </Modal>
  );
}
