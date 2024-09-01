import Calendar from "./Calendar";
import SidePanel from "./SidePanel";

export default function DashboardPage() {
  return (
    <div className="h-screen overflow-y-auto">
      <SidePanel
        profileImage="https://dummyimage.com/40x40/000/fff.png&text=logo"
        profileText="Company name"
      />
      <div className="light min-h-screen p-4 sm:ml-64">
        <div className="rounded-lg bg-bg_color p-4">
          <Calendar />
        </div>
      </div>
    </div>
  );
}
