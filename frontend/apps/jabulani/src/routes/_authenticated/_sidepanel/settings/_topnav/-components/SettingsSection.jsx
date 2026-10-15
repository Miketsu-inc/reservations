export default function SettingsSection({
  title,
  description,
  children,
  styles,
}) {
  return (
    <section className={styles}>
      {(title || description) && (
        <div className="mb-8">
          {title && <h2 className="text-fg text-xl font-semibold">{title}</h2>}
          {description && (
            <p className="text-fg-muted mt-1 text-sm leading-6">
              {description}
            </p>
          )}
        </div>
      )}
      {children}
    </section>
  );
}
