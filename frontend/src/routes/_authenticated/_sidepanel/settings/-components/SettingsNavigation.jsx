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
          <PersonIcon styles="dark:fill-gray-400 fill-gray-600" />
        </NavigationItem>
        <NavigationItem label="Merchant" path="/settings/merchant">
          <Briefcase styles="dark:stroke-gray-400 stroke-gray-600" />
        </NavigationItem>
        <NavigationItem label="Calendar" path="/settings/calendar">
          <CalendarIcon styles="dark:stroke-gray-400 stroke-gray-600 h-6 w-6" />
        </NavigationItem>
        <NavigationItem label="Billing" path="/settings/billing">
          <CreditCardIcon styles="dark:stroke-gray-400 stroke-gray-600" />
        </NavigationItem>
      </ul>
    </nav>
  );
}
