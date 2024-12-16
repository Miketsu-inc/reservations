import ServerError from "@components/ServerError";
import XIcon from "@icons/XIcon";
import AppointmentForm from "./AppointmentForm";

export default function SidepanelForm({
  isOpen,
  addService,
  setIsAdding,
  formError,
  setFormError,
}) {
  return (
    <div
      className={`fixed inset-y-0 right-0 w-full bg-layer_bg px-6 shadow-lg duration-300
        ease-in-out sm:w-96 lg:w-[28rem] xl:w-[32rem]
        ${isOpen ? "translate-x-0" : "translate-x-full"}`}
    >
      <div
        className="mt-4 flex w-full items-center justify-between border-b-2 border-text_color/50
          pb-2"
      >
        <h2 className="text-lg font-semibold text-text_color">Add Service</h2>
        <XIcon
          onClick={() => {
            setIsAdding(false);
            setFormError("");
          }}
          styles="hover:bg-hvr_gray w-8 h-8 rounded-lg"
        />
      </div>
      <ServerError styles="mt-4" error={formError} />
      {isOpen && <AppointmentForm sendInputData={addService} />}
    </div>
  );
}
