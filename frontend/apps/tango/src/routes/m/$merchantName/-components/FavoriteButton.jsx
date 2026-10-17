import { FavouriteIcon } from "@hugeicons/core-free-icons";
import {
  Icon,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@reservations/components";
import { addFavorite, removeFavorite, useToast } from "@reservations/lib";
import { queryOptions, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

async function fetchIsFavorite(merchantId) {
  const response = await fetch(`/api/v1/favorites/${merchantId}`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  if (response.status === 401) {
    return { is_favorite: false, is_authenticated: false };
  }

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return { is_favorite: result.data.is_favorite, is_authenticated: true };
  }
}

function isFavoriteQueryOptions(merchantId) {
  return queryOptions({
    queryKey: ["is-favourite", merchantId],
    queryFn: () => fetchIsFavorite(merchantId),
    retry: false,
  });
}

function withConditionalTooltip(condition, content, children) {
  if (!condition) return children;

  return (
    <Tooltip>
      <TooltipTrigger asChild>{children}</TooltipTrigger>
      <TooltipContent>{content}</TooltipContent>
    </Tooltip>
  );
}

export default function FavoriteButton({ merchantId, styles }) {
  const queryClient = useQueryClient();
  const { showToast } = useToast();

  const { data: favoriteData } = useQuery({
    ...isFavoriteQueryOptions(merchantId),
    enabled: Boolean(merchantId),
  });

  const isFavourite = favoriteData?.is_favorite;
  const [isTogglingFav, setIsTogglingFav] = useState(false);

  async function handleToggleFavorite() {
    if (!favoriteData?.is_authenticated) {
      showToast({
        message: "Please log in to save this merchant to your favorites",
        variant: "info",
      });
      return;
    }

    if (!merchantId || isTogglingFav) return;

    try {
      setIsTogglingFav(true);

      if (isFavourite) {
        await removeFavorite(merchantId);
      } else {
        await addFavorite(merchantId);
      }

      await queryClient.invalidateQueries({
        queryKey: ["is-favourite", merchantId],
      });
    } catch {
      showToast({ message: "Failed to update favorite", variant: "error" });
    } finally {
      setIsTogglingFav(false);
    }
  }

  return withConditionalTooltip(
    !favoriteData?.is_authenticated,
    "Log in to save favorites",
    <button
      className={`group cursor-pointer transition-all duration-200
        active:scale-95 disabled:pointer-events-none disabled:opacity-50
        ${styles || ""}`}
      onClick={handleToggleFavorite}
      disabled={isTogglingFav}
      aria-label={isFavourite ? "Remove from favorites" : "Add to favorites"}
    >
      <Icon
        icon={FavouriteIcon}
        styles={`size-5 transition-all duration-200 group-hover:scale-105 ${
          isFavourite
            ? "text-fg-danger fill-danger "
            : "text-fg group-hover:text-fg-danger "
          }`}
      />
    </button>
  );
}
