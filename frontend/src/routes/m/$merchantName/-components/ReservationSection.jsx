export default function ReservationSection({ children, name, show }) {
  return (
    show && (
      <div>
        <p className="pb-2 text-lg font-bold lg:text-xl">{name}</p>
        {children}
      </div>
    )
  );
}
