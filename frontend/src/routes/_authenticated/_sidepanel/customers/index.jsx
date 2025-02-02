import ServerError from "@components/ServerError";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import CustomerModal from "./-components/CustomerModal";
import CustomersTable from "./-components/CustomersTable";

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
  const [modalData, setModalData] = useState();
  const [serverError, setServerError] = useState();

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
        setServerError();
        router.invalidate();
        setModalData();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="flex h-screen justify-center px-4">
      <CustomerModal
        data={modalData}
        isOpen={showCustomerModal}
        onClose={() => setShowCustomerModal(false)}
        onSubmit={modalHandler}
      />
      <div className="w-full md:w-3/4">
        <ServerError error={serverError} />
        <p className="text-xl">Customers</p>
        <CustomersTable
          customersData={loaderData}
          onNewItem={() => {
            // the first condition is necessary for it to not cause an error
            // in case of a new item
            if (modalData && modalData.id) {
              setModalData();
            }
            setShowCustomerModal(true);
          }}
          onEdit={(customer) => {
            setModalData(customer);
            setShowCustomerModal(true);
          }}
          onDelete={deleteHandler}
        />
      </div>
    </div>
  );
}
