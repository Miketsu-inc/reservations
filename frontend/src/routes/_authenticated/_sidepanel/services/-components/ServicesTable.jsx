import {
  AllCommunityModule,
  colorSchemeDarkBlue,
  themeAlpine,
} from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";

import TableActions from "../../-components/TableActions";

function currencyFormatter(params) {
  return params.value.toLocaleString();
}

const columns = [
  { field: "name", flex: 1 },

  { field: "duration", cellClass: "text-right" },

  {
    field: "price",
    headerName: "Price (HUF)",
    valueFormatter: currencyFormatter,
    cellClass: "text-right",
  },

  {
    field: "actions",
    headerName: "",
    cellRenderer: () => {
      return <TableActions />;
    },
    resizable: false,
  },
];

export default function ServicesTable({ searchText, servicesData }) {
  return (
    <div className="h-full w-full">
      <AgGridReact
        theme={themeAlpine.withPart(colorSchemeDarkBlue)}
        quickFilterText={searchText}
        modules={[AllCommunityModule]}
        rowData={servicesData}
        autoSizeStrategy={{
          type: "fitCellContents",
          colIds: ["name", "duration", "price", "actions"],
        }}
        columnDefs={columns}
        defaultColDef={{ sortable: true }}
        rowSelection={{ mode: "singleRow" }}
      />
    </div>
  );
}
