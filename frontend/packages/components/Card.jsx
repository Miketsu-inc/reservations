export default function Card({ styles, children }) {
  return (
    <div
      className={`${styles} border-border bg-bg-raised size-full rounded-lg
        border p-4 shadow-sm`}
    >
      {children}
    </div>
  );
}
