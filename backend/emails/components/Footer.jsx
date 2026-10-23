import React from "react";
import { Link, Section, Text } from "react-email";
import { t } from "../template";

void React;

const TANGO_URL = process.env.TANGO_URL;

export default function Footer() {
  return (
    <Section className="px-5 pt-5 text-fg-muted">
      <Text className="m-0 text-center text-[12px]">
        © {new Date().getFullYear()} Cég Neve
      </Text>
      <Text className="m-0 text-center text-[12px]">
        123 Utca Neve, Város, IR 12345
      </Text>
      <Text className="mt-2 text-center text-[12px]">
        <Link href={`${TANGO_URL}/privacy`} className="text-fg-muted">
          <u>{t("Footer.privacy_policy")}</u>
        </Link>
        {" • "}
        <Link href={`${TANGO_URL}/terms`} className="text-fg-muted">
          <u>{t("Footer.terms_of_service")}</u>
        </Link>
        {" • "}
        <Link href={`${TANGO_URL}/unsubscribe`} className="text-fg-muted">
          <u>{t("Footer.unsubscribe")}</u>
        </Link>
      </Text>
    </Section>
  );
}
