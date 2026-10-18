export default function Card({ styles, children }) {
  return (
    <div
      className={`${styles} border-border bg-bg-raised shadow-card size-full
        rounded-xl border p-4`}
    >
      {children}
    </div>
  );
}
