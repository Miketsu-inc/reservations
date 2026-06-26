import { useMemo, useState } from "react";
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
  labelText,
  ...props
}) {
  const [searchText, setSearchText] = useState("");

  const filteredOptions = useMemo(() => {
    if (!options) return [];
    if (!searchText) return options;

    const lowerSearch = searchText.toLowerCase();
    return options?.filter((option) =>
      option.label.toLowerCase().includes(lowerSearch)
    );
  }, [options, searchText]);

  return (
    <Select
      options={filteredOptions}
      allOptions={options}
      value={value}
      onSelect={onSelect}
      placeholder={placeholder}
      labelText={labelText}
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
      {...props}
    />
  );
}
