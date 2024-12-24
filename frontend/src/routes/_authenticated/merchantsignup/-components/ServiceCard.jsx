import Button from "@components/Button";
import Input from "@components/Input";
import EditIcon from "@icons/EditIcon";
import TickIcon from "@icons/TickIcon";
import XIcon from "@icons/XIcon";
import { useState } from "react";

export default function ServiceCard({
  index,
  service,
  handleEdit,
  handleDelete,
  exists,
}) {
  const [tempData, setTempData] = useState(service);
  const [isEditing, setIsEditing] = useState(false);
  const [isEmpty, setIsEmpty] = useState(false);
  const [error, setError] = useState("");

  function handleChange(data) {
    setIsEmpty(false);
    setTempData((prevData) => ({ ...prevData, [data.name]: data.value }));
    setError("");
  }

  function handleCancel() {
    setTempData(service);
    setError("");
    setIsEditing(false);
  }

  function handleSave(e) {
    e.preventDefault();
    const form = e.target;
    if (!form.checkValidity()) {
      setIsEmpty(true);
      return;
    }
    if (exists(tempData.name, index)) {
      setError("This name is already used");
      return;
    }
    setIsEditing(false);
    handleEdit(index, tempData);
  }

  return (
    <div
      className="rounded-lg bg-slate-300/45 shadow-md dark:border-2 dark:border-gray-500
        dark:bg-hvr_gray"
    >
      {!isEditing ? (
        <>
          <div className="flex flex-row-reverse justify-between">
            <div className="flex gap-1">
              <div className="h-min p-1 hover:bg-gray-300/20">
                <EditIcon
                  styles="w-4 h-4 flex-shrink-0 cursor-pointer"
                  onClick={() => {
                    setIsEditing(true);
                  }}
                />
              </div>
              <XIcon
                onClick={() => {
                  handleDelete(index);
                }}
                styles="hover:bg-red-600/50 w-6 h-6 rounded-tr-lg flex-shrink-0 cursor-pointer"
              />
            </div>
            <h3 className="mb-6 truncate pl-5 pt-3 text-lg font-semibold text-text_color">
              {service.name}
            </h3>
          </div>
          <div className="flex items-center justify-between px-5 pb-3">
            <span className="text-sm tracking-tight text-gray-600 dark:text-gray-400">
              {service.duration} min
            </span>
            <span className="text-sm tracking-tight text-green-600 dark:text-green-400">
              {service.price} FT
            </span>
          </div>
        </>
      ) : (
        <form
          noValidate
          className="flex flex-col justify-center gap-2 px-4 py-3"
          onSubmit={handleSave}
        >
          {/* the border styles here don't apply, i think because in the component the border styles are applied conditionally*/}
          <Input
            id="card_name"
            styles={`focus:border-gray-400 border-b-2 border-x-0 border-t-0 border-gray-700
              bg-transparent pl-1 text-lg font-semibold text-text_color outline-none p-0`}
            type="text"
            value={tempData.name}
            name="name"
            inputData={handleChange}
            placeholder="Service type"
            pattern=".{0,255}"
            errorText="Please enter a valid name"
            errorStyles="text-xs"
            hasError={isEmpty}
          />

          <Input
            id="card_duration"
            styles="focus:border-gray-400 border-b-2 border-x-0 border-t-0 border-gray-700
              bg-transparent pl-1 text-sm tracking-tight text-gray-400 outline-none p-0"
            type="text"
            value={tempData.duration}
            name="duration"
            placeholder="Duration (min)"
            inputData={handleChange}
            pattern="^[0-9]{0,255}$"
            errorText="Please enter a valid duration"
            hasError={isEmpty}
          />

          <Input
            id="card_price"
            styles="border-b-2 border-x-0 border-t-0 border-gray-700 bg-transparent pl-1 text-sm
              tracking-tight text-green-600 outline-none focus:border-gray-400
              dark:text-green-400 p-0"
            type="text"
            placeholder="Price (FT)"
            name="price"
            value={tempData.price}
            inputData={handleChange}
            pattern="^[0-9]{0,255}$"
            errorText="Please enter a valid price"
            hasError={isEmpty}
          />

          <div className="mb-1 mt-3 flex items-center justify-end gap-3">
            <Button
              styles="text-xs bg-transparent rounded-md dark:border-gray-500 hover:border-gray-700
                border-[1px] py-[0.20rem] tracking-tight pr-2 border-gray-500"
              buttonText="Cancel"
              type="reset"
              onClick={handleCancel}
            >
              <XIcon styles="h-5 w-5" />
            </Button>
            <Button
              styles="text-xs rounded-md border-black dark:border-gray-300 border-[1px] py-[0.20rem]
                tracking-tight pr-2 bg-text_color text-white dark:text-text_color
                dark:bg-transparent hover:bg-text_color/85"
              buttonText="Save"
              type="submit"
            >
              <TickIcon styles="fill-white h-5 w-5" />
            </Button>
          </div>
          {error && (
            <span className="text-center text-sm text-red-500">{error}</span>
          )}
        </form>
      )}
    </div>
  );
}
