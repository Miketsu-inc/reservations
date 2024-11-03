import { createFileRoute } from "@tanstack/react-router";
import ReservationPage from "../../pages/reservation/ReservationPage";

export const Route = createFileRoute("/m/$merchantName")({
  component: ReservationPage,
});
