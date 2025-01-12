export default function TableColorPicker({ value }) {
  return (
    <input
      disabled={true}
      id="colorPicker"
      className="h-full bg-transparent"
      name="colorPicker"
      type="color"
      value={value}
    />
  );
}
