export default function ReservationSection({ children, name, show }) {
  return (
    show && (
      <div className="">
        <h2 className="pb-6 text-xl font-semibold lg:text-2xl">{name}</h2>
        {children}
      </div>
    )
  );
}
