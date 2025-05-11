import Button from "@components/Button";
import ProductIcon from "@icons/ProductIcon";
import WarningIcon from "@icons/WarningIcon";
import XIcon from "@icons/XIcon";
import { Link } from "@tanstack/react-router";
import { useState } from "react";

export default function LowStockProductsAlert({ products, route }) {
  const [dismissedProducts, setDismissedProducts] = useState([]);

  const activeAlerts = products?.filter(
    (product) => !dismissedProducts.includes(product.id)
  );

  const hasAlerts = activeAlerts?.length > 0;

  const dismissProduct = (id) => {
    setDismissedProducts([...dismissedProducts, id]);
  };

  const restoreAlerts = () => {
    setDismissedProducts([]);
  };

  return (
    <div className="bg-layer_bg flex w-full flex-col rounded-lg shadow-sm">
      <div
        className="flex items-center justify-between rounded-lg bg-orange-50 p-4
          dark:bg-orange-900/20"
      >
        <div className="flex gap-3">
          <WarningIcon styles="size-5 text-orange-500 dark:text-orange-400" />
          <span className="font-medium text-gray-800 dark:text-gray-100">
            Low Stock Products Alert
          </span>
        </div>
      </div>
      <div
        className={`${hasAlerts ? "h-48" : ""} overflow-y-auto px-4 dark:[color-scheme:dark]`}
      >
        {hasAlerts ? (
          <ul className="divide-y divide-gray-100 dark:divide-gray-700">
            {activeAlerts.map((product) => (
              <li
                key={product.id}
                className="flex items-center justify-between py-3"
              >
                <div className="flex-1">
                  <p className="font-medium text-gray-800 dark:text-gray-200">
                    {product.name}
                  </p>
                  <div className="mt-1 flex items-center">
                    <div className="mr-2 size-3 rounded-full bg-amber-500"></div>
                    <span className="text-sm font-medium text-amber-600 dark:text-amber-400">
                      {product.stock_quantity}{" "}
                      {product.stock_quantity === 1 ? "unit" : "units"} left
                    </span>
                  </div>
                </div>
                <button
                  onClick={() => dismissProduct(product.id)}
                  className="rounded-full p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600
                    dark:text-gray-500 dark:hover:bg-gray-700 dark:hover:text-gray-300"
                  aria-label="Dismiss alert"
                >
                  <XIcon styles="stroke-text_color size-5" />
                </button>
              </li>
            ))}
          </ul>
        ) : (
          <div className="flex flex-col items-center justify-center py-10 text-center">
            <div className="mb-3 rounded-full bg-gray-300 p-3 dark:bg-gray-700">
              <ProductIcon styles="text-gray-500 dark:text-gray-400 size-5" />
            </div>
            <h4 className="mb-1 text-lg font-medium text-gray-700 dark:text-gray-300">
              No Low Stock Alerts
            </h4>
            <p className="mb-4 text-gray-500 dark:text-gray-400">
              All your products have sufficient inventory
            </p>
            {dismissedProducts.length > 0 && (
              <button
                onClick={restoreAlerts}
                className="text-sm font-medium text-orange-500 hover:text-orange-600 dark:text-orange-400
                  dark:hover:text-orange-300"
              >
                Restore dismissed alerts ({dismissedProducts.length})
              </button>
            )}
          </div>
        )}
      </div>
      {hasAlerts && (
        <div
          className="mt-auto border-t border-gray-100 bg-gray-50 p-4 dark:border-gray-700
            dark:bg-gray-900/40"
        >
          <Link from={route.fullPath} to="/products">
            <Button styles="w-full py-2" buttonText="Manage Products" />
          </Link>
        </div>
      )}
    </div>
  );
}
