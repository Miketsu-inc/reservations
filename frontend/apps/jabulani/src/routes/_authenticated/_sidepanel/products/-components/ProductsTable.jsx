import { ProductIcon } from "@reservations/assets";
import { DeleteModal, Loading } from "@reservations/components";
import { getDisplayUnit, useWindowSize } from "@reservations/lib";
import { lazy, Suspense, useState } from "react";
import TableActions from "../../-components/TableActions";

const Table = lazy(() => import("@reservations/components/Table"));

function currencyFormatter(params) {
  return params.value?.number?.toLocaleString();
}

export default function ProductsTable({
  products,
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
      headerValueGetter: (params) =>
        `Price (${params.value?.currency || "HUF"})`,
      valueFormatter: currencyFormatter,
      cellClass: "text-right",
    },
    {
      headerName: "In Stock  [unit]",
      cellRenderer: (params) => {
        const { current_amount, max_amount, unit } = params.data;

        const {
          current,
          max,
          unit: displayUnit,
        } = getDisplayUnit(current_amount, max_amount, unit);

        return (
          <div className="flex items-center justify-center gap-6 text-center">
            <span>
              {current} / {max}
            </span>
            <span> {displayUnit} </span>
          </div>
        );
      },
      sortable: false,
      valueGetter: (params) => {
        const { current_amount, max_amount, unit } = params.data;
        const {
          current,
          max,
          unit: displayUnit,
        } = getDisplayUnit(current_amount, max_amount, unit);

        return `${current} / ${max} ${displayUnit}`;
      },
    },
    {
      headerName: "Connected Services",
      cellRenderer: (params) => {
        return (
          <div className="flex h-full w-full items-center justify-center gap-2">
            {params.data.services?.map((service) => (
              <span
                key={service.id}
                className="h-4 w-4 shrink-0 rounded-full"
                style={{ backgroundColor: service.color }}
              ></span>
            ))}
          </div>
        );
      },
      sortable: false,
      valueGetter: (params) => {
        return (
          params.data.services?.map((service) => service.name).join(", ") || ""
        );
      },
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
          columnsToAutoSize={["price", "stock_amount", "actions"]}
          itemName="product"
          onNewItem={onNewItem}
          exportName="product_table"
          noRowsOverlayComponent={DefaultNoRowsOverlay}
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
      <ProductIcon styles="size-16" />
      <div className="flex flex-col items-center gap-2">
        <p className="text-text_color text-base font-medium">
          No products added
        </p>
        <p className="text-sm md:w-2/3">
          Use this space to keep track of your inventory. Add items you offer so
          you can monitor stock levels, identify what’s running low, and stay
          organized.
        </p>
      </div>
    </div>
  );
}
