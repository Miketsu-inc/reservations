export default function SettingsPageHeader({ title, description, action }) {
  return (
    <header
      className={`flex flex-col justify-between gap-4 sm:flex-row
        sm:items-start pt-5 ${
          action
            ? `md:bg-bg_color md:sticky md:top-0 md:z-20 md:-mx-2 md:px-2
              md:pt-3 md:pb-3`
            : ""
        }`}
    >
      <div className="min-w-0">
        <h1 className="text-text_color text-2xl font-semibold tracking-tight">
          {title}
        </h1>
        {description && (
          <p className="text-text_color/65 mt-1 max-w-2xl text-sm leading-6">
            {description}
          </p>
        )}
      </div>
      {action}
    </header>
  );
}
