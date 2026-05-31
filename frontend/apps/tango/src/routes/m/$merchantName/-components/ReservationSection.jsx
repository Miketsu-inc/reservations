export default function ReservationSection({ children, name, show }) {
  return (
    show && (
      <div className="">
        <p className="pb-6 text-xl font-semibold lg:text-2xl">{name}</p>
        {children}
      </div>
    )
  );
}
