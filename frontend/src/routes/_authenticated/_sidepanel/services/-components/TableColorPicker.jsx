export default function TableColorPicker({ value, onChange }) {
  return (
    <input
      id="colorPicker"
      className="h-full cursor-pointer bg-transparent"
      name="colorPicker"
      type="color"
      value={value}
      onChange={onChange}
    />
  );
}
