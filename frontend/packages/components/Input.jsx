import {
  defaultCountries,
  FlagImage,
  usePhoneInput,
} from "react-international-phone";
import { ComboBox } from ".";
import InputBase from "./InputBase";

const countryOptions = defaultCountries.map(([name, iso2, dialCode]) => ({
  value: iso2,
  label: `${name} (+${dialCode})`,
  icon: <FlagImage iso2={iso2} size="24px" />,
}));

// TODO: maybe move this over to InputBase and decide what to do with the FloatingLabelInput
const sharedStyles = `
  peer border bg-layer_bg outline-hidden placeholder-stone-500
  dark:placeholder-zinc-400 transition-[border-color,box-shadow]
  ease-in-out duration-150 border-input_border_color
  focus:border-primary focus:ring-4 focus:ring-primary/30
  disabled:text-text_color/70 disabled:border-input_border_color/60
  disabled:bg-gray-200/60 disabled:dark:bg-gray-700/20
`;

export default function Input({ type, ...props }) {
  if (type === "tel") {
    return <PhoneInput {...props} />;
  }

  return <StandardInput type={type} {...props} />;
}

function StandardInput({
  id,
  name,
  styles,
  labelText,
  inputData,
  value,
  required,
  children,
  childrenSide = "right",
  ...props
}) {
  function handleChange(e) {
    inputData({
      name: name,
      value: e.target.value,
    });
  }

  return (
    <>
      <LabelWrapper id={id} labelText={labelText} required={required}>
        <div
          className={`${childrenSide !== "right" ? "flex-row-reverse" : "flex-row"}
            flex items-center`}
        >
          <InputBase
            styles={`${styles} ${
              children &&
              (childrenSide === "right"
                ? "border-r-0 rounded-r-none"
                : "border-l-0 rounded-l-none")
              } ${sharedStyles}`}
            id={id}
            name={name}
            onChange={handleChange}
            required={required === undefined ? true : required}
            onBlur={() => {}}
            value={value}
            {...props}
          />
          {children}
        </div>
      </LabelWrapper>
    </>
  );
}

function PhoneInput({
  id,
  name,
  styles,
  labelText,
  value,
  required,
  inputData,
  ...props
}) {
  const { inputValue, handlePhoneValueChange, country, setCountry } =
    usePhoneInput({
      defaultCountry: "hu",
      value,
      onChange: (data) => {
        inputData({
          name,
          value: data.phone, // E.164 format: +36301234567
        });
      },
    });

  return (
    <LabelWrapper id={id} labelText={labelText} required={required}>
      <div className={`${styles} flex w-full items-center`}>
        <ComboBox
          styles="w-fit! border-r-0 rounded-r-none"
          value={country.iso2}
          options={countryOptions}
          onSelect={(option) => {
            setCountry(option.value);
          }}
          dropDownSameWidth={false}
          showOnlyIcon={true}
        />
        <InputBase
          styles={`${sharedStyles} flex-1 rounded-l-none p-2`}
          id={id}
          name={name}
          type="tel"
          value={inputValue}
          onChange={handlePhoneValueChange}
          required={required === undefined ? true : required}
          onBlur={() => {}}
          {...props}
        />
      </div>
    </LabelWrapper>
  );
}

function LabelWrapper({ id, labelText, required, children }) {
  return (
    <label htmlFor={id} className="flex w-full flex-col">
      {labelText && (
        <span className="flex items-center gap-1 pb-1 text-sm">
          {labelText}
          {required !== false && (
            <span className="text-base leading-none text-red-500">*</span>
          )}
        </span>
      )}
      {children}
    </label>
  );
}
