import { Loading, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouteContext } from "@tanstack/react-router";
import { useState } from "react";
import ProductModal from "./-components/ProductModal";
import ProductsTable from "./-components/ProductsTable";

async function fetchProducts(merchantId) {
  const response = await fetch(`/api/v1/merchants/${merchantId}/products`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function productsQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "products"],
    queryFn: () => fetchProducts(merchantId),
  });
}

export const Route = createFileRoute("/_authenticated/_sidepanel/products/")({
  component: ProductsPage,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(productsQueryOptions(merchantId));
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function ProductsPage() {
  const [showProductModal, setShowProductModal] = useState(false);
  const [modalData, setModalData] = useState();
  const [serverError, setServerError] = useState();
  const { queryClient } = useRouteContext({ from: Route.id });
  const { showToast } = useToast();
  const { merchantId } = useAuth();

  const { data, isLoading, isError, error } = useQuery(
    productsQueryOptions(merchantId)
  );

  async function deleteHandler(product) {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/products/${product.id}`,
        {
          method: "DELETE",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
        }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        await queryClient.invalidateQueries({
          queryKey: [merchantId, "products"],
        });

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
      url = `/api/v1/merchants/${merchantId}/products/${product.id}`;
      method = "PUT";
    } else {
      // for correct json sending
      delete product.id;
      url = `/api/v1/merchants/${merchantId}/products`;
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
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message:
            method === "POST"
              ? "Product added successfully"
              : "Product modified successfully",
          variant: "success",
        });
        setServerError();

        await queryClient.invalidateQueries({
          queryKey: [merchantId, "products"],
        });
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setModalData();
    }
  }

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  return (
    <div className="flex h-screen justify-center px-4 py-2 md:px-0 md:py-0">
      <ProductModal
        key={modalData?.id || "new"}
        data={modalData}
        isOpen={showProductModal}
        onClose={() => setShowProductModal(false)}
        onSubmit={modalHandler}
      />
      <div className="flex w-full flex-col gap-5 pb-4">
        <p className="text-xl">Products</p>
        <ServerError error={serverError} />
        <div className="h-2/3 w-full">
          <ProductsTable
            products={data}
            onNewItem={() => {
              // the first condition is necessary for it to not cause an error
              // in case of a new item
              if (modalData && modalData.id) {
                setModalData();
              }
              setTimeout(() => setShowProductModal(true), 0);
            }}
            onEdit={(product) => {
              setModalData(product);
              setTimeout(() => setShowProductModal(true), 0);
            }}
            onDelete={deleteHandler}
          />
        </div>
      </div>
    </div>
  );
}
