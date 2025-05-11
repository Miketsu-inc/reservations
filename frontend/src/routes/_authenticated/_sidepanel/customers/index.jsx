import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import BlacklistModal from "./-components/BlacklistModal";
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
    invalidateLocalSotrageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export const Route = createFileRoute("/_authenticated/_sidepanel/customers/")({
  component: CustomersPage,
  loader: async () => {
    const customers = await fetchCustomers();

    return {
      crumb: "Customers",
      customers: customers,
    };
  },
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
  const [showBlacklistModal, setShowBlacklistModal] = useState(false);
  const [blacklistModalData, setBlacklistModalData] = useState();
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
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setCustomerModalData();
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

  async function blacklistHandler(data) {
    try {
      const response = await fetch(
        `/api/v1/merchants/customers/blacklist/${data.id}`,
        {
          method: data.method,
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
        if (data.method === "POST") {
          showToast({
            message: "Customer blacklisted successfully",
            variant: "success",
          });
        } else if (data.method === "DELETE") {
          showToast({
            message: "Customer removed from blacklist successfully",
            variant: "success",
          });
        }
        router.invalidate();
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="flex h-screen justify-center px-4 py-2 md:px-0 md:py-0">
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
      <BlacklistModal
        data={blacklistModalData}
        isOpen={showBlacklistModal}
        onClose={() => setShowBlacklistModal(false)}
        // both adding to and removing from blacklist goes through the same modal and handler
        // so the customer.is_blacklisted field determines the action
        onSubmit={(customer) =>
          blacklistHandler({
            method: customer.is_blacklisted ? "DELETE" : "POST",
            id: customer.id,
          })
        }
      />
      <div className="flex w-full flex-col gap-5 py-4">
        <p className="text-xl">Customers</p>
        <ServerError error={serverError} />
        <div className="h-2/3 w-full">
          <CustomersTable
            customersData={loaderData.customers}
            onNewItem={() => {
              // the first condition is necessary for it to not cause an error
              // in case of a new item
              if (customerModalData && customerModalData.id) {
                setCustomerModalData();
              }
              setTimeout(() => setShowCustomerModal(true), 0);
            }}
            onTransfer={(index) => {
              setTransferModalData({
                fromIndex: index,
                customers: loaderData.customers,
              });
              setTimeout(() => setShowTransferModal(true), 0);
            }}
            onEdit={(customer) => {
              setCustomerModalData(customer);
              setTimeout(() => setShowCustomerModal(true), 0);
            }}
            onDelete={deleteHandler}
            onBlackList={(customer) => {
              setBlacklistModalData(customer);
              setTimeout(() => setShowBlacklistModal(true), 0);
            }}
          />
        </div>
      </div>
    </div>
  );
}
