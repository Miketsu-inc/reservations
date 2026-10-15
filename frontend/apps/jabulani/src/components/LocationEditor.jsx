import { Location01Icon } from "@hugeicons/core-free-icons";
import { SearchBox } from "@mapbox/search-js-react";
import { Button, Icon } from "@reservations/components";
import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import {
  coordinatesFromLocation,
  locationFromMapboxFeature,
} from "./locationUtils.js";

const accessToken = import.meta.env.VITE_MAPBOX_TOKEN;
const LocationMap = lazy(() => import("./LocationMap.jsx"));
const searchBoxTheme = {
  variables: {
    colorText: "var(--fg)",
    colorSecondary: "var(--fg-muted)",
    colorBackground: "var(--bg-raised)",
    colorBackgroundHover: "var(--hover)",
    colorBackgroundActive: "var(--active)",
    border: "1px solid var(--border-strong)",
    borderRadius: "0.5rem",
    boxShadow: "none",
  },
  cssText: `
    .Input,
    .Input:focus {
      height: 40px;
      color: var(--colorText);
    }

    .Results {
      max-height: min(20rem, calc(100dvh - 17rem));
      box-shadow: 0 12px 28px rgba(0, 0, 0, 0.22);
      scrollbar-color: var(--colorSecondary) transparent;
      scrollbar-width: thin;
    }
  `,
};

export default function LocationEditor({ value, onChange, compact = false }) {
  const [address, setAddress] = useState(value?.formatted_location ?? "");
  const [showMap, setShowMap] = useState(() =>
    Boolean(coordinatesFromLocation(value))
  );
  const [searchBoxKey, setSearchBoxKey] = useState(0);
  const locationRef = useRef(value);
  const addressRef = useRef(address);

  useEffect(() => {
    locationRef.current = value;
  }, [value]);

  const emitLocation = useCallback(
    (location) => {
      locationRef.current = location;
      onChange?.(location);
    },
    [onChange]
  );

  function handleAddressChange(nextAddress) {
    addressRef.current = nextAddress;
    setAddress(nextAddress);

    if (locationRef.current) {
      emitLocation({
        ...locationRef.current,
        formatted_location: nextAddress,
      });
    }
  }

  function handleRetrieve(result) {
    const feature = result.features?.[0];
    if (!feature) return;

    const formattedLocation = addressRef.current.trim();
    const location = locationFromMapboxFeature(
      feature,
      formattedLocation,
      locationRef.current
    );
    if (!location) return;

    setShowMap(true);
    setSearchBoxKey((current) => current + 1);
    emitLocation(location);
  }

  function handleManualPin() {
    setShowMap(true);

    navigator.geolocation?.getCurrentPosition(
      ({ coords }) => {
        emitLocation({
          ...(locationRef.current ?? {}),
          geo_point: {
            latitude: coords.latitude,
            longitude: coords.longitude,
          },
          formatted_location: addressRef.current.trim(),
        });
      },
      () => {},
      { enableHighAccuracy: false, timeout: 5000, maximumAge: 300000 }
    );
  }

  return (
    <div className={`flex flex-col ${compact ? "gap-3 lg:gap-4" : "gap-4"}`}>
      <div>
        <label className="text-fg mb-1 block text-sm font-medium">
          Address
        </label>
        <SearchBox
          key={searchBoxKey}
          accessToken={accessToken}
          popoverOptions={{ flip: true, offset: 4 }}
          theme={searchBoxTheme}
          value={address}
          onChange={handleAddressChange}
          onRetrieve={handleRetrieve}
          placeholder="Type your address..."
        />
        <p className="text-fg-muted mt-2 text-xs leading-5">
          {compact && (
            <span className="lg:hidden">
              Shown exactly as typed. Select a suggestion to place the pin.
            </span>
          )}
          <span className={compact ? "hidden lg:inline" : undefined}>
            {compact
              ? "Shown exactly as typed. Select a suggestion to place the pin."
              : "This exact text will be shown to customers. Choose an autocomplete result to position the pin, then edit the text further if needed."}
          </span>
        </p>
      </div>

      {!showMap && address.trim() && (
        <Button
          type="button"
          variant="tertiary"
          styles="w-fit px-3 py-2"
          buttonText="Set pin manually"
          onClick={handleManualPin}
        />
      )}

      {showMap ? (
        <Suspense
          fallback={
            <div
              className={`border-border bg-bg-sunken animate-pulse rounded-lg
                border ${compact ? "h-52 lg:h-72" : "h-72 sm:h-96"}`}
            />
          }
        >
          <LocationMap
            value={value}
            onChange={emitLocation}
            compact={compact}
          />
        </Suspense>
      ) : (
        <div
          className={`border-border relative overflow-hidden rounded-lg border
            ${compact ? "h-52 lg:h-72" : "h-72 sm:h-96"}`}
        >
          <div
            className="bg-bg-sunken flex h-full items-center justify-center p-6"
          >
            <div className="text-center">
              <Icon
                icon={Location01Icon}
                styles="mx-auto mb-3 size-14 text-fg-muted"
              />
              <p className="text-fg font-medium">
                {compact ? (
                  <span className="lg:hidden">
                    Choose an address to place the pin
                  </span>
                ) : (
                  "Choose an address to position the pin"
                )}
                {compact && (
                  <span className="hidden lg:inline">
                    Choose an address to position the pin
                  </span>
                )}
              </p>
              <p className="text-fg-muted mt-1 text-sm">
                {compact ? (
                  <span className="lg:hidden">
                    No matching result? Set the pin manually.
                  </span>
                ) : (
                  "If there is no matching result, type the address and set the pin manually."
                )}
                {compact && (
                  <span className="hidden lg:inline">
                    If there is no matching result, type the address and set the
                    pin manually.
                  </span>
                )}
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
