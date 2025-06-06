import Button from "@components/Button";
import SearchInput from "@components/SearchInput";
import PlusIcon from "@icons/PlusIcon";
import {
  CellStyleModule,
  ClientSideRowModelModule,
  ColumnApiModule,
  ColumnAutoSizeModule,
  QuickFilterModule,
  themeAlpine,
} from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";
import { useCallback, useRef, useState } from "react";

export default function Table({
  rowData,
  columnDef,
  columnsToAutoSize,
  itemName,
  onNewItem,
}) {
  const tableRef = useRef();
  const [searchText, setSearchText] = useState("");
  const [isLoading, setIsLoading] = useState(true);

  // you should only autosize columns which does not have a flex field
  // as with that the autosize will get applied instead of flex
  // making the table potentially not fill it's grid
  const resetView = useCallback(() => {
    tableRef.current.api.resetColumnState();
    tableRef.current.api.autoSizeColumns(columnsToAutoSize);
  }, [columnsToAutoSize]);

  return (
    <div
      className="md:bg-layer_bg md:border-border_color flex h-full w-full flex-1 flex-col
        md:rounded-lg md:border md:px-4 md:py-4 md:shadow-sm"
    >
      <div className="flex flex-col-reverse justify-between gap-2 pb-2 sm:flex-row sm:gap-0">
        <SearchInput
          searchText={searchText}
          onChange={(text) => setSearchText(text)}
        />
        <div className="flex flex-row justify-between sm:gap-3">
          <Button
            variant="primary"
            onClick={resetView}
            styles="p-2 text-sm w-fit"
            buttonText="Reset view"
          />
          <Button
            variant="primary"
            onClick={onNewItem}
            styles="p-2 text-sm w-fit"
            buttonText={`New ${itemName}`}
          >
            <PlusIcon styles="size-4 md:size-5 mr-1 text-white" />
          </Button>
        </div>
      </div>
      <div className={`${isLoading ? "invisible" : "visible"} h-full w-full`}>
        <AgGridReact
          ref={tableRef}
          theme={themeAlpine}
          quickFilterText={searchText}
          modules={[
            ClientSideRowModelModule,
            QuickFilterModule,
            CellStyleModule,
            ColumnApiModule,
            ColumnAutoSizeModule,
          ]}
          rowData={rowData}
          columnDefs={columnDef}
          defaultColDef={{ sortable: true, suppressMovable: true }}
          getRowId={(params) => String(params.data.id)}
          onFirstDataRendered={(params) => {
            params.api.autoSizeColumns(columnsToAutoSize);
          }}
          onGridReady={() => setIsLoading(false)}
          // suppressColumnVirtualisation is needed for autosizing to work on mobile
          // if disabled only columns in view will get autosized
          suppressColumnVirtualisation={true}
        />
      </div>
    </div>
  );
}
