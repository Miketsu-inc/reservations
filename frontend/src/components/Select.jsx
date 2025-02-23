import DropdownBase from "./DropDownBase";

export default function Select({
  options,
  value,
  onSelect,
  placeholder,
  styles,
  maxVisibleItems = 7,
}) {
  return (
    <DropdownBase
      options={options}
      value={value}
      onSelect={onSelect}
      placeholder={placeholder}
      styles={styles}
      maxVisibleItems={maxVisibleItems}
    />
  );
}
