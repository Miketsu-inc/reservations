import Button from "@components/Button";
import Card from "@components/Card";
import SearchInput from "@components/SearchInput";
import ServerError from "@components/ServerError";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import ReservationSection from "./-components/ReservationSection";
import ServiceCategoryItem from "./-components/ServiceCategoryItem";
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
  announcement: "",
  about_us: "",
  parking_info: "",
  payment_info: "",
  services: [],
  business_hours: { 1: [], 2: [], 3: [], 4: [], 5: [], 6: [], 0: [] },
};

const days = [
  "Sunday",
  "Monday",
  "Thuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

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
        announcement: loaderData.announcement,
        about_us: loaderData.about_us,
        parking_info: loaderData.parking_info,
        payment_info: loaderData.payment_info,
        services: loaderData.services,
        business_hours: loaderData.business_hours,
      });
    }
  }, [loaderData]);

  const filteredServicesGroupedByCategories = merchantInfo.services.map(
    (category) => ({
      ...category,
      services: category.services.filter((service) =>
        service.name.toLowerCase().includes(searchText.toLowerCase())
      ),
    })
  );

  return (
    <Card styles="mx-auto min-h-screen max-w-7xl md:p-6">
      <div className="mb-5 flex flex-col-reverse gap-4 lg:mb-0 lg:h-80 lg:flex-row lg:gap-14">
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
            <p className="text-justify">{merchantInfo.announcement}</p>
          </div>
        </div>
        <div className="h-40 overflow-hidden rounded-2xl md:h-72 lg:size-10/12">
          <img
            className="size-full object-cover"
            src="https://dummyimage.com/1920x1080/d156c3/000000.jpg"
          ></img>
        </div>
      </div>
      <hr className="border-border_color border" />
      <div className="flex flex-col gap-10 pt-5 lg:flex-row lg:pt-10">
        <div className="lg:w-2/3">
          <p className="pb-8 text-lg font-bold">Services</p>
          <SearchInput
            searchText={searchText}
            onChange={(text) => setSearchText(text)}
          />
          <ul>
            {filteredServicesGroupedByCategories.map((category) => (
              <li key={category.id}>
                <ServiceCategoryItem category={category}>
                  <ul className="divide-border_color divide-y">
                    {category.services.map((service) => (
                      <li className="py-4" key={service.id}>
                        <ServiceItem service={service}>
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
                              styles="py-2 px-4"
                              name="Reserve"
                              buttonText="Reserve"
                            />
                          </Link>
                        </ServiceItem>
                      </li>
                    ))}
                  </ul>
                </ServiceCategoryItem>
              </li>
            ))}
          </ul>
        </div>
        <div className="flex flex-col gap-6 lg:w-1/3">
          <ReservationSection name="About us" show={merchantInfo.about_us}>
            <p>{merchantInfo.about_us}</p>
          </ReservationSection>
          <ReservationSection name="Opening hours" show={true}>
            <div className="flex flex-col gap-2">
              {[1, 2, 3, 4, 5, 6, 0].map((dayIndex) => {
                const slots = merchantInfo.business_hours[dayIndex];
                return (
                  <div key={dayIndex} className="flex justify-between">
                    <p className="w-1/3">{days[dayIndex]}</p>
                    <p className="w-2/3">
                      {slots?.length > 0
                        ? slots
                            .map(
                              (slot) =>
                                `${slot.start_time.slice(0, 5)} - ${slot.end_time.slice(0, 5)}`
                            )
                            .join(" & ")
                        : "Closed"}
                    </p>
                  </div>
                );
              })}
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
    </Card>
  );
}
