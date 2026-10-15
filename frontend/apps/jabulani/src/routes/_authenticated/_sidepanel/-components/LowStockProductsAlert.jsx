import {
  Alert02Icon,
  ArrowLeft01Icon,
  ShoppingBag02Icon,
  Tick02Icon,
} from "@hugeicons/core-free-icons";
import { Card, Icon, Loading, ServerError } from "@reservations/components";
import { getDisplayUnit, invalidateLocalStorageAuth } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

async function fetchLowStockProducts(merchantId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/products/low-stock`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function lowStockProductsQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "low-stock-products"],
    queryFn: () => fetchLowStockProducts(merchantId),
    // staleTime: 30_000,
    // gcTime: 5 * 60 * 1000,
  });
}

export default function LowStockProductsAlert({ merchantId, route }) {
  const {
    data: products,
    isLoading,
    isError,
    error,
  } = useQuery(lowStockProductsQueryOptions(merchantId));

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return (
      <Card>
        <Loading />
      </Card>
    );
  }

  const totalLowStock = products.length;
  const hasLowStock = totalLowStock > 0;

  function getStockSeverity(stock, fill_ratio) {
    const percentage = fill_ratio * 100;
    if (stock === 0) return "text-fg-danger";
    if (percentage < 5) return "text-fg-danger";
    if (percentage <= 10) return "text-fg-warning";
    if (percentage <= 20) return "text-fg-warning";
    if (percentage <= 35) return "text-fg-warning";
    return "text-fg-warning";
  }

  function getProgressColor(stock, fill_ratio) {
    const percentage = fill_ratio * 100;
    if (stock === 0) return "bg-danger";
    if (percentage < 5) return "bg-danger";
    if (percentage <= 10) return "bg-warning";
    if (percentage <= 20) return "bg-warning";
    if (percentage <= 35) return "bg-warning";
    return "bg-warning";
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
          className="border-border flex items-center justify-between border-b
            p-4"
        >
          <div className="flex items-center">
            <div
              className={`mr-3 rounded-md p-2 ${
                hasLowStock ? "bg-warning-subtle" : "bg-success-subtle"
              }`}
            >
              {hasLowStock ? (
                <Icon
                  icon={Alert02Icon}
                  styles="size-6 shrink-0 text-fg-warning"
                />
              ) : (
                <Icon
                  icon={ShoppingBag02Icon}
                  styles="size-6 shrink-0 text-fg-success"
                />
              )}
            </div>
            <div>
              <h3 className="text-lg text-nowrap dark:font-semibold">
                Low Stock {hasLowStock ? `(${totalLowStock})` : ""}
              </h3>
              {hasLowStock && (
                <p className="text-fg-muted text-xs">
                  Products about to run out
                </p>
              )}
            </div>
          </div>

          <Link
            from={route.fullPath}
            to="/products"
            className="text-fg-muted hover:bg-hover flex items-center
              justify-end gap-1 rounded-lg p-0 text-sm sm:p-2"
          >
            <span className="hidden text-right sm:block">Manage Products</span>
            <Icon
              icon={ArrowLeft01Icon}
              styles="stroke-fg-muted size-5 shrink-0 rotate-180"
            />
          </Link>
        </div>

        {hasLowStock ? (
          <div className="max-h-57.5 overflow-y-auto px-2 dark:scheme-dark">
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
                          className={`size-2 rounded-full ${getProgressColor(
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
                        className="text-fg-muted flex items-center
                          justify-center gap-2 text-xs font-medium
                          whitespace-nowrap"
                      >
                        <span>
                          {current} / {max}
                        </span>
                        <span> {displayUnit} </span>
                      </div>
                    </div>
                    <div className="flex items-center">
                      <div
                        className="bg-bg-sunken mr-2 h-1.5 w-full rounded-full"
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
              className="bg-success-subtle mb-4 flex w-min items-center
                justify-center rounded-full p-2"
            >
              <Icon icon={Tick02Icon} styles="size-12 text-fg-success" />
            </div>

            <p className="text-fg-muted text-sm">
              All products are well-stocked
            </p>
            <p className="text-fg-muted text-xs">
              No items require immediate attention
            </p>
          </div>
        )}
      </div>
    </Card>
  );
}
