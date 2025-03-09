import DropdownBase from "./DropdownBase";

export default function Select({
  options,
  value,
  onSelect,
  placeholder,
  styles,
  maxVisibleItems = 7,
  emptyText,
}) {
  return (
    <DropdownBase
      options={options}
      value={value}
      onSelect={onSelect}
      placeholder={placeholder}
      styles={styles}
      maxVisibleItems={maxVisibleItems}
      emptyText={emptyText}
    />
  );
}
