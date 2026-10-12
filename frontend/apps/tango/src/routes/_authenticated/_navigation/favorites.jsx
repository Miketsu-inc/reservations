import {
  FavouriteIcon,
  Location01Icon,
  StarIcon,
} from "@hugeicons/core-free-icons";
import {
  Button,
  Icon,
  Loading,
  SearchInput,
  ServerError,
} from "@reservations/components";
import { removeFavorite, useToast } from "@reservations/lib";
import { queryOptions, useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";

async function fetchFavorites() {
  const response = await fetch("/api/v1/favorites", {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

function favoritesQueryOptions() {
  return queryOptions({
    queryKey: ["favourites"],
    queryFn: fetchFavorites,
  });
}

export const Route = createFileRoute("/_authenticated/_navigation/favorites")({
  loader: async ({ context: { queryClient } }) => {
    await queryClient.ensureQueryData(favoritesQueryOptions());
  },
  pendingComponent: Loading,
  errorComponent: ({ error }) => <ServerError error={error.message} />,
  component: FavoritesPage,
});

function FavoritesPage() {
  const [searchText, setSearchText] = useState("");
  const [removingId, setRemovingId] = useState(null);
  const { data: favorites } = useSuspenseQuery(favoritesQueryOptions());
  const { queryClient } = Route.useRouteContext();
  const { showToast } = useToast();

  async function handleRemoveFavorite(merchantId) {
    if (removingId) return;
    try {
      setRemovingId(merchantId);

      await removeFavorite(merchantId);

      await queryClient.invalidateQueries({ queryKey: ["favourites"] });
    } catch {
      showToast({ message: "Failed to remove favorite", variant: "error" });
    } finally {
      setRemovingId(null);
    }
  }

  const filteredFavorites = favorites.filter((fav) => {
    if (!searchText?.trim()) return true;
    const query = searchText.toLowerCase();
    return (
      fav.merchant_name?.toLowerCase().includes(query) ||
      fav.formatted_location?.toLowerCase().includes(query)
    );
  });

  return (
    <div className="flex justify-center">
      <div className="flex w-full max-w-2xl flex-col justify-center">
        <div className="flex flex-row items-center justify-between pt-4 pb-6">
          <p className="text-2xl font-semibold">Favorites</p>
          {favorites.length > 0 && (
            <SearchInput
              styles="w-40! md:w-60!"
              searchText={searchText}
              onChange={setSearchText}
            />
          )}
        </div>

        {favorites.length === 0 ? (
          <div
            className="flex h-full flex-col items-center justify-center py-16
              text-center"
          >
            <div className="bg-primary/10 text-primary mb-4 rounded-full p-4">
              <Icon icon={FavouriteIcon} styles="size-10" />
            </div>
            <p className="text-lg font-medium">You do not have any favorites</p>
            <p className="text-text_color/60 mt-1 max-w-sm text-sm">
              Explore businesses and tap the heart icon on any merchant page to
              save them here for quick access.
            </p>
            <Link to="/home" className="mt-6">
              <Button buttonText="Explore" styles="px-6 py-2" />
            </Link>
          </div>
        ) : filteredFavorites.length === 0 ? (
          <div
            className="flex flex-col items-center justify-center py-12
              text-center"
          >
            <p className="text-base">No favorites match "{searchText}"</p>
            <button
              className="text-primary mt-2 cursor-pointer text-sm underline"
              onClick={() => setSearchText("")}
            >
              Clear search
            </button>
          </div>
        ) : (
          <div className="grid w-full grid-cols-1 gap-4 sm:grid-cols-2">
            {filteredFavorites.map((merchant) => (
              <FavoriteMerchantCard
                key={merchant.merchant_id}
                merchant={merchant}
                onRemove={() => handleRemoveFavorite(merchant.merchant_id)}
                isRemoving={removingId === merchant.merchant_id}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function FavoriteMerchantCard({ merchant, onRemove, isRemoving }) {
  const merchantUrl = `/m/${merchant.merchant_url_name}`;

  return (
    <div
      className="border-border_color bg-layer_bg dark:hover:bg-layer_bg/80
        relative flex w-full flex-col justify-between gap-4 rounded-lg border
        p-2 shadow-sm transition-colors hover:border-gray-400 hover:bg-gray-100
        dark:hover:border-zinc-700"
    >
      <div className="flex flex-col gap-3">
        <div
          className="relative flex h-36 w-full shrink-0 overflow-hidden
            rounded-lg"
        >
          <img
            className="size-full object-cover"
            src="https://dummyimage.com/300x150/c402c4/000000"
            alt="service photo"
          />
          <button
            className="group bg-layer_bg border-border_color absolute top-2
              right-2 h-fit cursor-pointer rounded-full border p-1.5 shadow-sm
              transition-all duration-200 hover:scale-105 hover:shadow-md
              active:scale-95 disabled:pointer-events-none disabled:opacity-50"
            onClick={onRemove}
            disabled={isRemoving}
          >
            <Icon
              icon={FavouriteIcon}
              styles="size-6 text-red-500 fill-red-500 transition-all
                duration-200 group-hover:scale-105"
            />
          </button>
        </div>
        <div className="flex flex-col items-start gap-1.5 px-2">
          <div className="flex w-full items-center justify-between">
            <p className="text-2xl font-semibold">{merchant.merchant_name}</p>
            <div className="flex items-center gap-1 text-sm">
              <Icon
                icon={StarIcon}
                styles="size-5 text-yellow-500 fill-yellow-500"
              />
              <span className="font-medium">4.6</span>
              <span className="ml-1 cursor-pointer text-gray-500">(9)</span>
            </div>
          </div>
          {merchant.formatted_location && (
            <div
              className="text-text_color/70 flex flex-row items-center gap-1
                text-sm"
            >
              <Icon icon={Location01Icon} styles="size-4 shrink-0" />
              <span className="line-clamp-1">
                {merchant.formatted_location}
              </span>
            </div>
          )}
        </div>
      </div>

      <div className="px-2 pb-1">
        <Link to={merchantUrl}>
          <Button buttonText="Reserve" styles="py-2 text-sm w-full" />
        </Link>
      </div>
    </div>
  );
}
