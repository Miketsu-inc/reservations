export default function BackArrowIcon({ styles }) {
  return (
    <svg
      className={`${styles} h-6 w-6 fill-none stroke-gray-500`}
      aria-hidden="true"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
        d="m15 19-7-7 7-7"
      />
    </svg>
  );
}
