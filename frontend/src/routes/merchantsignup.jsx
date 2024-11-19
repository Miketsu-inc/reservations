import { createFileRoute, redirect } from "@tanstack/react-router";
import Loading from "../components/Loading";
import { isAuthenticated } from "../lib/lib";
import MerchantSignup from "../pages/onboarding/MerchantSignup";

export const Route = createFileRoute("/merchantsignup")({
  component: MerchantSignup,
  beforeLoad: async () => {
    if (!(await isAuthenticated("api/v1/auth/user"))) {
      throw redirect({
        to: "/login",
      });
    }
  },
  pendingComponent: Loading,
});
