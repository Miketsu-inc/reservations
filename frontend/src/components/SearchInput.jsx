import Input from "@components/Input";
import SearchIcon from "@icons/SearchIcon";

export default function SearchInput({ searchText, onChange }) {
  return (
    <div className="relative">
      <div className="pointer-events-none absolute inset-y-0 start-0 flex items-center ps-3">
        <SearchIcon styles="!h-4 !w-4" />
      </div>
      <Input
        styles="p-2 !ps-9 w-44 md:w-full text-sm"
        name="search"
        type="search"
        pattern=".{0,255}"
        value={searchText}
        required={false}
        placeholder="Search"
        inputData={(data) => {
          onChange(data.value);
        }}
        hasError={false}
      />
    </div>
  );
}
