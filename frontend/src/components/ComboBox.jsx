import { useState } from "react";
import SearchInput from "./SearchInput";
import Select from "./Select";

export default function ComboBox({
  options,
  value,
  onSelect,
  placeholder,
  styles,
  maxVisibleItems = 7,
  emptyText,
}) {
  const [searchText, setSearchText] = useState("");

  const filteredOptions = options?.filter((option) =>
    option.label.toLowerCase().includes(searchText.toLowerCase())
  );

  return (
    <Select
      options={filteredOptions}
      allOptions={options}
      value={value}
      onSelect={onSelect}
      placeholder={placeholder}
      styles={styles}
      maxVisibleItems={maxVisibleItems}
      extraContent={
        <SearchInput
          searchText={searchText}
          onChange={setSearchText}
          styles="border-t-0 border-x-0 border-b focus:border-b-gray-300 border-b-gray-300 py-3
            dark:border-b-gray-500 dark:focus:border-b-gray-500 ring-0!"
          autoFocus={true}
        />
      }
      onClose={() => setSearchText("")}
      emptyText={emptyText}
    />
  );
}
