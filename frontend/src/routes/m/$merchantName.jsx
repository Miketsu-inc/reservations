import { createFileRoute } from "@tanstack/react-router";
import MerchantPage from "../../pages/reservation/MerchantPage";

export const Route = createFileRoute("/m/$merchantName")({
  component: MerchantPage,
});
