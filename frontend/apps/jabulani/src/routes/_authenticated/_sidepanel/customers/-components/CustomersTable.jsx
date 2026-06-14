import {
  CheckmarkCircle02Icon,
  Delete02Icon,
  Edit03Icon,
  UnavailableIcon,
  User03Icon,
  UserSwitchIcon,
} from "@hugeicons/core-free-icons";
import { Avatar, DeleteModal, Icon, Loading } from "@reservations/components";
import { useWindowSize } from "@reservations/lib";
import { lazy, Suspense, useState } from "react";
import BookingRing from "./BookingRing";

const Table = lazy(() => import("@reservations/components/Table"));

export default function CustomersTable({
  customersData,
  onDelete,
  onEdit,
  onTransfer,
  onNewItem,
  onBlackList,
  onRowClick,
  noRowsOverlayComponent,
}) {
  const windowSize = useWindowSize();
  const [selected, setSelected] = useState({
    id: 0,
    first_name: "",
    last_name: "",
  });
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const isSmallScreen =
    windowSize === "sm" || windowSize === "md" || windowSize === "lg";

  const columnDef = [
    { field: "id", hide: true },
    {
      field: "name",
      headerName: "Name",
      flex: 1,
      resizable: false,
      ...(isSmallScreen ? { minWidth: 180 } : {}),
      valueGetter: (params) =>
        `${params.data.first_name} ${params.data.last_name}`.trim(),
      cellRenderer: (params) => {
        return (
          <div className="flex h-full flex-row items-center gap-2">
            <Avatar
              styles="size-12! text-sm shrink-0"
              initials={`${params.data.first_name[0]}${params.data.last_name[0]}`}
            />
            <div className="flex flex-col justify-center">
              <p className="text-base">{`${params.data.first_name} ${params.data.last_name}`}</p>
              <p className="text-text_color/60 text-sm">{params.data.email}</p>
            </div>
          </div>
        );
      },
    },
    {
      field: "phone_number",
      headerName: "Phone number",
      flex: 1,
      resizable: false,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
      cellStyle: {
        display: "flex",
        alignItems: "center",
      },
    },
    {
      field: "booking_history",
      headerName: "Booking history",
      flex: 1,
      resizable: false,
      hide: isSmallScreen,
      cellRenderer: (params) => {
        return (
          <div className="flex h-full flex-row items-center gap-2">
            <p>
              {params.data.times_booked} / {params.data.times_cancelled}
            </p>
            {params.data.times_booked || params.data.times_cancelled ? (
              <BookingRing
                booked={params.data.times_booked}
                cancelled={params.data.times_cancelled}
              />
            ) : (
              <></>
            )}
          </div>
        );
      },
    },
    {
      field: "times_booked",
      headerName: "Total Bookings",
      sort: "desc",
      hide: true,
    },
    {
      field: "times_cancelled",
      headerName: "Cancellations",
      hide: true,
    },
    {
      field: "actions",
      flex: 1,
      headerName: "",
      cellRenderer: (params) => {
        return (
          <div
            key={params.data.id}
            className="flex h-full flex-row items-center justify-center"
          >
            <button
              className="cursor-pointer"
              onClick={() => onEdit(customersData[params.node.sourceRowIndex])}
            >
              <Icon icon={Edit03Icon} styles="size-5 mx-1" />
            </button>
            {params.data.is_dummy ? (
              <>
                {params.data.times_booked || params.data.times_cancelled ? (
                  <button
                    className="cursor-pointer"
                    onClick={() => onTransfer(params.data.id)}
                  >
                    <Icon icon={UserSwitchIcon} styles="size-5 mx-1" />
                  </button>
                ) : (
                  <></>
                )}
                <button
                  className="cursor-pointer"
                  onClick={() => {
                    setSelected({
                      id: params.data.id,
                      first_name: params.data.first_name,
                      last_name: params.data.last_name,
                    });
                    setShowDeleteModal(true);
                  }}
                >
                  <Icon
                    icon={Delete02Icon}
                    styles="size-5 mx-1 text-red-600 dark:text-red-500"
                  />
                </button>
              </>
            ) : (
              // the handler will decide to blacklist or to delete from blacklist
              <>
                {params.data.is_blacklisted ? (
                  <button
                    className="cursor-pointer"
                    onClick={() => onBlackList(params.data)}
                  >
                    <Icon icon={CheckmarkCircle02Icon} styles="size-5 mx-1" />
                  </button>
                ) : (
                  <button
                    className="cursor-pointer"
                    onClick={() => onBlackList(params.data)}
                  >
                    <Icon icon={UnavailableIcon} styles="size-5 mx-1" />
                  </button>
                )}
              </>
            )}
          </div>
        );
      },
      resizable: false,
      sortable: false,
      minWidth: 120,
      maxWidth: 120,
    },
  ];

  return (
    <div className="h-full w-full">
      <DeleteModal
        itemName={`${selected.first_name} ${selected.last_name}`}
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onDelete={() => onDelete(selected)}
      />
      <Suspense fallback={<Loading />}>
        <Table
          rowData={customersData}
          columnDef={columnDef}
          itemName="customer"
          onNewItem={onNewItem}
          exportName="customers_table"
          onRowClick={onRowClick}
          columnsToExport={[
            "name",
            "email",
            "phone_number",
            "times_booked",
            "times_cancelled",
          ]}
          noRowsOverlayComponent={
            noRowsOverlayComponent || DefaultNoRowsOverlay
          }
          showControls={false}
          rowHeight={80}
        />
      </Suspense>
    </div>
  );
}

function DefaultNoRowsOverlay() {
  return (
    <div
      className="mb-16 flex flex-col items-center gap-4 px-2 text-gray-600
        dark:text-gray-300"
    >
      <Icon icon={User03Icon} styles="text-gray-200 size-16" />
      <div className="flex flex-col items-center gap-2">
        <p className="text-text_color text-base font-medium">
          No customers found
        </p>
        <p className="text-sm md:w-2/3">
          Once a customer makes a booking or a purchase, their information will
          appear here. You can use this list to track and manage your regular
          clients.
        </p>
      </div>
    </div>
  );
}
