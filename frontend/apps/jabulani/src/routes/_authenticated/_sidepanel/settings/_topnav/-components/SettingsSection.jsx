export default function SettingsSection({
  title,
  description,
  children,
  styles,
}) {
  return (
    <section className={styles}>
      {(title || description) && (
        <div className="mb-5">
          {title && <h2 className="text-fg font-semibold">{title}</h2>}
          {description && (
            <p className="text-fg-muted mt-0.5 text-sm">{description}</p>
          )}
        </div>
      )}
      {children}
    </section>
  );
}
