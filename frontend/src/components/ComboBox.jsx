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
  onOpenChange,
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
          autoFocus={true}
        />
      }
      onClose={() => setSearchText("")}
      emptyText={emptyText}
      onOpenChange={onOpenChange}
    />
  );
}
