import DeleteModal from "@components/DeleteModal";
import Loading from "@components/Loading";
import EditIcon from "@icons/EditIcon";
import TransferIcon from "@icons/TransferIcon";
import TrashBinIcon from "@icons/TrashBinIcon";
import { useWindowSize } from "@lib/hooks";
import { lazy, Suspense, useState } from "react";

const Table = lazy(() => import("@components/Table"));

export default function CustomersTable({
  customersData,
  onDelete,
  onEdit,
  onTransfer,
  onNewItem,
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
      field: "times_booked",
      headerName: "Times booked",
      cellClass: "text-right",
      sort: "desc",
    },
    {
      field: "times_cancelled",
      headerName: "Times cancelled",
      cellClass: "text-right",
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
              <></>
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
      ></DeleteModal>
      <Suspense fallback={<Loading />}>
        <Table
          rowData={customersData}
          columnDef={columnDef}
          itemName="customer"
          columnsToAutoSize={[
            "email",
            "phone_number",
            "times_booked",
            "times_cancelled",
            "is_dummy",
          ]}
          onNewItem={onNewItem}
        />
      </Suspense>
    </div>
  );
}
