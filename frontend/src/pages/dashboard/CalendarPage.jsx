import { Suspense, lazy } from "react";
import Loading from "../../components/Loading";
const Calendar = lazy(() => import("./Calendar"));

export default function CalendarPage() {
  return (
    <span className="light">
      <Suspense fallback={<Loading />}>
        <Calendar />
      </Suspense>
    </span>
  );
}
