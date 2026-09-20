import {
  BirthdayCakeIcon,
  Call02Icon,
  CheckmarkCircle02Icon,
  Delete02Icon,
  Edit03Icon,
  Mail01Icon,
  MoreVerticalIcon,
  UnavailableIcon,
  UserSwitchIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Card,
  DeleteModal,
  Icon,
  Loading,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  ServerError,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  customersQueryOptions,
  invalidateLocalStorageAuth,
  useToast,
  useWindowSize,
} from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import BlacklistModal from "../-components/BlacklistModal";
import TransferAppsModal from "../-components/TransferAppsModal";
import BookingItem from "./-components/BookingItem";
import CustomerStats from "./-components/CustomerStats";
import ExpandableNote from "./-components/ExpandableNote";
import PaginatedList from "./-components/PaginatedList";

async function fetchCustomerInfo(merchantId, customerId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/customers/${customerId}/stats`,
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

function customerInfoQueryOptions(merchantId, id) {
  return queryOptions({
    queryKey: [merchantId, "customer-info", id],
    queryFn: () => fetchCustomerInfo(merchantId, id),
  });
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
  "/_authenticated/_sidepanel/customers/_topnav/$customerId/"
)({
  component: CustomerDetailsPage,
  loader: async ({
    params,
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(
      customerInfoQueryOptions(merchantId, params.customerId)
    );
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function CustomerDetailsPage() {
  const navigate = Route.useNavigate();
  const { windowSize } = useWindowSize();
  const [showBlacklistModal, setShowBlacklistModal] = useState(false);
  const [showTransferModal, setShowTransferModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();
  const { merchantId } = useAuth();

  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { customerId } = Route.useParams();

  const { data, isLoading, isError, error } = useQuery(
    customerInfoQueryOptions(merchantId, customerId)
  );

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  const completedBookings = data.bookings.filter(
    (booking) => booking.status === "completed"
  );

  const lastVisited = completedBookings[0]
    ? monthDateFormat(new Date(completedBookings[0].to_date))
    : null;

  async function deleteHandler(id) {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/customers/${id}`,
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
        showToast({
          message: "Customer deleted successfully",
          variant: "success",
        });
        await queryClient.invalidateQueries(customersQueryOptions(merchantId));
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

    if (data.method === "PUT") {
      options.body = JSON.stringify({
        id: data.id,
        blacklist_reason: data.blacklist_reason,
      });
    }

    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/customers/${data.id}/blacklist`,
        options
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        if (data.method === "PUT") {
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
        await queryClient.invalidateQueries({
          queryKey: [merchantId, "customer-info", customerId],
        });
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="flex justify-center pt-4">
      <TransferAppsModal
        fromCustomerId={data.id}
        isOpen={showTransferModal}
        onClose={() => setShowTransferModal(false)}
      />
      <BlacklistModal
        key={data?.id || "new"}
        data={data}
        isOpen={showBlacklistModal}
        onClose={() => setShowBlacklistModal(false)}
        onSubmit={(customer) =>
          blacklistHandler({
            method: customer.is_blacklisted ? "DELETE" : "PUT",
            id: customer.id,
            blacklist_reason: customer.blacklist_reason,
          })
        }
      />
      <DeleteModal
        itemName={`${data.first_name} ${data.last_name}`}
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onDelete={() => deleteHandler(data.id)}
      />
      <div
        className="flex w-full flex-col gap-5 px-3 lg:w-2/3 lg:px-0 2xl:w-1/2"
      >
        <ServerError error={serverError} />
        <Card styles="flex flex-col items-start gap-4">
          <div className="flex w-full justify-between">
            <div className="flex items-center gap-4">
              <Avatar
                initials={`${data.first_name.charAt(0)}${data.last_name.charAt(0)}`}
              />

              <div
                className={`flex flex-col ${lastVisited ? "gap-2" : "gap-0"}`}
              >
                <div
                  className={`flex flex-col gap-2
                    ${lastVisited && windowSize !== "sm" ? "sm:flex-row sm:gap-4" : ""}`}
                >
                  <h2 className="text-text_color text-lg font-bold">
                    {data.first_name} {data.last_name}
                  </h2>
                  {data.is_blacklisted && (
                    <span
                      className="inline-flex w-fit items-center gap-1
                        rounded-full bg-red-700/20 px-2 py-0.5 text-xs
                        font-medium text-red-800 dark:text-red-500"
                    >
                      <Icon icon={UnavailableIcon} styles="size-4" />
                      Blacklisted
                    </span>
                  )}

                  {data.is_dummy && (
                    <span
                      className="bg-hvr_gray text-text_color/90 w-fit
                        rounded-full px-2 py-0.5 text-xs font-medium"
                    >
                      Customer Added by You
                    </span>
                  )}
                </div>

                {lastVisited &&
                  (windowSize !== "sm" ||
                    (!data.is_blacklisted && !data.is_dummy)) && (
                    <p className="text-text_color/70 text-sm">
                      Last visited: {lastVisited}
                    </p>
                  )}
              </div>
            </div>
            <div className="flex flex-col items-start">
              <Popover>
                <PopoverTrigger asChild>
                  <button
                    className="hover:bg-hvr_gray hover:*:stroke-text_color h-fit
                      cursor-pointer rounded-lg p-1"
                  >
                    <Icon
                      icon={MoreVerticalIcon}
                      styles="size-6 text-gray-400 dark:text-gray-500 rotate-90"
                    />
                  </button>
                </PopoverTrigger>
                <PopoverContent side="left" styles="w-auto">
                  <div
                    className="itmes-start flex w-auto flex-col *:flex *:w-full
                      *:flex-row *:items-center *:rounded-lg *:p-2"
                  >
                    {!data.is_dummy && (
                      <PopoverClose asChild>
                        <button
                          onClick={() => setShowBlacklistModal(true)}
                          className="hover:bg-hvr_gray text-text_color
                            cursor-pointer gap-3"
                        >
                          {!data.is_blacklisted ? (
                            <Icon
                              icon={UnavailableIcon}
                              styles="size-6 shrink-0"
                            />
                          ) : (
                            <Icon
                              icon={CheckmarkCircle02Icon}
                              styles="size-6"
                            />
                          )}
                          <p className="text-nowrap">
                            {!data.is_blacklisted
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
                          navigate({
                            from: Route.fullPath,
                            to: `/customers/edit/${data.id}`,
                          });
                        }}
                      >
                        <Icon icon={Edit03Icon} styles="size-5" />
                        <p>Edit customer</p>
                      </button>
                    </PopoverClose>
                    {data.is_dummy && (
                      <>
                        {data.bookings.length !== 0 ? (
                          <PopoverClose asChild>
                            <button
                              className="hover:bg-hvr_gray cursor-pointer gap-3"
                              onClick={() => setShowTransferModal(true)}
                            >
                              <Icon icon={UserSwitchIcon} styles="size-5" />
                              <p>Transfer customer</p>
                            </button>
                          </PopoverClose>
                        ) : (
                          <></>
                        )}
                        <PopoverClose asChild>
                          <button
                            onClick={() => setShowDeleteModal(true)}
                            className="hover:bg-hvr_gray cursor-pointer gap-3"
                          >
                            <Icon
                              icon={Delete02Icon}
                              styles="size-5 mb-0.5 text-red-600
                                dark:text-red-500"
                            />
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

          <div
            className="text-text_color/70 flex flex-col items-start gap-3
              text-sm sm:flex-row sm:items-center sm:gap-6"
          >
            {data.email && (
              <div className="flex items-center gap-2">
                <Icon icon={Mail01Icon} styles="size-5 text-text_color/70" />
                {data.email}
              </div>
            )}
            <div className="flex items-center gap-6 sm:justify-start">
              {data.phone_number && (
                <div className="flex items-center gap-2">
                  <Icon
                    icon={Call02Icon}
                    styles="size-4 mb-0.5 text-text_color/70"
                  />
                  {data.phone_number}
                </div>
              )}
              {data.birthday && (
                <div className="flex items-center gap-2">
                  <Icon
                    icon={BirthdayCakeIcon}
                    styles="size-5 mb-0.5 text-text_color/70"
                  />
                  {formatBirthday(data.birthday)}
                </div>
              )}
            </div>
          </div>
          <ExpandableNote text={data.note} />
          <CustomerStats customer={data} />
        </Card>

        <PaginatedList
          data={data.bookings}
          itemsPerPage={8}
          title="Booking History"
          emptyMessage="No bookings found for this customer"
          renderItem={(booking) => (
            <BookingItem booking={booking} customerName={data.first_name} />
          )}
        />
      </div>
    </div>
  );
}
