import InputBase from "./InputBase";

export default function Input({
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
      <label htmlFor={id} className="flex w-full flex-col">
        {labelText && <span className="pb-1 text-sm">{labelText}</span>}
        <div
          className={`${childrenSide !== "right" ? "flex-row-reverse" : "flex-row"} flex items-center`}
        >
          <InputBase
            styles={`${styles} ${ children &&
              (childrenSide === "right"
                ? "border-r-0 rounded-r-none"
                : "border-l-0 rounded-l-none")
              } peer border bg-layer_bg outline-hidden placeholder-stone-500
              dark:placeholder-zinc-400 transition-[border-color,box-shadow] ease-in-out
              duration-150 border-input_border_color focus:border-primary focus:ring-4
              focus:ring-primary/30`}
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
      </label>
    </>
  );
}
