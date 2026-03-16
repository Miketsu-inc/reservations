export default function ({ styles }) {
  return (
    <svg
      className={`${styles} `}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={2}
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M13.5 4.5L11 13l-4-2-2 5m8-3l4 4m0-11a2 2 0 11-4 0 2 2 0 014 0z"
      />
    </svg>
  );
}
