import { Download01Icon, PlusSignIcon } from "@hugeicons/core-free-icons";
import { useWindowSize } from "@reservations/lib";
import {
  CellStyleModule,
  ClientSideRowModelModule,
  ColumnApiModule,
  ColumnAutoSizeModule,
  CsvExportModule,
  QuickFilterModule,
  themeQuartz,
} from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";
import { useCallback, useRef, useState } from "react";
import { Icon } from ".";
import Button from "./Button";
import SearchInput from "./SearchInput";

const theme = themeQuartz.withParams({
  browserColorScheme: "inherit",
  fontFamily: "inherit",
  headerVerticalPaddingScale: 1,
  rowBorder: true,
  // wrapperBorder: false,
  borderColor: "rgb(var(--border-color))",
  backgroundColor: "rgb(var(--bg-color))",
  headerBackgroundColor: "rgb(var(--layer-bg))",
  accentColor: "rgb(var(--primary))",
});

export default function Table({
  rowData,
  columnDef,
  columnsToAutoSize,
  itemName,
  onNewItem,
  exportName = "export",
  onRowClick,
  columnsToExport,
  showControls = true,
  ...props
}) {
  const tableRef = useRef();
  const [searchText, setSearchText] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const windowSize = useWindowSize();

  const onBtnExport = useCallback(() => {
    tableRef.current.api.exportDataAsCsv({
      fileName: `${exportName}.csv`,
      columnKeys: columnsToExport,
    });
  }, [exportName, columnsToExport]);

  return (
    <div className="flex h-full w-full flex-1 flex-col">
      <div
        className="flex flex-col-reverse justify-between gap-2 pb-2 sm:flex-row
          sm:gap-0"
      >
        <div className="flex items-center justify-center gap-3 pt-2 md:pt-0">
          <div className="w-full md:w-auto">
            <SearchInput
              searchText={searchText}
              onChange={(text) => setSearchText(text)}
            />
          </div>
          <Button
            variant="tertiary"
            styles="p-2 text-sm w-fit text-nowrap"
            buttonText={windowSize != "sm" ? "Export" : ""}
            onClick={onBtnExport}
          >
            <Icon
              icon={Download01Icon}
              styles="text-text_color md:mr-2 md:mb-0.5 size-5"
            />
          </Button>
        </div>
        {showControls && (
          <div className="flex flex-row justify-between sm:gap-3">
            <Button
              variant="primary"
              onClick={onNewItem}
              styles="p-2 text-sm w-fit"
              buttonText={`New ${itemName}`}
            >
              <Icon
                icon={PlusSignIcon}
                styles="size-4 md:size-5 mr-1 text-white"
              />
            </Button>
          </div>
        )}
      </div>
      <div
        className={`${isLoading ? "invisible" : "visible"} h-full w-full
          ${rowData?.length === 0 ? "ag-header-hidden" : ""} `}
      >
        <AgGridReact
          ref={tableRef}
          theme={theme}
          quickFilterText={searchText}
          modules={[
            ClientSideRowModelModule,
            QuickFilterModule,
            CellStyleModule,
            ColumnApiModule,
            ColumnAutoSizeModule,
            CsvExportModule,
          ]}
          rowData={rowData}
          columnDefs={columnDef}
          defaultColDef={{ sortable: true, suppressMovable: true }}
          getRowId={(params) => String(params.data.id)}
          onFirstDataRendered={(params) => {
            params.api.autoSizeColumns(columnsToAutoSize || []);
          }}
          onGridReady={() => setIsLoading(false)}
          // suppressColumnVirtualisation is needed for autosizing to work on mobile
          // if disabled only columns in view will get autosized
          suppressColumnVirtualisation={true}
          onRowClicked={onRowClick}
          {...props}
        />
      </div>
    </div>
  );
}
