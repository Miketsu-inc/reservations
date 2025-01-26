import Button from "@components/Button";
import SearchInput from "@components/SearchInput";
import ServerError from "@components/ServerError";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import ReservationSection from "./-components/ReservationSection";
import ServiceItem from "./-components/ServiceItem";

async function fetchMerchantInfo(name) {
  const response = await fetch(`/api/v1/merchants/info?name=${name}`, {
    method: "GET",
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

export const Route = createFileRoute("/m/$merchantName/")({
  component: MerchantPage,
  loader: async ({ params }) => {
    return fetchMerchantInfo(params.merchantName);
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

const defaultMerchantInfo = {
  merchant_name: "",
  location_id: 0,
  contact_email: "",
  short_location: "",
  introduction: "",
  annoucement: "",
  about_us: "",
  parking_info: "",
  payment_info: "",
  services: [],
};

function MerchantPage() {
  const [merchantInfo, setMerchantInfo] = useState(defaultMerchantInfo);
  const loaderData = Route.useLoaderData();
  const [searchText, setSearchText] = useState("");

  useEffect(() => {
    if (loaderData) {
      const shortLocation =
        loaderData.address +
        ", " +
        loaderData.city +
        " " +
        loaderData.postal_code;

      setMerchantInfo({
        merchant_name: loaderData.merchant_name,
        location_id: loaderData.location_id,
        contact_email: loaderData.contact_email,
        short_location: shortLocation,
        introduction: loaderData.introduction,
        annoucement: loaderData.annoucement,
        about_us: loaderData.about_us,
        parking_info: loaderData.parking_info,
        payment_info: loaderData.payment_info,
        services: loaderData.services,
      });
    }
  }, [loaderData]);

  const filteredServices = merchantInfo.services.filter(
    (service) =>
      service.name.toLowerCase().includes(searchText.toLowerCase()) ||
      service.description.toLowerCase().includes(searchText.toLowerCase())
  );

  return (
    <div className="mx-auto min-h-screen max-w-screen-xl bg-layer_bg px-8">
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
                {merchantInfo.merchant_name}
              </h1>
              <p className="text-sm">{merchantInfo.short_location}</p>
            </div>
          </div>
          <div className="flex w-full flex-col gap-2 md:items-end lg:items-start">
            <p className="text-justify">{merchantInfo.introduction}</p>
            <p className="text-justify">{merchantInfo.annoucement}</p>
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
          <p className="pb-8 text-lg font-bold">Services</p>
          <SearchInput
            searchText={searchText}
            onChange={(text) => setSearchText(text)}
          />
          {filteredServices.map((service) => (
            <ServiceItem
              key={service.id}
              id={service.id}
              name={service.name}
              price={service.price}
              description={service.description}
            >
              <Link
                from={Route.fullPath}
                to="booking"
                search={{
                  locationId: merchantInfo.location_id,
                  serviceId: service.id,
                  day: new Date().toISOString().split("T")[0],
                }}
              >
                <Button
                  variant="primary"
                  styles="p-2"
                  name="Reserve"
                  buttonText="Reserve"
                />
              </Link>
            </ServiceItem>
          ))}
        </div>
        <div className="flex flex-col gap-6 lg:w-1/3">
          <ReservationSection name="About us" show={merchantInfo.about_us}>
            <p>{merchantInfo.about_us}</p>
          </ReservationSection>
          <ReservationSection name="Opening hours" show={true}>
            <div className="flex flex-col gap-2 *:grid *:grid-cols-3">
              <div>
                <p>Monday</p>
                <p>Closed</p>
              </div>
              <div>
                <p>Tuesday</p>
                <p>Closed</p>
              </div>
              <div>
                <p>Wednesday</p>
                <p>Closed</p>
              </div>
              <div>
                <p>Thursday</p>
                <p>16:00 - 22:00</p>
              </div>
              <div>
                <p>Friday</p>
                <p>14:00 - 23:45</p>
              </div>
              <div>
                <p>Saturday</p>
                <p>14:00 - 23:45</p>
              </div>
              <div>
                <p>Sunday</p>
                <p>Closed</p>
              </div>
            </div>
          </ReservationSection>
          <ReservationSection name="Payment" show={merchantInfo.payment_info}>
            <p>{merchantInfo.payment_info}</p>
          </ReservationSection>
          <ReservationSection
            name="Location"
            show={merchantInfo.short_location}
          >
            <p>{merchantInfo.short_location}</p>
          </ReservationSection>
          <ReservationSection name="Parking" show={merchantInfo.parking_info}>
            <p>{merchantInfo.parking_info}</p>
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
