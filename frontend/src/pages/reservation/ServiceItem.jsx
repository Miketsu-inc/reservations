import Button from "../../components/Button";

export default function ServiceItem({
  id,
  name,
  description,
  price,
  serviceClick,
}) {
  function onClickHandler() {
    serviceClick(id);
  }

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
        <div className="ml-auto text-right">
          <Button
            styles="p-4"
            name="Reserve"
            buttonText="Reserve"
            onClick={onClickHandler}
          />
        </div>
      </div>
      <hr className="border-gray-500" />
    </>
  );
}
