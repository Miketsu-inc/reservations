import { Search01Icon } from "@hugeicons/core-free-icons";
import { Icon } from ".";
import Input from "./Input";

export default function SearchInput({
  searchText,
  onChange,
  styles,
  autoFocus,
}) {
  return (
    <div className="relative">
      <div
        className="pointer-events-none absolute inset-y-0 start-0 flex
          items-center ps-3"
      >
        <Icon icon={Search01Icon} styles="size-4" />
      </div>
      <Input
        styles={`ps-9 w-44 md:w-full text-sm ${styles}`}
        name="search"
        type="search"
        pattern=".{0,255}"
        value={searchText}
        required={false}
        placeholder="Search"
        inputData={(data) => {
          onChange(data.value);
        }}
        autoFocus={autoFocus}
      />
    </div>
  );
}
