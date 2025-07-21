import Card from "@components/Card";
import DeleteModal from "@components/DeleteModal";
import { Popover, PopoverContent, PopoverTrigger } from "@components/Popover";
import ServerError from "@components/ServerError";
import ApproveIcon from "@icons/ApproveIcon";
import BanIcon from "@icons/BanIcon";
import CakeIcon from "@icons/CakeIcon";
import EditIcon from "@icons/EditIcon";
import EnvelopeIcon from "@icons/EnvelopeIcon";
import PhoneIcon from "@icons/PhoneIcon";
import ThreeDotsIcon from "@icons/ThreeDotsIcon";
import TrashBinIcon from "@icons/TrashBinIcon";
import { useToast, useWindowSize } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { PopoverClose } from "@radix-ui/react-popover";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import BlacklistModal from "../-components/BlacklistModal";
import AppointmentItem from "./-components/AppointmentItem";
import CustomerStats from "./-components/CustomerStats";
import ExpandableNote from "./-components/ExpandableNote";
import PaginatedList from "./-components/PaginatedList";

async function fetchCustomerInfo(customerId) {
  const response = await fetch(
    `/api/v1/merchants/customers/stats/${customerId}`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function monthDateFormat(date) {
  return date.toLocaleDateString([], {
    weekday: "short",
    month: "short",
    day: "numeric",
  });
}

function formatBirthday(datestr) {
  const date = new Date(datestr);
  return date.toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/$customerId/"
)({
  component: CustomerDetailsPage,
  loader: async ({ params }) => {
    return fetchCustomerInfo(params.customerId);
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.mesage} />;
  },
});

function CustomerDetailsPage() {
  const router = useRouter();
  const navigate = Route.useNavigate();
  const loaderData = Route.useLoaderData();
  const windowSize = useWindowSize();
  const [showBlacklistModal, setShowBlacklistModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();
  const now = new Date();

  const completedAppointments = loaderData.appointments
    .filter((appt) => {
      const toDate = new Date(appt.to_date);
      const wasCancelled = appt.cancelled_by_user || appt.cancelled_by_merchant;
      return toDate < now && !wasCancelled;
    })
    .sort((a, b) => new Date(b.to_date) - new Date(a.to_date));

  const lastVisited = completedAppointments[0]
    ? monthDateFormat(new Date(completedAppointments[0].to_date))
    : null;

  async function deleteHandler(id) {
    try {
      const response = await fetch(`/api/v1/merchants/customers/${id}`, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Customer deleted successfully",
          variant: "success",
        });
        navigate({
          from: Route.fullPath,
          to: "/customers",
        });
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  async function blacklistHandler(data) {
    const options = {
      method: data.method,
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    };

    if (data.method === "POST") {
      options.body = JSON.stringify({
        id: data.id,
        reason: data.reason,
      });
    }

    try {
      const response = await fetch(
        `/api/v1/merchants/customers/blacklist/${data.id}`,
        options
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
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
    <div className="flex justify-center py-5">
      <BlacklistModal
        data={loaderData}
        isOpen={showBlacklistModal}
        onClose={() => setShowBlacklistModal(false)}
        onSubmit={(customer) =>
          blacklistHandler({
            method: customer.is_blacklisted ? "DELETE" : "POST",
            id: customer.id,
            reason: customer.reason,
          })
        }
      />
      <DeleteModal
        itemName={`${loaderData.first_name} ${loaderData.last_name}`}
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onDelete={() => deleteHandler(loaderData.id)}
      />
      <div className="flex w-full flex-col gap-5 px-3 sm:px-0 lg:w-2/3 2xl:w-1/2">
        <ServerError error={serverError} />
        <Card styles="flex flex-col items-start gap-4">
          <div className="flex w-full justify-between">
            <div className="flex items-center gap-4">
              <div className="from-secondary to-primary bg-primary flex size-16 items-center justify-center rounded-md text-lg text-white dark:bg-linear-to-br">
                {`${loaderData.first_name.charAt(0)}${loaderData.last_name.charAt(0)}`.toUpperCase()}
              </div>

              <div
                className={`flex flex-col ${lastVisited ? "gap-2" : "gap-0"}`}
              >
                <div
                  className={`flex flex-col gap-2 ${lastVisited && windowSize !== "sm" ? "sm:flex-row sm:gap-4" : ""}`}
                >
                  <h2 className="text-text_color text-lg font-bold">
                    {loaderData.first_name} {loaderData.last_name}
                  </h2>
                  {loaderData.is_blacklisted && (
                    <span className="inline-flex w-fit items-center gap-1 rounded-full bg-red-700/20 px-2 py-0.5 text-xs font-medium text-red-800 dark:text-red-500">
                      <BanIcon styles="size-4" />
                      Blacklisted
                    </span>
                  )}

                  {loaderData.is_dummy && (
                    <span className="bg-hvr_gray text-text_color/90 w-fit rounded-full px-2 py-0.5 text-xs font-medium">
                      User Added by You
                    </span>
                  )}
                </div>

                {lastVisited &&
                  (windowSize !== "sm" ||
                    (!loaderData.is_blacklisted && !loaderData.is_dummy)) && (
                    <p className="text-text_color/70 text-sm">
                      Last visited: {lastVisited}
                    </p>
                  )}
              </div>
            </div>
            <div className="flex flex-col items-start">
              <Popover>
                <PopoverTrigger asChild>
                  <button className="hover:bg-hvr_gray hover:*:stroke-text_color h-fit cursor-pointer rounded-lg p-1">
                    <ThreeDotsIcon styles="size-6 stroke-4 stroke-gray-400 dark:stroke-gray-500" />
                  </button>
                </PopoverTrigger>
                <PopoverContent side="left" styles="w-auto">
                  <div className="itmes-start flex w-auto flex-col *:flex *:w-full *:flex-row *:items-center *:rounded-lg *:p-2">
                    {!loaderData.is_dummy && (
                      <PopoverClose asChild>
                        <button
                          onClick={() => setShowBlacklistModal(true)}
                          className="hover:bg-hvr_gray text-text_color cursor-pointer gap-3"
                        >
                          {!loaderData.is_blacklisted ? (
                            <BanIcon styles="size-6 ml-0.5 shrink-0" />
                          ) : (
                            <ApproveIcon styles="size-6" />
                          )}
                          <p className="text-nowrap">
                            {!loaderData.is_blacklisted
                              ? "Blacklist Customer"
                              : "Unban customer"}
                          </p>
                        </button>
                      </PopoverClose>
                    )}
                    <PopoverClose asChild>
                      <button
                        className="hover:bg-hvr_gray cursor-pointer gap-3"
                        onClick={() => {
                          router.navigate({
                            from: Route.fullPath,
                            to: `/customers/edit/${loaderData.id}`,
                          });
                        }}
                      >
                        <EditIcon styles="size-4" />
                        <p>Edit customer</p>
                      </button>
                    </PopoverClose>
                    {loaderData.is_dummy && (
                      <>
                        <PopoverClose asChild>
                          <button
                            onClick={() => setShowDeleteModal(true)}
                            className="hover:bg-hvr_gray cursor-pointer gap-3"
                          >
                            <TrashBinIcon styles="size-5 mb-0.5" />
                            <p className="text-red-600 dark:text-red-500">
                              Delete Customer
                            </p>
                          </button>
                        </PopoverClose>
                      </>
                    )}
                  </div>
                </PopoverContent>
              </Popover>
            </div>
          </div>

          <div className="text-text_color/70 flex flex-col items-start gap-3 text-sm sm:flex-row sm:items-center sm:gap-6">
            {loaderData.email && (
              <div className="flex items-center gap-2">
                <EnvelopeIcon styles="size-5 text-text_color/70" />
                {loaderData.email}
              </div>
            )}
            <div className="flex items-center gap-6 sm:justify-start">
              {loaderData.phone_number && (
                <div className="flex items-center gap-2">
                  <PhoneIcon styles="size-4 mb-0.5 fill-text_color/70 stroke-text_color/10" />
                  {loaderData.phone_number}
                </div>
              )}
              {loaderData.birthday && (
                <div className="flex items-center gap-2">
                  <CakeIcon styles="size-5 mb-0.5 text-text_color/70" />
                  {formatBirthday(loaderData.birthday)}
                </div>
              )}
            </div>
          </div>
          <ExpandableNote text={loaderData.note} />
          <CustomerStats customer={loaderData} />
        </Card>

        <PaginatedList
          data={loaderData.appointments}
          itemsPerPage={8}
          title="Appointment History"
          emptyMessage="No appointments found for this customer"
          renderItem={(appointment) => (
            <AppointmentItem
              appointment={appointment}
              customerName={loaderData.first_name}
            />
          )}
        />
      </div>
    </div>
  );
}
