import {
  Button,
  CloseButton,
  Input,
  Modal,
  Select,
  Textarea,
} from "@reservations/components";
import {
  convertFromBaseUnit,
  convertToBaseUnit,
  unitConversionMap,
  unitOptions,
} from "@reservations/lib";
import { useState } from "react";

const defaultProductData = {
  id: null,
  name: "",
  description: "",
  price: {
    number: "",
    currency: "HUF",
  },
  duration: "",
  max_amount: "",
  max_amount_unit: "",
  current_amount: "",
  current_amount_unit: "",
  services: [],
};

export default function ProductModal({ data, isOpen, onClose, onSubmit }) {
  const [productData, setProductData] = useState(
    data
      ? {
          ...data,
          current_amount: convertFromBaseUnit(data.current_amount, data.unit)
            .value,
          current_amount_unit: convertFromBaseUnit(
            data.current_amount,
            data.unit
          ).unit,
          max_amount: convertFromBaseUnit(data.max_amount, data.unit).value,
          max_amount_unit: convertFromBaseUnit(data.max_amount, data.unit).unit,
          services: data.services || [],
        }
      : defaultProductData
  );
  const [unitError, setUnitError] = useState();
  const [isSelectOpen, setIsSelectOpen] = useState(false);

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
        price: productData.price?.number ? productData.price : null,
        unit: unitConversionMap[productData.current_amount_unit].base,
        max_amount: maxBase,
        current_amount: currentBase,
      });
    }

    setProductData(defaultProductData);
    onClose();
  }

  function updateProductData(data) {
    setProductData((prev) => ({ ...prev, ...data }));
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      disableFocusTrap={true}
      suspendCloseOnClickOutside={isSelectOpen}
    >
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
              placeholder="Product name"
              value={productData.name}
              inputData={(data) => updateProductData({ name: data.value })}
            />
            <Input
              styles="p-2"
              id="price"
              labelText="Price"
              name="price"
              type="number"
              placeholder="0"
              required={false}
              min={0}
              max={10000000}
              value={productData.price?.number || ""}
              inputData={(data) =>
                updateProductData({
                  price: {
                    number: data.value,
                    currency: productData.price?.currency || "HUF",
                  },
                })
              }
            >
              <p
                className="border-input_border_color rounded-r-lg border px-4
                  py-2"
              >
                {productData.price?.currency || "HUF"}
              </p>
            </Input>
            <Textarea
              styles="p-2 md:max-h-32 md:min-h-32 max-h-20 min-h-20 text-sm"
              id="description"
              name="description"
              labelText="Description"
              required={false}
              placeholder="About this product..."
              value={productData.description}
              inputData={(data) =>
                updateProductData({ description: data.value })
              }
            />
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
                  inputData={(data) =>
                    updateProductData({ current_amount: data.value })
                  }
                >
                  <Select
                    options={unitOptions}
                    value={productData.current_amount_unit}
                    onSelect={(selected) => {
                      updateProductData({
                        current_amount_unit: selected.value,
                      });
                      setUnitError("");
                    }}
                    placeholder=""
                    styles="w-24! rounded-l-none"
                    onOpenChange={(open) => setIsSelectOpen(open)}
                  />
                </Input>
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
                  inputData={(data) =>
                    updateProductData({ max_amount: data.value })
                  }
                >
                  <Select
                    options={unitOptions}
                    value={productData.max_amount_unit}
                    onSelect={(selected) => {
                      updateProductData({ max_amount_unit: selected.value });
                      setUnitError("");
                    }}
                    placeholder=""
                    styles="w-24! rounded-l-none"
                    onOpenChange={(open) => setIsSelectOpen(open)}
                  />
                </Input>
              </div>
              {unitError && (
                <span className="my-1 text-sm text-red-500">{unitError}</span>
              )}
              {productData.services.length > 0 && (
                <div className="flex flex-col gap-2">
                  <span>Connected Services</span>
                  <div
                    className="mt-1 flex gap-2 overflow-x-auto scroll-smooth
                      rounded-lg pb-2 outline-none md:max-h-24 md:flex-wrap
                      md:overflow-y-auto dark:scheme-dark"
                  >
                    {productData.services.map((service) => (
                      <div
                        key={service.id}
                        className="bg-hvr_gray flex max-w-44 items-center gap-2
                          rounded-full px-3 py-1 text-sm md:max-w-36"
                      >
                        <span
                          className="size-3 shrink-0 rounded-full"
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
