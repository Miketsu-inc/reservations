import { useEffect, useState } from "react";
import XIcon from "../../assets/icons/XIcon";
import Button from "../../components/Button";
import AppointmentSidepanel from "./SidepanelForm";

export default function AppointmentsAdder() {
  const [apps, setApps] = useState([]);
  const [isAdding, setIsAdding] = useState(false);
  const [fromError, setFormError] = useState("");
  const [submitError, setSubmitError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  useEffect(() => {
    if (isSubmitting) {
      const sendRequest = async () => {
        try {
          const response = await fetch("/api/v1/auth/merchant/appointments", {
            method: "POST",
            headers: {
              Accept: "application/json",
              "content-type": "application/json",
            },
            body: JSON.stringify(apps),
          });
          const result = await response.json();
          if (result.error) {
            return;
          }
        } catch (err) {
          console.error("Error messsage from server:", err.message);
        } finally {
          setIsSubmitting(false);
        }
      };
      sendRequest();
    }
  }, [apps, isSubmitting]);

  function handleSubmit() {
    if (apps.length === 0) {
      setSubmitError("Please make at least one appointment");
      return;
    }
    setIsSubmitting(true);
  }

  function addAppointment(newAppointemnt) {
    setApps((prevAppointments) => {
      const exists = prevAppointments.some(
        (appointment) => appointment.name === newAppointemnt.name
      );
      if (exists) {
        setFormError("You cant add appointments with the same name");
        return prevAppointments;
      }
      setIsAdding(false);
      setFormError("");
      return [...prevAppointments, newAppointemnt];
    });
  }

  function deleteAppointment(deleteIndex) {
    setApps((prevApps) => prevApps.filter((_, index) => index !== deleteIndex));
  }

  return (
    <>
      <div className="relative">
        <div
          className={`p-6 transition-all duration-300 ${isAdding ? "sm:mr-96 lg:pr-20 xl:pr-40" : ""}`}
        >
          <h1 className="text-3xl">Your Appointments</h1>
          <div
            className={`mt-6 grid w-full grid-cols-1 gap-6
              ${isAdding ? "sm:grid-cols-1 md:grid-cols-2 xl:grid-cols-3" : "sm:grid-cols-3 xl:grid-cols-4"}`}
          >
            {apps.map((appointment, index) => (
              <div
                key={index}
                className="rounded-lg bg-slate-200/70 shadow-md dark:border-[1px] dark:border-gray-600
                  dark:bg-hvr_gray"
              >
                <div className="flex flex-row-reverse justify-between">
                  <XIcon
                    onClick={() => deleteAppointment(index)}
                    styles="hover:bg-red-600/50 w-6 h-6 rounded-tr-lg flex-shrink-0"
                  />
                  <h3 className="mb-6 truncate pl-5 pt-3 text-lg font-semibold text-text_color">
                    {appointment.name}
                  </h3>
                </div>
                <div className="flex items-center justify-between px-5 pb-3">
                  <span className="text-sm tracking-tight text-gray-600 dark:text-gray-400">
                    {appointment.duration} min
                  </span>
                  <span className="text-sm tracking-tight text-green-600 dark:text-green-400">
                    {appointment.price} FT
                  </span>
                </div>
              </div>
            ))}

            {/* Add New Appointment Button */}
            <button
              className="flex h-auto flex-col items-center justify-center gap-2 rounded-lg
                bg-slate-200/70 p-3 hover:bg-slate-300/45 hover:shadow-lg dark:bg-hvr_gray
                dark:hover:bg-gray-700"
              onClick={() => {
                setSubmitError("");
                setIsAdding(true);
              }}
            >
              <div className="h-12 w-12 rounded-full bg-slate-300/45 p-3 dark:bg-gray-700">
                <span className="text-text_color">+</span>
              </div>
              <span className="text-sm font-medium dark:text-gray-300">
                Add Appointment
              </span>
            </button>
          </div>
          <p className="mt-4 text-center text-sm">
            You can also add and remove appointments later
          </p>
          <div className="mt-4 flex w-full flex-col items-center justify-center">
            <Button
              type="submit"
              styles="p-2 w-5/6 mt-10 font-semibold focus-visible:outline-1 bg-primary
                hover:bg-hvr_primary text-white"
              onClick={handleSubmit}
              buttonText="Save Appointments"
              isLoading={isSubmitting}
            />
            {submitError && (
              <span className="mt-4 text-red-600">{submitError}</span>
            )}
          </div>
        </div>
      </div>
      <AppointmentSidepanel
        addAppointment={addAppointment}
        setIsAdding={setIsAdding}
        isOpen={isAdding}
        formError={fromError}
        setFormError={setFormError}
      />
    </>
  );
}
