import { useParams } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import ServerError from "../../components/ServerError";
import ReservationSection from "./ReservationSection";
import ServiceItem from "./ServiceItem";

const defaultReservation = {
  merchant_name: "Hair salon",
  service_id: 0,
  location_id: 0,
  day: "",
  from_hour: "",
};

const defaultMerchantInfo = {
  merchantName: "",
  shortLocation: "",
  contact_email: "",
  shortDescription: "",
  parkingInfo: "",
  aboutUs: "",
  annoucement: "",
  services: [],
};

export default function ReservationPage() {
  const [reservation, setReservation] = useState(defaultReservation);
  const [merchantInfo, setMerchantInfo] = useState(defaultMerchantInfo);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [serverError, setServerError] = useState(undefined);
  const { merchantName } = useParams({ strict: false });

  useEffect(() => {
    async function fetchMerchantInfo() {
      try {
        const response = await fetch(
          `/api/v1/merchants/info?name=${merchantName}`,
          {
            method: "GET",
          }
        );

        const result = await response.json();

        if (!response.ok) {
          setServerError(result.error.message);
        } else {
          setServerError(undefined);

          setReservation({
            merchant_name: result.data.merchant_name,
            service_id: 1,
            location_id: result.data.location_id,
            from_date: "",
            to_date: "",
          });

          const shortLocation =
            result.data.address +
            ", " +
            result.data.city +
            " " +
            result.data.postal_code;

          setMerchantInfo({
            merchantName: result.data.merchant_name,
            contact_email: result.data.contact_email,
            shortLocation: shortLocation,
            services: result.data.services,
          });
        }
      } catch (err) {
        setServerError(err.message);
      }
    }

    fetchMerchantInfo();
  }, [merchantName]);

  useEffect(() => {
    if (isSubmitting) {
      const sendRequest = async () => {
        try {
          const response = await fetch("/api/v1/appointments", {
            method: "POST",
            headers: {
              "Content-type": "application/json; charset=UTF-8",
            },
            body: JSON.stringify(reservation),
          });

          if (!response.ok) {
            const result = await response.json();
            setServerError(result.error.message);
          }
        } catch (err) {
          setServerError(err.message);
        } finally {
          setIsSubmitting(false);
        }
      };

      sendRequest();
    }
  }, [isSubmitting, reservation]);

  function serviceClickHandler(id) {
    console.log(id);
  }

  return (
    <div className="mx-auto min-h-screen max-w-screen-xl bg-layer_bg px-10">
      <ServerError styles="my-4" error={serverError} />
      <div className="flex flex-col-reverse gap-4 py-5 lg:h-96 lg:flex-row lg:gap-14 lg:py-10">
        <div className="flex flex-col gap-6 md:flex-row lg:w-1/3 lg:flex-col">
          <div className="flex w-full flex-row">
            <div className="w-14 sm:w-20 lg:w-24">
              <img
                className="h-auto w-full rounded-3xl object-cover"
                src="https://dummyimage.com/200x200/d156c3/000000.jpg"
              />
            </div>
            <div className="flex flex-col justify-center pl-5 lg:gap-2">
              <h1 className="text-2xl font-bold lg:text-4xl">
                {merchantInfo.merchantName}
              </h1>
              <p className="text-sm">{merchantInfo.shortLocation}</p>
            </div>
          </div>
          <div className="flex w-full flex-col gap-2 md:items-end lg:items-start">
            <p className="text-justify">Hair stylist open on somedays</p>
            <p className="text-justify">An annoucement</p>
          </div>
        </div>
        <div className="h-40 sm:h-52 md:h-72 lg:h-full lg:max-h-full lg:w-2/3">
          <img
            className="h-full w-full rounded-2xl object-cover"
            src="https://dummyimage.com/1920x1080/d156c3/000000.jpg"
          ></img>
        </div>
      </div>
      <hr className="border-gray-500" />
      <div className="flex flex-col gap-10 pt-5 lg:flex-row lg:pt-10">
        <div className="lg:w-2/3">
          <p className="pb-5 text-lg font-bold">Services</p>
          {merchantInfo.services.map((service) => (
            <ServiceItem
              key={service.ID}
              id={service.ID}
              name={service.name}
              price={service.price}
              description="Exmaple description of an item in services. I'm really trying to test wether the length will cause any errors. Or how will it look. Ok it's time to stop yapping"
              serviceClick={serviceClickHandler}
            />
          ))}
        </div>
        <div className="flex flex-col gap-6 lg:w-1/3">
          <ReservationSection name="About us" show={true}>
            <p>
              Short description about the core values of the company, maybe also
              what they belive in. How they do their buisness. What they would
              like to achive in the future. I'm basically just bullshiting at
              this point. Should have used lorem ipsum
            </p>
          </ReservationSection>
          <ReservationSection name="Opening hours" show={true}>
            <div className="flex flex-col gap-2">
              <div className="grid grid-cols-3">
                <p>Monday</p>
                <p>Closed</p>
              </div>
              <div className="grid grid-cols-3">
                <p>Tuesday</p>
                <p>Closed</p>
              </div>
              <div className="grid grid-cols-3">
                <p>Wednesday</p>
                <p>Closed</p>
              </div>
              <div className="grid grid-cols-3">
                <p>Thursday</p>
                <p>16:00 - 22:00</p>
              </div>
              <div className="grid grid-cols-3">
                <p>Friday</p>
                <p>14:00 - 23:45</p>
              </div>
              <div className="grid grid-cols-3">
                <p>Saturday</p>
                <p>14:00 - 23:45</p>
              </div>
              <div className="grid grid-cols-3">
                <p>Sunday</p>
                <p>Closed</p>
              </div>
            </div>
          </ReservationSection>
          <ReservationSection name="Location" show={merchantInfo.shortLocation}>
            <p>{merchantInfo.shortLocation}</p>
          </ReservationSection>
          <ReservationSection name="Parking" show={merchantInfo.parkingInfo}>
            <p>{merchantInfo.parkingInfo}</p>
          </ReservationSection>
          <ReservationSection
            name="Contact us"
            show={merchantInfo.contact_email}
          >
            <p>Email: {merchantInfo.contact_email}</p>
            <p>Facebook: </p>
            <p>Instagram: </p>
            <p>Phone: </p>
          </ReservationSection>
        </div>
      </div>
    </div>
  );
}
