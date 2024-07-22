import React from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider, createBrowserRouter } from "react-router-dom";
import Calendar from "./pages/dashboard/Calendar.jsx";
import LandingPage from "./pages/landing/LandingPage.jsx";
import LogIn from "./pages/onboarding/LogIn.jsx";
import SingUp from "./pages/onboarding/SignUp.jsx";

const router = createBrowserRouter([
  {
    path: "/",
    element: <LandingPage />,
  },
  {
    path: "/signup",
    element: <SingUp />,
  },
  {
    path: "/login",
    element: <LogIn />,
  },
  {
    path: "/calendar",
    element: <Calendar />,
  },
]);

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
);
