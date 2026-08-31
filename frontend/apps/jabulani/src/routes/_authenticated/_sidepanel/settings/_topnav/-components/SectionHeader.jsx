export default function SectionHeader({ title, styles }) {
  return (
    <div className="flex flex-col gap-1">
      <div className={`${styles} text-2xl`}>{title}</div>
      <hr className="border-gray-300 dark:border-gray-600" />
    </div>
  );
}
