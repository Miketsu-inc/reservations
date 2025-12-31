import {
  BriefcaseIcon,
  CalendarIcon,
  CreditCardIcon,
  PersonIcon,
} from "@reservations/assets";
import NavigationItem from "./NavigationItem";

export default function SettingsNavigation() {
  return (
    <nav className="">
      <ul className="flex flex-col items-start justify-center">
        <NavigationItem label="Profile" path="/settings/profile">
          <PersonIcon styles="h-6 w-6 fill-current" />
        </NavigationItem>
        <NavigationItem label="Merchant" path="/settings/merchant">
          <BriefcaseIcon styles="h-6 w-6" />
        </NavigationItem>
        <NavigationItem label="Calendar" path="/settings/calendar">
          <CalendarIcon styles="h-6 w-6" />
        </NavigationItem>
        <NavigationItem label="Billing" path="/settings/billing">
          <CreditCardIcon styles="h-6 w-6" />
        </NavigationItem>
        <NavigationItem label="Scheduling" path="/settings/scheduling">
          <CreditCardIcon styles="h-6 w-6" />
        </NavigationItem>
      </ul>
    </nav>
  );
}
