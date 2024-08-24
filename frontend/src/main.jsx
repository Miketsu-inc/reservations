import React from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider, createBrowserRouter } from "react-router-dom";
import DashboardPage from "./pages/dashboard/DashboardPage.jsx";
import LandingPage from "./pages/landing/LandingPage.jsx";
import LoginPage from "./pages/onboarding/LoginPage.jsx";
import SingUpPage from "./pages/onboarding/SignUpPage.jsx";
import ReservationPage from "./pages/reservation/ReservationPage.jsx";

// make sure to also add the route path to router.go
const router = createBrowserRouter([
  {
    path: "/",
    element: <LandingPage />,
  },
  {
    path: "/signup",
    element: <SingUpPage />,
  },
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/reservation",
    element: <ReservationPage />,
  },
  {
    path: "/dashboard",
    element: <DashboardPage />,
  },
]);

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
);
