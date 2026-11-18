import { StarIcon } from "@hugeicons/core-free-icons";
import { Fragment } from "react";
import Icon from "./Icon.jsx";

const STARS = [1, 2, 3, 4, 5];

export default function StarRating({
  value = 0,
  onChange,
  iconStyles = "size-5",
  styles = "gap-1",
}) {
  function renderStar(star) {
    return (
      <Icon
        icon={StarIcon}
        styles={`${iconStyles} ${
          star <= value
            ? "fill-warning text-warning"
            : "fill-border-strong text-border-strong"
          }`}
      />
    );
  }

  return (
    <div className={`flex items-center ${styles}`}>
      {STARS.map((star) =>
        onChange ? (
          <button
            key={star}
            type="button"
            onClick={() => onChange(star)}
            className="cursor-pointer p-1"
            aria-label={`${star} star${star > 1 ? "s" : ""}`}
          >
            {renderStar(star)}
          </button>
        ) : (
          <Fragment key={star}>{renderStar(star)}</Fragment>
        )
      )}
    </div>
  );
}
