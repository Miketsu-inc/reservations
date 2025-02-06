import Briefcase from "@icons/BriefcaseIcon";
import CalendarIcon from "@icons/CalendarIcon";
import CreditCardIcon from "@icons/CreditCardIcon";
import PersonIcon from "@icons/PersonIcon";
import NavigationItem from "./NavigationItem";

export default function SettingsNavigation() {
  return (
    <nav className="">
      <ul className="flex flex-col items-start justify-center">
        <NavigationItem label="Profile" path="/settings/profile">
          <PersonIcon styles="h-6 w-6" />
        </NavigationItem>
        <NavigationItem label="Merchant" path="/settings/merchant">
          <Briefcase styles="h-6 w-6" />
        </NavigationItem>
        <NavigationItem label="Calendar" path="/settings/calendar">
          <CalendarIcon styles="h-6 w-6" />
        </NavigationItem>
        <NavigationItem label="Billing" path="/settings/billing">
          <CreditCardIcon styles="h-6 w-6" />
        </NavigationItem>
      </ul>
    </nav>
  );
}
