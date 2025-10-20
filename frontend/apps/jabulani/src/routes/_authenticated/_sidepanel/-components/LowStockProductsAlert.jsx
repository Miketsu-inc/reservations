import {
  BackArrowIcon,
  ProductIcon,
  TickIcon,
  WarningIcon,
} from "@reservations/assets";
import { Card } from "@reservations/components";
import { getDisplayUnit } from "@reservations/lib";
import { Link } from "@tanstack/react-router";

export default function LowStockProductsAlert({ products, route }) {
  const totalLowStock = products.length;
  const hasLowStock = totalLowStock > 0;

  function getStockSeverity(stock, fill_ratio) {
    const percentage = fill_ratio * 100;
    if (stock === 0) return "text-red-600";
    if (percentage < 5) return "text-red-500";
    if (percentage <= 10) return "text-orange-500";
    if (percentage <= 20) return "text-amber-500";
    if (percentage <= 35) return "text-yellow-500";
    return "text-amber-400";
  }

  function getProgressColor(stock, fill_ratio) {
    const percentage = fill_ratio * 100;
    if (stock === 0) return "bg-red-600";
    if (percentage < 5) return "bg-red-500";
    if (percentage <= 10) return "bg-orange-500";
    if (percentage <= 20) return "bg-amber-500";
    if (percentage <= 35) return "bg-yellow-500";
    return "bg-amber-400";
  }

  function getPercentLabel(stock, fill_ratio) {
    if (stock === 0) return "Out of stock";
    const percent = fill_ratio * 100;
    if (percent < 1) return "<1%";
    if (percent < 10) return `${percent.toFixed(1)}%`;
    return `${Math.round(percent)}%`;
  }

  return (
    <Card styles="p-0!">
      <div
        className={`flex flex-col justify-center ${hasLowStock ? "" : "gap-12"}`}
      >
        <div
          className="border-border_color flex items-center justify-between
            border-b p-4"
        >
          <div className="flex items-center">
            <div
              className={`mr-3 rounded-md p-2 ${
                hasLowStock
                  ? "bg-amber-500/20 dark:bg-amber-500/20"
                  : "bg-green-100 dark:bg-green-500/20"
                }`}
            >
              {hasLowStock ? (
                <WarningIcon
                  styles="h-6 w-6 shrink-0 text-amber-600 dark:text-amber-400"
                />
              ) : (
                <ProductIcon styles="h-6 w-6 shrink-0 text-green-600" />
              )}
            </div>
            <div>
              <h3 className="text-lg text-nowrap dark:font-semibold">
                Low Stock {hasLowStock ? `(${totalLowStock})` : ""}
              </h3>
              {hasLowStock && (
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  Products about to run out
                </p>
              )}
            </div>
          </div>

          <Link
            from={route.fullPath}
            to="/products"
            className="text-text_color/70 hover:bg-hvr_gray flex items-center
              justify-end gap-1 rounded-lg p-0 text-sm sm:p-2"
          >
            <span className="hidden text-right sm:block">Manage Products</span>
            <BackArrowIcon
              styles="stroke-text_color/70 h-5 w-5 shrink-0 rotate-180"
            />
          </Link>
        </div>

        {hasLowStock ? (
          <div className="max-h-[230px] overflow-y-auto px-2 dark:scheme-dark">
            <div>
              {products.map((product) => {
                const {
                  current,
                  max,
                  unit: displayUnit,
                } = getDisplayUnit(
                  product.current_amount,
                  product.max_amount,
                  product.unit
                );

                return (
                  <div key={product.id} className="mb-1 rounded-md px-2 py-2.5">
                    <div
                      className="mb-1 flex items-center justify-between gap-4"
                    >
                      <div className="flex items-center">
                        <div
                          className={`h-2 w-2 rounded-full ${getProgressColor(
                            product.current_amount, product.fill_ratio )}`}
                        ></div>
                        <span
                          className="max-w-48 truncate pl-2 text-sm font-medium
                            sm:max-w-xl"
                        >
                          {product.name}
                        </span>
                      </div>
                      <div
                        className="flex items-center justify-center gap-2
                          text-xs font-medium whitespace-nowrap text-gray-500
                          dark:text-gray-400"
                      >
                        <span>
                          {current} / {max}
                        </span>
                        <span> {displayUnit} </span>
                      </div>
                    </div>
                    <div className="flex items-center">
                      <div
                        className="dark:bg-bg_color mr-2 h-1.5 w-full
                          rounded-full bg-gray-200"
                      >
                        <div
                          className={`h-1.5 rounded-full ${getProgressColor(
                            product.current_amount, product.fill_ratio )}`}
                          style={{
                            width: `${Math.min(100, (product.current_amount / product.max_amount) * 100)}%`,
                          }}
                        ></div>
                      </div>
                      <div
                        className={`text-xs font-semibold text-nowrap
                          ${getStockSeverity( product.current_amount,
                          product.fill_ratio )}`}
                      >
                        {getPercentLabel(
                          product.current_amount,
                          product.fill_ratio
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center">
            <div
              className="mb-4 flex w-min items-center justify-center
                rounded-full bg-green-100 p-2 dark:bg-green-500/20"
            >
              <TickIcon styles="h-12 w-12 text-green-600" />
            </div>

            <p className="text-sm text-gray-600 dark:text-gray-400">
              All products are well-stocked
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-500">
              No items require immediate attention
            </p>
          </div>
        )}
      </div>
    </Card>
  );
}
