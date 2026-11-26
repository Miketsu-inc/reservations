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

async function fetchIsFavorite(merchantId, locationId) {
  const response = await fetch(
    `/api/v1/favorites/merchants/${merchantId}/locations/${locationId}`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

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

function isFavoriteQueryOptions(merchantId, locationId) {
  return queryOptions({
    queryKey: ["is-favorite", merchantId, locationId],
    queryFn: () => fetchIsFavorite(merchantId, locationId),
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

export default function FavoriteButton({ merchantId, locationId, styles }) {
  const queryClient = useQueryClient();
  const { showToast } = useToast();

  const { data: favoriteData } = useQuery({
    ...isFavoriteQueryOptions(merchantId, locationId),
    enabled: Boolean(merchantId && locationId),
  });

  const isFavorite = favoriteData?.is_favorite;
  const [isToggling, setIsToggling] = useState(false);

  async function handleToggleFavorite() {
    if (!favoriteData?.is_authenticated) {
      showToast({
        message: "Please log in to save this merchant to your favorites",
        variant: "info",
      });
      return;
    }

    if (!merchantId || !locationId || isToggling) return;

    try {
      setIsToggling(true);

      if (isFavorite) {
        await removeFavorite(merchantId, locationId);
      } else {
        await addFavorite(merchantId, locationId);
      }

      await queryClient.invalidateQueries({
        queryKey: isFavoriteQueryOptions(merchantId, locationId).queryKey,
      });
    } catch {
      showToast({ message: "Failed to update favorite", variant: "error" });
    } finally {
      setIsToggling(false);
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
      disabled={isToggling}
      aria-label={isFavorite ? "Remove from favorites" : "Add to favorites"}
    >
      <Icon
        icon={FavouriteIcon}
        styles={`size-5 transition-all duration-200 group-hover:scale-105 ${
          isFavorite
            ? "text-danger fill-danger "
            : "text-fg group-hover:text-fg-danger "
          }`}
      />
    </button>
  );
}
