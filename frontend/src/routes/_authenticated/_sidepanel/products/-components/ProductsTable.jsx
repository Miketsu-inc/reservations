import DeleteModal from "@components/DeleteModal";
import Loading from "@components/Loading";
import { useWindowSize } from "@lib/hooks";
import { lazy, Suspense, useState } from "react";
import TableActions from "../../-components/TableActions";

const Table = lazy(() => import("@components/Table"));

function currencyFormatter(params) {
  return params.value.toLocaleString();
}

export default function ProductsTable({
  products,
  serviceData,
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

  const isSmallScreen = windowSize !== "xl" || windowSize !== "2xl";

  const columnDef = [
    { field: "id", hide: true },
    {
      field: "name",
      headerName: "Product Name",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "description",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
      headerName: "Description",
    },
    {
      field: "price",
      headerName: "Price (HUF)",
      valueFormatter: currencyFormatter,
      cellClass: "text-right",
    },
    {
      field: "stock_quantity",
      headerName: "In Stock",
      cellClass: "text-right",
    },
    {
      headerName: "Connected Services",
      cellRenderer: (params) => {
        return (
          <div className="flex h-full w-full items-center justify-center gap-2">
            {params.data.service_ids?.map((serviceId) => {
              // Find the service that matches the current serviceId
              const service = serviceData.find(
                (service) => service.Id === serviceId
              );
              return (
                service && (
                  <span
                    key={service.Id}
                    className="h-4 w-4 shrink-0 rounded-full"
                    style={{
                      backgroundColor: service.Color,
                    }}
                  ></span>
                )
              );
            })}
          </div>
        );
      },
      sortable: false,
    },
    {
      field: "actions",
      headerName: "",
      cellRenderer: (params) => {
        return (
          <TableActions
            key={params.data.id}
            onEdit={() => onEdit(products[params.node.sourceRowIndex])}
            onDelete={() => {
              setSelected({ id: params.data.id, name: params.data.name });
              setShowDeleteModal(true);
            }}
          />
        );
      },
      resizable: false,
      sortable: false,
    },
  ];

  return (
    <div className="h-full w-full">
      <DeleteModal
        itemName={selected.name}
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onDelete={() => onDelete(selected)}
      ></DeleteModal>
      <Suspense fallback={<Loading />}>
        <Table
          rowData={products}
          columnDef={columnDef}
          columnsToAutoSize={["price", "stock_quantity", "actions"]}
          itemName="product"
          onNewItem={onNewItem}
        />
      </Suspense>
    </div>
  );
}
