import ToastProvider from "@components/ToastProvider";
import { createRouter, RouterProvider } from "@tanstack/react-router";
import React from "react";
import ReactDOM from "react-dom/client";
import { routeTree } from "./routeTree.gen";

const router = createRouter({
  routeTree,
  // defaultPreload: "intent",
  // defaultPreloadStaleTime: 1000,
});

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <ToastProvider>
      <RouterProvider router={router} />
    </ToastProvider>
  </React.StrictMode>
);
