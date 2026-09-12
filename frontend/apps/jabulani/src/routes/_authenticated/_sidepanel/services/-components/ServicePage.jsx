import {
  Button,
  DeleteModal,
  ScrollSpyNav,
  ScrollSpyProvider,
  ScrollSpySection,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { Block, useRouter } from "@tanstack/react-router";
import { useMemo, useRef, useState } from "react";
import ProductAdder from "./ProductAdder";
import { ServiceBasicDetails } from "./ServiceBasicDetails";
import ServiceBookingSettings from "./ServiceBookingSettings";
import { ServicePricingDuration } from "./ServicePricingDuration";
import { normalizeServicePhases } from "./servicehooks";

export default function ServicePage({
  service,
  categories,
  products,
  onSave,
  route,
}) {
  const originalData = useMemo(
    () => ({
      id: service?.id,
      booking_type: service?.booking_type ?? "appointment",
      name: service?.name || "",
      description: service?.description || "",
      color: service?.color || "#2334b8",
      price: service?.price ?? null,
      price_type: service?.price_type || "fixed",
      category_id: service?.category_id || null,
      is_active: service?.is_active ?? true,
      duration: service?.total_duration || "",
      duration_unit: service?.duration_unit || "min",
      min_participants: service?.min_participants || undefined,
      max_participants: service?.max_participants || undefined,
      settings: {
        cancel_deadline: service?.settings?.cancel_deadline || null,
        booking_window_min: service?.settings?.booking_window_min || null,
        booking_window_max: service?.settings?.booking_window_max || null,
        buffer_time: service?.settings?.buffer_time || null,
        approval_policy: service?.settings?.approval_policy || null,
      },
      phases: service?.phases || [],
      employee_ids: [],
      used_products: service?.used_products || [],
    }),
    [service]
  );
  const router = useRouter();
  const [serviceData, setServiceData] = useState(originalData);
  const lastSavedData = useRef(originalData);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const { showToast } = useToast();
  const { merchantId } = useAuth();

  const isGroupService = serviceData.booking_type !== "appointment";

  function updateServiceData(data) {
    setServiceData((prev) => ({ ...prev, ...data }));
  }

  async function deleteHandler() {
    const response = await fetch(
      `/api/v1/merchants/${merchantId}/services/${service.id}`,
      {
        method: "DELETE",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
      }
    );

    if (!response.ok) {
      invalidateLocalStorageAuth(response.status);
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      lastSavedData.current = serviceData;
      router.navigate({
        from: route.fullPath,
        to: "/services",
      });
      showToast({
        message: "Service deleted successfully",
        variant: "success",
      });
    }
  }

  async function saveService() {
    const phases = normalizeServicePhases(serviceData.phases);

    if (phases.length === 0) {
      showToast({
        message: "Please set a duration",
        variant: "error",
      });
      return;
    }

    const data = {
      ...serviceData,
      phases: isGroupService
        ? [{ ...phases[0], sequence: 1, phase_type: "active" }]
        : phases,
    };

    delete data.duration;
    delete data.duration_unit;

    const didSave = await onSave(data);

    if (didSave) {
      lastSavedData.current = serviceData;
      router.navigate({ from: route.fullPath, to: "/services" });
    }
  }

  return (
    <Block
      shouldBlockFn={() => {
        if (
          JSON.stringify(serviceData) === JSON.stringify(lastSavedData.current)
        )
          return false;

        const canLeave = confirm(
          "You have unsaved changes, are you sure you want to leave?"
        );
        return !canLeave;
      }}
    >
      {service && (
        <DeleteModal
          isOpen={showDeleteModal}
          onClose={() => setShowDeleteModal(false)}
          onDelete={deleteHandler}
          itemName={service.name}
        />
      )}
      <ScrollSpyProvider scrollOffset={80}>
        <div className="w-full">
          <div className="mx-auto grid w-full max-w-6xl">
            <div
              className="flex flex-row items-center justify-between px-4 pt-6
                pb-2 md:py-4"
            >
              <p className="text-2xl">
                {serviceData.id ? "Edit service" : "New service"}
              </p>
              <div
                className="bg-layer_bg border-t-border_color fixed bottom-0
                  left-0 z-30 w-full border-t p-4 md:static md:w-auto
                  md:border-t-0 md:bg-transparent md:p-0"
              >
                <Button
                  styles="py-2 px-6 w-full"
                  variant="primary"
                  buttonText="Save"
                  onClick={saveService}
                />
              </div>
            </div>
            <div
              className="grid grid-cols-1 md:grid-cols-[13rem_minmax(0,1fr)]
                md:gap-8 md:px-4"
            >
              <ScrollSpyNav />
              <div className="min-w-0 px-4 pt-4 pb-16 md:px-0 md:pt-0 md:pb-0">
                <div className="flex flex-col gap-16">
                  <ScrollSpySection id="basicDetails" label="Basic details">
                    <p className="mb-8 text-xl font-semibold">Basic details</p>
                    <ServiceBasicDetails
                      service={serviceData}
                      categories={categories}
                      onUpdate={updateServiceData}
                    />
                  </ScrollSpySection>
                  <ScrollSpySection
                    id="pricingDuration"
                    label="Pricing & duration"
                  >
                    <p className="mb-8 text-xl font-semibold">
                      Pricing & duration
                    </p>
                    <ServicePricingDuration
                      service={serviceData}
                      setService={setServiceData}
                      onUpdate={updateServiceData}
                    />
                  </ScrollSpySection>
                  <ScrollSpySection id="products" label="Products">
                    <p className="mb-8 text-xl font-semibold">Products</p>
                    <ProductAdder
                      availableProducts={products}
                      usedProducts={serviceData.used_products}
                      onUpdate={(updated) =>
                        updateServiceData({ used_products: updated })
                      }
                    />
                  </ScrollSpySection>
                  <ScrollSpySection id="bookings" label="Bookings">
                    <p className="mb-8 text-xl font-semibold">Bookings</p>
                    <ServiceBookingSettings
                      onUpdate={updateServiceData}
                      settings={serviceData.settings}
                    />
                  </ScrollSpySection>
                  {service && (
                    <Button
                      type="button"
                      styles="py-4 mb-2 shadow-none bg-transparent
                        hover:bg-transparent! text-red-500!"
                      buttonText="Delete service"
                      onClick={() => setShowDeleteModal(true)}
                    ></Button>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </ScrollSpyProvider>
    </Block>
  );
}
