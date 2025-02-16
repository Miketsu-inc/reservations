import DeleteModal from "@components/DeleteModal";
import Loading from "@components/Loading";
import ApproveIcon from "@icons/ApproveIcon";
import BanIcon from "@icons/BanIcon";
import EditIcon from "@icons/EditIcon";
import TransferIcon from "@icons/TransferIcon";
import TrashBinIcon from "@icons/TrashBinIcon";
import { useWindowSize } from "@lib/hooks";
import { lazy, Suspense, useState } from "react";
import BookingRing from "./BookingRing";

const Table = lazy(() => import("@components/Table"));

export default function CustomersTable({
  customersData,
  onDelete,
  onEdit,
  onTransfer,
  onNewItem,
  onBlackList,
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
      field: "first_name",
      headerName: "First name",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "last_name",
      headerName: "Last name",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "email",
      headerName: "Email",
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "phone_number",
      headerName: "Phone number",
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "booking_history",
      headerName: "Booking history",
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
      sort: "desc",
      hide: true,
    },
    {
      field: "is_dummy",
      headerName: "Add by me",
    },
    {
      field: "actions",
      headerName: "",
      cellRenderer: (params) => {
        return (
          <div
            key={params.data.id}
            className="flex h-full flex-row items-center justify-center"
          >
            {params.data.is_dummy ? (
              <>
                {params.data.times_booked || params.data.times_cancelled ? (
                  <button
                    className="cursor-pointer"
                    onClick={() => onTransfer(params.node.sourceRowIndex)}
                  >
                    <TransferIcon styles="w-5 h-5 mx-1" />
                  </button>
                ) : (
                  <></>
                )}
                <button
                  className="cursor-pointer"
                  onClick={() =>
                    onEdit(customersData[params.node.sourceRowIndex])
                  }
                >
                  <EditIcon styles="w-4 h-4 mx-1" />
                </button>
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
                  <TrashBinIcon styles="w-5 h-5 text-white mx-1" />
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
                    <ApproveIcon styles="w-5 h-5" />
                  </button>
                ) : (
                  <button
                    className="cursor-pointer"
                    onClick={() => onBlackList(params.data)}
                  >
                    <BanIcon styles="w-5 h-5" />
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
          columnsToAutoSize={[
            "email",
            "phone_number",
            "booking_history",
            "is_dummy",
          ]}
          onNewItem={onNewItem}
        />
      </Suspense>
    </div>
  );
}
