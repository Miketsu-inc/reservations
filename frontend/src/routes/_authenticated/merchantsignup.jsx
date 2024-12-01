import { createFileRoute } from "@tanstack/react-router";
import MerchantSignup from "../../pages/onboarding/MerchantSignup";

export const Route = createFileRoute("/_authenticated/merchantsignup")({
  component: MerchantSignup,
});
