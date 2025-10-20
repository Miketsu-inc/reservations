export default function ReservationSection({ children, name, show }) {
  return (
    show && (
      <div>
        <p className="pb-2 text-xl font-bold">{name}</p>
        {children}
      </div>
    )
  );
}
