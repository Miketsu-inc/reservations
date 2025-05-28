export default function Card({ styles, children }) {
  return (
    <div
      className={`${styles} border-border_color bg-layer_bg size-full rounded-lg border p-4
        shadow-sm`}
    >
      {children}
    </div>
  );
}
