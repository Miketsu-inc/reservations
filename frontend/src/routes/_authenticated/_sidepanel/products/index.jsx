import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import ProductModal from "./-components/ProductModal";
import ProductsTable from "./-components/ProductsTable";

async function fetchTableData() {
  const response = await fetch(`/api/v1/merchants/products/table-info`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalSotrageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export const Route = createFileRoute("/_authenticated/_sidepanel/products/")({
  component: ProductsPage,
  loader: async () => {
    const data = await fetchTableData();

    return {
      crumb: "Products",
      data: data,
    };
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function ProductsPage() {
  const router = useRouter();
  const loaderData = Route.useLoaderData();
  const [showProductModal, setShowProductModal] = useState(false);
  const [modalData, setModalData] = useState();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  async function deleteHandler(product) {
    try {
      const response = await fetch(`/api/v1/merchants/products/${product.id}`, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
      });

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        router.invalidate();
        setServerError();
        showToast({
          message: "Product deleted successfully",
          variant: "success",
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  async function modalHandler(product) {
    let url = "";
    let method = "";

    // means that the product was already added and now should be modified
    if (product.id != null) {
      url = `/api/v1/merchants/products/${product.id}`;
      method = "PUT";
    } else {
      // for correct json sending
      delete product.id;
      url = "/api/v1/merchants/products";
      method = "POST";
    }

    try {
      const response = await fetch(url, {
        method: method,
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(product),
      });

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        router.invalidate();
        showToast({
          message:
            method === "POST"
              ? "Product added successfully"
              : "Product modified successfully",
          variant: "success",
        });
        setServerError();
        router.invalidate();
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setModalData();
    }
  }

  return (
    <div className="flex h-screen justify-center">
      <ProductModal
        data={modalData}
        isOpen={showProductModal}
        onClose={() => setShowProductModal(false)}
        onSubmit={modalHandler}
        serviceData={loaderData.data.services}
      />
      <div className="w-full">
        <ServerError error={serverError} />
        <p className="text-xl">Products</p>
        <ProductsTable
          products={loaderData.data.products}
          serviceData={loaderData.data.services}
          onNewItem={() => {
            // the first condition is necessary for it to not cause an error
            // in case of a new item
            if (modalData && modalData.id) {
              setModalData();
            }
            setShowProductModal(true);
          }}
          onEdit={(product) => {
            setModalData(product);
            setShowProductModal(true);
          }}
          onDelete={deleteHandler}
        />
      </div>
    </div>
  );
}
