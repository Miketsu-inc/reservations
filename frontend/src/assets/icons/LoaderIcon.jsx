export default function LoaderIcon({ styles }) {
  return (
    <svg
      className={`${styles} animate-spin fill-transparent`}
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
    >
      <circle
        className="stroke-transparent opacity-25"
        cx="12"
        cy="12"
        r="10"
        strokeWidth="4"
      ></circle>
      <path
        className="fill-current"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      ></path>
    </svg>
  );
}
