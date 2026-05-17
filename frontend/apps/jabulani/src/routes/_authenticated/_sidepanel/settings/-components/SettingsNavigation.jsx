import {
  Briefcase04Icon,
  Calendar02Icon,
  CreditCardIcon,
  TimeScheduleIcon,
  User03Icon,
} from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import NavigationItem from "./NavigationItem";

export default function SettingsNavigation() {
  return (
    <nav className="">
      <ul className="flex flex-col items-start justify-center">
        <NavigationItem label="Profile" path="/settings/profile">
          <Icon icon={User03Icon} styles="size-6" />
        </NavigationItem>
        <NavigationItem label="Merchant" path="/settings/merchant">
          <Icon icon={Briefcase04Icon} styles="size-6" />
        </NavigationItem>
        <NavigationItem label="Calendar" path="/settings/calendar">
          <Icon icon={Calendar02Icon} styles="size-6" />
        </NavigationItem>
        <NavigationItem label="Billing" path="/settings/billing">
          <Icon icon={CreditCardIcon} styles="size-6" />
        </NavigationItem>
        <NavigationItem label="Scheduling" path="/settings/scheduling">
          <Icon icon={TimeScheduleIcon} styles="size-6" />
        </NavigationItem>
      </ul>
    </nav>
  );
}
