import Button from "@components/Button";
import Card from "@components/Card";
import ComboBox from "@components/ComboBox";
import Input from "@components/Input";
import BackArrowIcon from "@icons/BackArrowIcon";
import EditIcon from "@icons/EditIcon";
import PlusIcon from "@icons/PlusIcon";
import ProductIcon from "@icons/ProductIcon";
import TrashBinIcon from "@icons/TrashBinIcon";
import { useWindowSize } from "@lib/hooks";
import { useEffect, useMemo, useState } from "react";

export default function ProductAdder({
  availableProducts = [],
  usedProducts = [],
  onUpdate,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [editProduct, setEditProduct] = useState(null);

  const filteredAvailableProducts = availableProducts.filter(
    (p) =>
      !usedProducts.find((up) => up.id === p.id) ||
      (editProduct && p.id === editProduct.id)
  );

  function handleAddProduct(product) {
    const selectedProduct = filteredAvailableProducts.find(
      (p) => p.id === product.id
    );

    const enrichedProduct = {
      id: product.id,
      amount_used: parseInt(product.amount_used),
      name: selectedProduct.name,
      unit: selectedProduct.unit,
    };

    const exists = usedProducts.some((p) => p.id === product.id);

    const updated = exists
      ? usedProducts.map((p) => (p.id === product.id ? enrichedProduct : p))
      : [...usedProducts, enrichedProduct];

    onUpdate(updated);
    setEditProduct(null);
  }

  function handleRemove(id) {
    const updated = usedProducts.filter((p) => p.id !== id);
    onUpdate(updated);
  }

  return (
    <Card styles="!p-0 flex flex-col">
      <div
        role="button"
        onClick={() => setIsOpen(!isOpen)}
        className={`${isOpen ? "border-border_color border-b" : ""} flex cursor-pointer items-center
          justify-between p-4`}
      >
        <div className="flex items-center justify-center gap-2">
          <ProductIcon styles="size-6 mb-0.5 text-text_color" />
          <p className="text-lg">Products</p>
        </div>
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="hover:bg-hvr_gray cursor-pointer rounded-lg p-2"
        >
          <BackArrowIcon
            styles={`size-6 stroke-text_color transition-transform duration-200 ${
              isOpen ? "rotate-90" : "-rotate-90" }`}
          />
        </button>
      </div>
      {/* TODO: same issue as with dropdowns in the recurSection */}
      <div
        className={`px-4 transition-[max-height,opacity] duration-200 ease-in-out ${
          isOpen
            ? "max-h-[1000px] py-4 opacity-100"
            : "max-h-0 overflow-hidden opacity-0"
          }`}
      >
        <div className="flex flex-col gap-5 xl:flex-row xl:gap-10">
          <ProductForm
            product={editProduct || {}}
            onSubmit={handleAddProduct}
            availableProducts={filteredAvailableProducts}
            usedProducts={usedProducts}
            onSelectNewProduct={(p) => {
              if (p.id !== editProduct?.id) {
                setEditProduct(null);
              }
            }}
          />
          {usedProducts.length > 0 ? (
            <div className="flex flex-col gap-2 xl:w-1/2">
              <p className="font-medium">Connected Products</p>
              <div className="flex flex-col gap-2 overflow-y-auto xl:pr-2 dark:[color-scheme:dark]">
                {usedProducts.map((product) => {
                  return (
                    <div
                      key={product.id}
                      className="border-border_color flex flex-row items-center justify-center gap-4 rounded-md
                        border px-4 py-2 dark:border-gray-600"
                    >
                      <span className="text-text_color flex-1 font-medium">
                        {product?.name}
                      </span>

                      <div className="mr-6 flex gap-3 text-gray-500">
                        <span>{product.amount_used}</span>
                        <span>{product?.unit}</span>
                      </div>
                      <EditIcon
                        onClick={() => {
                          setEditProduct({
                            id: product.id,
                            unit: product.unit,
                            amount_used: product.amount_used,
                          });
                        }}
                        styles="size-4 cursor-pointer"
                      />
                      <TrashBinIcon
                        onClick={() => handleRemove(product.id)}
                        styles="size-5 cursor-pointer"
                      />
                    </div>
                  );
                })}
              </div>
            </div>
          ) : (
            <div className="mb-4 flex flex-col items-center justify-center xl:w-1/2">
              <div className="mb-4 flex w-min items-center justify-center rounded-full">
                <ProductIcon styles="size-12 dark:text-gray-500 text-gray-400" />
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                No products added yet
              </p>
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}

function ProductForm({
  product,
  onSubmit,
  onSelectNewProduct,
  availableProducts,
  usedProducts,
}) {
  const [productData, setProductData] = useState({
    id: 0,
    unit: "",
    amount_used: "",
  });
  const windowSize = useWindowSize();
  const isWindowSmall = ["sm", "md", "lg"].includes(windowSize);
  const productOptions = useMemo(
    () =>
      availableProducts.map((product) => ({
        value: product.id,
        label: product.name,
      })),
    [availableProducts]
  );

  const isEdit = usedProducts.some((p) => p.id === productData.id);

  useEffect(() => {
    if (product?.id) {
      setProductData({
        id: product.id,
        unit: product.unit,
        amount_used: product.amount_used,
      });
    }
  }, [product]);

  function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    onSubmit({
      id: productData.id,
      amount_used: productData.amount_used,
    });

    setProductData({
      id: 0,
      unit: "",
      amount_used: "",
    });
  }

  return (
    <form
      onSubmit={submitHandler}
      className="flex flex-col items-center gap-4 rounded-md xl:w-1/2"
    >
      <div className="w-full">
        {!isWindowSmall && (
          <label className="mb-2 block font-medium">Product</label>
        )}
        <ComboBox
          styles="w-full"
          placeholder="Select a product to add"
          value={productData.id}
          options={productOptions}
          emptyText={
            availableProducts.length === 0 ? "You have no product to add" : ""
          }
          onSelect={(option) => {
            const selected = availableProducts.find(
              (p) => p.id === option.value
            );
            const used = usedProducts.find((p) => p.id === selected.id);
            setProductData((prev) => ({
              ...prev,
              id: selected.id,
              unit: selected.unit,
              amount_used: used ? used.amount_used : "",
            }));
            onSelectNewProduct(selected);
          }}
        />
      </div>
      <div
        className={`w-full overflow-hidden transition-all duration-300 ease-in-out ${
          productData.id !== 0 || !isWindowSmall
            ? "max-h-[300px] opacity-100"
            : "pointer-events-none max-h-0 opacity-0"
          }`}
      >
        <div className="flex w-full flex-col gap-4">
          <Input
            styles="p-2 w-full"
            id="productAmount"
            name="productAmount"
            type="number"
            min={1}
            labelText={`Amount ${productData.unit && `(${productData.unit})`}`}
            hasError={false}
            placeholder="40"
            value={productData.amount_used}
            inputData={(data) =>
              setProductData((prev) => ({ ...prev, amount_used: data.value }))
            }
          />

          <div className="flex justify-end gap-4">
            {isEdit ? (
              <Button
                styles="py-2 px-4 text-sm"
                variant="secondary"
                buttonText="Save"
                type="submit"
              />
            ) : (
              <Button
                styles="py-2 px-4 text-sm"
                variant="secondary"
                buttonText="Add Product"
                type="submit"
              >
                <PlusIcon styles="size-5 sm:mr-1" />
              </Button>
            )}
            {isWindowSmall && (
              <Button
                styles="py-2 px-4 text-sm"
                variant="tertiary"
                buttonText="Cancel"
                type="button"
                onClick={() => {
                  setProductData({ id: 0, unit: "", amount_used: "" });
                }}
              />
            )}
          </div>
        </div>
      </div>
    </form>
  );
}
