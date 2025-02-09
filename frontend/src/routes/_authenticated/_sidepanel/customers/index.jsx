import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import CustomerModal from "./-components/CustomerModal";
import CustomersTable from "./-components/CustomersTable";
import TransferAppsModal from "./-components/TransferAppsModal";

async function fetchCustomers() {
  const response = await fetch(`/api/v1/merchants/customers`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

export const Route = createFileRoute("/_authenticated/_sidepanel/customers/")({
  component: CustomersPage,
  loader: () => fetchCustomers(),
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function CustomersPage() {
  const router = useRouter();
  const loaderData = Route.useLoaderData();
  const [showCustomerModal, setShowCustomerModal] = useState(false);
  const [customerModalData, setCustomerModalData] = useState();
  const [showTransferModal, setShowTransferModal] = useState(false);
  const [transferModalData, setTransferModalData] = useState();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  async function deleteHandler(selected) {
    try {
      const response = await fetch(
        `/api/v1/merchants/customers/${selected.id}`,
        {
          method: "DELETE",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
        }
      );

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Customer deleted successfully",
          variant: "success",
        });
        router.invalidate();
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  async function modalHandler(customer) {
    let url = "";
    let method = "";

    // means that the customer was already added and now should be modified
    if (customer.id != null) {
      url = `/api/v1/merchants/customers/${customer.id}`;
      method = "PUT";
    } else {
      // for correct json sending
      delete customer.id;
      url = "/api/v1/merchants/customers";
      method = "POST";
    }

    try {
      const response = await fetch(url, {
        method: method,
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(customer),
      });

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message:
            method === "POST"
              ? "Customer added successfully"
              : "Customer modified successfully",
          variant: "success",
        });
        setServerError();
        router.invalidate();
        setCustomerModalData();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  async function transferHandler(data) {
    try {
      const response = await fetch(
        `/api/v1/merchants/customers/transfer?from=${data.from}&to=${data.to}`,
        {
          method: "PUT",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
        }
      );

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Appointments transferred successfully",
          variant: "success",
        });
        router.invalidate();
        setServerError();
        setTransferModalData();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="flex h-screen justify-center px-4">
      <CustomerModal
        data={customerModalData}
        isOpen={showCustomerModal}
        onClose={() => setShowCustomerModal(false)}
        onSubmit={modalHandler}
      />
      <TransferAppsModal
        data={transferModalData}
        isOpen={showTransferModal}
        onClose={() => setShowTransferModal(false)}
        onSubmit={transferHandler}
      />
      <div className="w-full md:w-3/4">
        <ServerError error={serverError} />
        <p className="text-xl">Customers</p>
        <CustomersTable
          customersData={loaderData}
          onNewItem={() => {
            // the first condition is necessary for it to not cause an error
            // in case of a new item
            if (customerModalData && customerModalData.id) {
              setCustomerModalData();
            }
            setShowCustomerModal(true);
          }}
          onTransfer={(index) => {
            setTransferModalData({ fromIndex: index, customers: loaderData });
            setShowTransferModal(true);
          }}
          onEdit={(customer) => {
            setCustomerModalData(customer);
            setShowCustomerModal(true);
          }}
          onDelete={deleteHandler}
        />
      </div>
    </div>
  );
}
