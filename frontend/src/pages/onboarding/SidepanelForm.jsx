import XIcon from "../../assets/icons/XIcon";
import AppointmentForm from "./AppointmentForm";

export default function AppointmentSidepanel({
  isOpen,
  addAppointment,
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
      <div className="mt-4 flex w-full items-center justify-between border-b-2 border-gray-300 pb-2">
        <h2 className="text-lg font-semibold text-text_color">
          Add Appointment
        </h2>
        <XIcon
          onClick={() => {
            setIsAdding(false);
            setFormError("");
          }}
          styles="hover:bg-hvr_gray w-8 h-8 rounded-lg"
        />
      </div>
      {formError && (
        <div
          className="mt-4 flex items-start gap-2 rounded-md border-[1px] border-red-800 bg-red-600/25
            px-2 py-3 text-red-950 dark:border-red-800 dark:bg-red-700/15 dark:text-red-500"
        >
          {/* <ExclamationIcon styles="" /> */}
          <span className="pl-3">Error:</span> {formError}
        </div>
      )}
      {isOpen && <AppointmentForm sendInputData={addAppointment} />}
    </div>
  );
}
