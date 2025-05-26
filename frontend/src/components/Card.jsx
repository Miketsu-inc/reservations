export default function Card({ styles, children, shadow = "shadow-sm" }) {
  return (
    <div
      className={`${styles} border-border_color bg-layer_bg size-full rounded-lg border p-4
        ${shadow}`}
    >
      {children}
    </div>
  );
}
