import Button from "@components/Button";
import CloseButton from "@components/CloseButton";
import Input from "@components/Input";
import Modal from "@components/Modal";
import Select from "@components/Select";
import {
  convertFromBaseUnit,
  convertToBaseUnit,
  unitConversionMap,
  unitOptions,
} from "@lib/units";
import { useEffect, useState } from "react";

const defaultProductData = {
  id: null,
  name: "",
  description: "",
  price: "",
  duration: "",
  max_amount: "",
  max_amount_unit: "",
  current_amount: "",
  current_amount_unit: "",
  services: [],
};

export default function ProductModal({ data, isOpen, onClose, onSubmit }) {
  const [productData, setProductData] = useState(defaultProductData);
  const [unitError, setUnitError] = useState();

  useEffect(() => {
    if (!data) {
      setProductData(defaultProductData);
      return;
    }

    const convertedCurrent = convertFromBaseUnit(
      data.current_amount,
      data.unit
    );
    const convertedMax = convertFromBaseUnit(data.max_amount, data.unit);

    setProductData({
      ...data,
      current_amount: convertedCurrent.value,
      current_amount_unit: convertedCurrent.unit,
      max_amount: convertedMax.value,
      max_amount_unit: convertedMax.unit,
      services: data.services || [],
    });
  }, [data]);

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    //max might be optional later
    if (!productData.current_amount_unit || !productData.max_amount_unit) {
      setUnitError("Unit must be selected.");
      return;
    }

    if (
      unitConversionMap[productData.current_amount_unit].type !==
      unitConversionMap[productData.max_amount_unit].type
    ) {
      setUnitError("Selected units must be of the same measurement type.");
      return;
    }

    const currentBase = convertToBaseUnit(
      productData.current_amount,
      productData.current_amount_unit
    );

    const maxBase = convertToBaseUnit(
      productData.max_amount,
      productData.max_amount_unit
    );

    if (currentBase > maxBase) {
      setUnitError("Max amount needs to be higher than the current");
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
        unit: unitConversionMap[productData.current_amount_unit].base,
        max_amount: maxBase,
        current_amount: currentBase,
      });
    }

    setProductData(defaultProductData);
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
              <div className="flex items-end justify-center gap-1">
                <Input
                  styles="p-2"
                  id="current_amount"
                  labelText="Current Amount"
                  name="current_amount"
                  type="number"
                  placeholder="0"
                  min={0}
                  step={1}
                  value={productData.current_amount}
                  inputData={onChangeHandler}
                />
                <Select
                  options={unitOptions}
                  value={productData.current_amount_unit}
                  onSelect={(selected) => {
                    onChangeHandler({
                      name: "current_amount_unit",
                      value: selected.value,
                    });
                    setUnitError("");
                  }}
                  placeholder=""
                  styles="w-24 mb-0.5"
                />
              </div>

              <div className="flex items-end justify-center gap-1">
                <Input
                  styles="p-2"
                  id="max_amount"
                  labelText="Max Amount"
                  name="max_amount"
                  type="number"
                  placeholder="0"
                  min={0}
                  step={1}
                  value={productData.max_amount}
                  inputData={onChangeHandler}
                />
                <Select
                  options={unitOptions}
                  value={productData.max_amount_unit}
                  onSelect={(selected) => {
                    onChangeHandler({
                      name: "max_amount_unit",
                      value: selected.value,
                    });
                    setUnitError("");
                  }}
                  placeholder=""
                  styles="w-24 mb-0.5"
                />
              </div>
              {unitError && (
                <span className="my-1 text-sm text-red-500">{unitError}</span>
              )}
              {productData.services.length > 0 && (
                <div className="flex flex-col gap-2">
                  <span>Connected Services</span>
                  <div
                    className="mt-1 flex gap-2 overflow-x-auto scroll-smooth rounded-lg pb-2 outline-none
                      md:max-h-24 md:flex-wrap md:overflow-y-auto dark:[color-scheme:dark]"
                  >
                    {productData.services.map((service) => (
                      <div
                        key={service.id}
                        className="bg-hvr_gray flex max-w-44 items-center gap-2 rounded-full px-3 py-1 text-sm
                          md:max-w-36"
                      >
                        <span
                          className="h-3 w-3 shrink-0 rounded-full"
                          style={{ backgroundColor: service.color }}
                        ></span>
                        <span className="text-text_color truncate">
                          {service.name}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
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
