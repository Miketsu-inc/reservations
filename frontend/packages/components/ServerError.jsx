export default function ServerError({ error, styles }) {
  return (
    <>
      {error && (
        <div
          role="alert"
          className={`${styles} border-border-danger bg-danger-subtle
          text-fg-danger flex items-start gap-2 rounded-lg border px-2 py-3
          text-sm`}
        >
          <span className="pl-3">Error:</span> {error}
        </div>
      )}
    </>
  );
}
