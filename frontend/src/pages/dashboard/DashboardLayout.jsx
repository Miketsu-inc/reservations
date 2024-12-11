import { Outlet } from "@tanstack/react-router";
import SidePanel from "./SidePanel";

export default function DashboardLayout() {
  return (
    <div className="h-screen overflow-y-auto">
      <SidePanel
        profileImage="https://dummyimage.com/40x40/000/fff.png&text=logo"
        profileText="Company name"
      />
      <div className="min-h-screen p-4 sm:ml-64">
        <div className="rounded-lg bg-bg_color p-4">
          <Outlet />
        </div>
      </div>
    </div>
  );
}