export default function RadioInputGroup({
  title,
  value,
  name,
  options,
  onChange,
  description,
}) {
  return (
    <>
      <div className="flex flex-col gap-4 lg:flex-row lg:gap-20">
        <span className="text-sm">{title}</span>
        <div
          className="flex justify-center gap-10 sm:gap-32 md:justify-baseline
            md:gap-10"
        >
          {options.map((option) => (
            <label
              key={option.value}
              htmlFor={option.value}
              className={`${
                value === option.value
                  ? "border-primary/60 bg-primary-subtle text-fg-primary "
                  : `border-border-strong bg-bg-input text-fg-muted
                    hover:border-primary/60 hover:text-fg`
              } flex cursor-pointer items-center rounded-lg border px-4 py-1.5
              text-sm`}
            >
              <input
                className="hidden"
                type="radio"
                name={name}
                id={option.value}
                value={option.value}
                checked={value === option.value}
                onChange={() => onChange(option.value)}
              />
              <span>{option.label}</span>
            </label>
          ))}
        </div>
      </div>
      {description && <p className="text-fg-muted text-sm">{description}</p>}
    </>
  );
}
