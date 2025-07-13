export default function ServerError({ error, styles }) {
  return (
    <>
      {error && (
        <div
          className={`${styles} flex items-start gap-2 rounded-md border border-red-800
          bg-red-600/25 px-2 py-3 text-red-950 dark:border-red-800 dark:bg-red-700/15
          dark:text-red-500`}
        >
          <span className="pl-3">Error:</span> {error}
        </div>
      )}
    </>
  );
}
