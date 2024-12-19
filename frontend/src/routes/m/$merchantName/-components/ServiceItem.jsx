export default function ServiceItem({ children, name, description, price }) {
  return (
    <>
      <div className="flex w-full flex-row items-center gap-5 py-4">
        <div className="flex flex-col">
          <p className="text-lg">{name}</p>
          <p className="hidden text-sm sm:block">{description}</p>
        </div>
        <p className="ml-auto w-full text-right">
          {parseFloat(price).toLocaleString()} HUF
        </p>
        <div className="ml-auto text-right">{children}</div>
      </div>
      <hr className="border-gray-500" />
    </>
  );
}
