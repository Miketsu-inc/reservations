import { Search01Icon } from "@hugeicons/core-free-icons";
import Icon from "./Icon.jsx";
import { StyledInputBase } from "./InputBase.jsx";

export default function SearchInput({
  searchText,
  onChange,
  styles,
  autoFocus,
  placeholder = "Search",
  ...props
}) {
  return (
    <div className="relative">
      <div
        className="pointer-events-none absolute inset-y-0 inset-s-0 flex
          items-center ps-3"
      >
        <Icon icon={Search01Icon} styles="size-4" />
      </div>
      <StyledInputBase
        styles={`ps-9 w-44 md:w-full text-sm ${styles}`}
        name="search"
        type="search"
        pattern=".{0,255}"
        maxLength={255}
        value={searchText}
        required={false}
        placeholder={placeholder}
        aria-label={props["aria-label"] || placeholder}
        onChange={(event) => onChange(event.target.value)}
        autoFocus={autoFocus}
        {...props}
      />
    </div>
  );
}
