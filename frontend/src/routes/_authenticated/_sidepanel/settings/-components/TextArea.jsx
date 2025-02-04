export default function TextArea({
  styles,
  id,
  placeholder,
  name,
  description,
  label,
  sendInputData,
  value,
}) {
  function handleChange(e) {
    sendInputData({
      name: name,
      value: e.target.value,
    });
  }

  return (
    <div className="flex w-full flex-col">
      <label htmlFor={id} className="flex flex-col gap-2 font-semibold">
        {label}
        <textarea
          className={`${styles} w-full overflow-auto rounded-lg border border-gray-400 bg-hvr_gray/50
            px-3 py-2 font-normal outline-none focus:border-2 focus:border-primary
            focus:bg-transparent md:resize dark:[color-scheme:dark]`}
          name={name}
          id={id}
          placeholder={placeholder}
          value={value}
          onChange={handleChange}
        />
      </label>
      {description && (
        <span className="mt-2 text-sm text-text_color/70">{description}</span>
      )}
    </div>
  );
}
