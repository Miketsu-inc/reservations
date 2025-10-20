import {
  ApproveIcon,
  BanIcon,
  EditIcon,
  PersonIcon,
  TransferIcon,
  TrashBinIcon,
} from "@reservations/assets";
import { DeleteModal, Loading } from "@reservations/components";
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
  const isSmallScreen = windowSize !== "xl" || windowSize !== "2xl";

  const columnDef = [
    { field: "id", hide: true },
    {
      field: "name",
      headerName: "Name",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 160 } : {}),
      valueGetter: (params) =>
        `${params.data.first_name} ${params.data.last_name}`.trim(),
    },
    {
      field: "email",
      headerName: "Email",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "phone_number",
      headerName: "Phone number",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "booking_history",
      headerName: "Booking history",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
      cellRenderer: (params) => {
        return (
          <div className="flex flex-row items-center justify-center gap-2">
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
              <EditIcon styles="size-4 mx-1" />
            </button>
            {params.data.is_dummy ? (
              <>
                {params.data.times_booked || params.data.times_cancelled ? (
                  <button
                    className="cursor-pointer"
                    onClick={() => onTransfer(params.data.id)}
                  >
                    <TransferIcon styles="size-5 mx-1" />
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
                  <TrashBinIcon styles="size-5 mx-1" />
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
                    <ApproveIcon styles="size-5 mx-1" />
                  </button>
                ) : (
                  <button
                    className="cursor-pointer"
                    onClick={() => onBlackList(params.data)}
                  >
                    <BanIcon styles="size-5 mx-1" />
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
      <PersonIcon styles="stroke-gray-200 size-16" />
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
