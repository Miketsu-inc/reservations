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
    inputData?.({
      name: name,
      value: e.target.value,
    });
  }

  return (
    <label htmlFor={id} className="flex w-full flex-1 flex-col">
      {labelText && (
        <span className="text-fg flex items-center gap-1 pb-1.5 text-sm">
          {labelText}
          {required !== false && (
            <span className="text-fg-danger text-base leading-none">*</span>
          )}
        </span>
      )}
      <textarea
        id={id}
        name={name}
        className={`${styles} bg-bg-input border-border-strong
          hover:not-disabled:not-focus:border-fg-muted focus:border-primary
          focus:ring-primary-focus disabled:text-fg-disabled
          disabled:border-border disabled:bg-bg-sunken resize-none rounded-lg
          border outline-hidden transition-[border-color,box-shadow]
          duration-150 ease-in-out focus:ring-3 md:text-sm dark:scheme-dark`}
        onChange={handleChange}
        required={required === undefined ? true : required}
        value={value}
        {...props}
      />
    </label>
  );
}
