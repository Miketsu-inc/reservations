import { Link, Section, Text } from "@react-email/components";
import React from "react";

void React;

export default function Footer() {
  return (
    <Section className="px-5 pt-5 text-gray-500">
      <Text className="m-0 text-center text-[12px]">
        © {new Date().getFullYear()} Cég Neve
      </Text>
      <Text className="m-0 text-center text-[12px]">
        123 Utca Neve, Város, IR 12345
      </Text>
      <Text className="mt-2 text-center text-[12px]">
        <Link href="http://localhost:5173/privacy" className="text-gray-500">
          <u>{"{{ T .Lang `Footer.privacy_policy` }}"}</u>
        </Link>
        {" • "}
        <Link href="http://localhost:5173/terms" className="text-gray-500">
          <u>{"{{ T .Lang `Footer.terms_of_service` }}"}</u>
        </Link>
        {" • "}
        <Link
          href="http://localhost:5173/unsubscribe"
          className="text-gray-500"
        >
          <u>{"{{ T .Lang `Footer.unsubscribe` }}"}</u>
        </Link>
      </Text>
    </Section>
  );
}
