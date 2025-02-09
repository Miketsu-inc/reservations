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
        <span className="font-semibold">{title}</span>
        <div className="flex justify-center gap-10 sm:gap-32 md:justify-baseline md:gap-10">
          {options.map((option) => (
            <label
              key={option.value}
              htmlFor={option.value}
              className={`${
              value === option.value
                  ? "border-primary bg-primary/10"
                  : "border-gray-500 hover:border-gray-400"
              } flex items-center rounded-lg border-2 px-5 py-1`}
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
      {description && (
        <p className="text-text_color/70 text-sm">{description}</p>
      )}
    </>
  );
}
