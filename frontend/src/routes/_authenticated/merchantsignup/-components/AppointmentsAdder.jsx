import Button from "@components/Button";
import ServerError from "@components/ServerError";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { useState } from "react";
import ServiceCard from "./ServiceCard";
import SidepanelForm from "./SidepanelForm";

export default function AppointmentsAdder({ redirect }) {
  const [services, setServices] = useState([]);
  const [isAdding, setIsAdding] = useState(false);
  const [fromError, setFormError] = useState();
  const [submitError, setSubmitError] = useState();
  const [isLoading, setIsLoading] = useState(false);

  async function handleSubmit() {
    if (services.length === 0) {
      setSubmitError("Please make at least one appointment");
      return;
    }

    setIsLoading(true);
    try {
      const response = await fetch("/api/v1/merchants/services", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(services),
      });

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setSubmitError(result.error.message);
      } else {
        redirect();
      }
    } catch (err) {
      setSubmitError(err.message);
    } finally {
      setIsLoading(false);
    }
  }

  function addService(newService) {
    setServices((prevServices) => {
      const exists = prevServices.some(
        (service) => service.name === newService.name
      );

      if (exists) {
        setFormError("You cant add appointments with the same name");
        return prevServices;
      }

      setIsAdding(false);
      setFormError("");

      return [...prevServices, newService];
    });
  }

  function deleteService(deleteIndex) {
    setServices((prevServices) =>
      prevServices.filter((_, index) => index !== deleteIndex)
    );
  }

  function handleEdit(index, newData) {
    setServices(
      services.map((service, i) =>
        i === index ? { ...service, ...newData } : service
      )
    );
  }

  return (
    <>
      <div className="relative">
        <div
          className={`transition-all duration-300 ${isAdding ? "sm:mr-96 lg:pr-20 xl:pr-40" : ""}`}
        >
          <ServerError error={submitError} styles="mb-4" />
          <h1 className="mb-14 mt-4 text-2xl">Your Services</h1>
          <div
            className={`mt-6 grid w-full grid-cols-1 gap-6
              ${isAdding ? "sm:grid-cols-1 md:grid-cols-2 xl:grid-cols-3" : "sm:grid-cols-3 xl:grid-cols-4"}`}
          >
            {services.map((service, index) => (
              <ServiceCard
                key={index}
                service={service}
                index={index}
                handleDelete={deleteService}
                handleEdit={handleEdit}
                exists={(newName, index) => {
                  return services.some(
                    (service, id) => service.name === newName && id !== index
                  );
                }}
              />
            ))}

            <button
              className="flex h-auto flex-col items-center justify-center gap-2 rounded-lg
                bg-slate-300/45 p-3 hover:bg-slate-300 hover:shadow-lg dark:bg-hvr_gray
                dark:hover:bg-gray-700"
              onClick={() => {
                setSubmitError("");
                setIsAdding(true);
              }}
            >
              <div className="h-12 w-12 rounded-full bg-slate-300 p-3 dark:bg-gray-700">
                <span className="text-text_color">+</span>
              </div>
              <span className="text-sm font-medium dark:text-gray-300">
                Add Service
              </span>
            </button>
          </div>
          <p className="mt-5 text-center text-sm">
            You can also add, remove or edit services later
          </p>
          <div className="mt-4 flex w-full flex-col items-center justify-center">
            <Button
              type="submit"
              styles="p-2 w-full mt-10 font-semibold focus-visible:outline-1 bg-primary
                hover:bg-hvr_primary text-white"
              onClick={handleSubmit}
              buttonText="Save Services"
              isLoading={isLoading}
            />
          </div>
        </div>
      </div>
      <SidepanelForm
        addService={addService}
        setIsAdding={setIsAdding}
        isOpen={isAdding}
        formError={fromError}
        setFormError={setFormError}
      />
    </>
  );
}
