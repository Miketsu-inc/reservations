export default function Button({ children, name, type, styles, onClick }) {
  return (
    <button
      onClick={onClick}
      className={`${styles} rounded-lg bg-primary py-2 font-medium shadow-md hover:bg-customhvr1`}
      name={name}
      type={type}
    >
      {children}
    </button>
  );
}
