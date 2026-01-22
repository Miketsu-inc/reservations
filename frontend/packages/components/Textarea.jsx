export default function Textarea({
  id,
  name,
  styles,
  labelText,
  inputData,
  value,
  required,
  ...props
}) {
  function handleChange(e) {
    inputData({
      name: name,
      value: e.target.value,
    });
  }

  return (
    <label htmlFor={id} className="flex w-full flex-1 flex-col">
      {labelText && (
        <span className="flex items-center gap-1 pb-1 text-sm">
          {labelText}
          {required !== false && (
            <span className="text-base leading-none text-red-500">*</span>
          )}
        </span>
      )}
      <textarea
        id={id}
        name={name}
        className={`${styles} bg-layer_bg border-input_border_color
          focus:border-primary focus:ring-primary/30 disabled:text-text_color/70
          disabled:border-input_border_color/60 resize-none rounded-lg border
          placeholder-stone-500 outline-hidden
          transition-[border-color,box-shadow] duration-150 ease-in-out
          focus:ring-4 disabled:bg-gray-200/60 dark:placeholder-zinc-400
          dark:scheme-dark disabled:dark:bg-gray-700/20`}
        onChange={handleChange}
        required={required === undefined ? true : required}
        value={value}
        {...props}
      />
    </label>
  );
}
