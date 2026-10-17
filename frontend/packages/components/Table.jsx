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
import Button from "./Button";
import Icon from "./Icon.jsx";
import SearchInput from "./SearchInput";

const theme = themeQuartz.withParams({
  browserColorScheme: "inherit",
  fontFamily: "inherit",
  headerVerticalPaddingScale: 1,
  rowBorder: true,
  // wrapperBorder: false,
  borderColor: "var(--border)",
  backgroundColor: "var(--bg)",
  foregroundColor: "var(--fg)",
  headerBackgroundColor: "var(--bg-raised)",
  headerTextColor: "var(--fg-muted)",
  headerFontSize: 13,
  rowHoverColor: "var(--hover)",
  accentColor: "var(--primary)",
  wrapperBorderRadius: 12,
});

const modules = [
  ClientSideRowModelModule,
  QuickFilterModule,
  CellStyleModule,
  ColumnApiModule,
  ColumnAutoSizeModule,
  CsvExportModule,
];

const defaultColDef = { sortable: true, suppressMovable: true };

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
  const { windowSize } = useWindowSize();

  const onBtnExport = useCallback(() => {
    tableRef.current?.api?.exportDataAsCsv({
      fileName: `${exportName}.csv`,
      columnKeys: columnsToExport,
    });
  }, [exportName, columnsToExport]);

  return (
    <div className="flex h-full min-h-0 w-full min-w-0 flex-1 flex-col">
      <div
        className="flex shrink-0 flex-col-reverse justify-between gap-2 pb-3
          sm:flex-row sm:gap-0"
      >
        <div className="flex items-center justify-center gap-2 pt-2 md:pt-0">
          <div className="w-full md:w-auto">
            <SearchInput
              searchText={searchText}
              onChange={(text) => setSearchText(text)}
            />
          </div>
          <Button
            variant="tertiary"
            styles="p-2 md:py-2 md:pr-3 md:pl-2.5 text-sm w-fit text-nowrap"
            buttonText={windowSize !== "sm" ? "Export" : ""}
            onClick={onBtnExport}
          >
            <Icon
              icon={Download01Icon}
              styles="text-fg-muted md:mr-1.5 size-4"
            />
          </Button>
        </div>
        {showControls && (
          <div className="flex flex-row justify-between sm:gap-3">
            <Button
              variant="primary"
              onClick={onNewItem}
              styles="py-2 pr-3 pl-2.5 text-sm w-fit"
              buttonText={`New ${itemName}`}
            >
              <Icon icon={PlusSignIcon} styles="size-4 mr-1.5" />
            </Button>
          </div>
        )}
      </div>
      <div
        className={`${isLoading ? "invisible" : "visible"} flex min-h-0 w-full
          min-w-0 flex-1 ${rowData?.length === 0 ? "ag-header-hidden" : ""} `}
      >
        <AgGridReact
          ref={tableRef}
          theme={theme}
          quickFilterText={searchText}
          modules={modules}
          rowData={rowData}
          columnDefs={columnDef}
          defaultColDef={defaultColDef}
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
