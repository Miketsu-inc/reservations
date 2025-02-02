import DeleteModal from "@components/DeleteModal";
import Loading from "@components/Loading";
import { useWindowSize } from "@lib/hooks";
import { lazy, Suspense, useState } from "react";
import TableActions from "../../-components/TableActions";

const Table = lazy(() => import("@components/Table"));

export default function CustomersTable({
  customersData,
  onDelete,
  onEdit,
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
    { field: "id", hide: true, sort: "asc" },
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
      field: "times_booked",
      headerName: "Times booked",
      cellClass: "text-right",
    },
    {
      field: "times_cancelled",
      headerName: "Times cancelled",
      cellClass: "text-right",
    },
    {
      field: "is_dummy",
      headerName: "Add by me",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "actions",
      headerName: "",
      cellRenderer: (params) => {
        if (params.data.is_dummy) {
          return (
            <TableActions
              key={params.data.id}
              onEdit={() => onEdit(customersData[params.node.sourceRowIndex])}
              onDelete={() => {
                setSelected({
                  id: params.data.id,
                  first_name: params.data.first_name,
                  last_name: params.data.last_name,
                });
                setShowDeleteModal(true);
              }}
            />
          );
        }
      },
      resizable: false,
      sortable: false,
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
            "first_name",
            "last_name",
            "times_booked",
            "times_cancelled",
            "is_dummy",
            "actions",
          ]}
          onNewItem={onNewItem}
        />
      </Suspense>
    </div>
  );
}
