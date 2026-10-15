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
        <span className="flex items-center gap-1 pb-1 text-sm">
          {labelText}
          {required !== false && (
            <span className="text-fg-danger text-base leading-none">*</span>
          )}
        </span>
      )}
      <textarea
        id={id}
        name={name}
        className={`${styles} bg-bg-raised border-border-strong
          focus:border-primary focus:ring-primary-focus disabled:text-fg-muted
          disabled:border-border disabled:bg-bg-sunken resize-none rounded-lg
          border outline-hidden transition-[border-color,box-shadow]
          duration-150 ease-in-out focus:ring-4 dark:scheme-dark`}
        onChange={handleChange}
        required={required === undefined ? true : required}
        value={value}
        {...props}
      />
    </label>
  );
}
