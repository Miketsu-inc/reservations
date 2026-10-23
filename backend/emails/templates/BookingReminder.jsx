import React from "react";
import {
  Body,
  Button,
  Container,
  Head,
  Heading,
  Hr,
  Html,
  Preview,
  Section,
  Text,
} from "react-email";
import { t, v } from "../template";
import EmailTailwind from "../components/EmailTailwind";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function BookingReminder() {
  return (
    <EmailTailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>{t("BookingReminder.preview")}</Preview>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Heading as="h1" className="mb-4 text-[22px] font-bold text-fg">
              {t("BookingReminder.heading")}
            </Heading>
            <Text className="mb-6 text-sm text-fg">
              {t("BookingReminder.main_text")}
            </Text>

            <Section
              className="mb-6 bg-bg-sunken pl-4 text-fg border-l-2 border-l-fg"
              style={{
                borderRadius: "6px",
              }}
            >
              <Text
                className="text-xs font-medium tracking-wide text-fg
                  uppercase"
              >
                {v("Date")}
              </Text>
              <Text className="mb-4 text-2xl font-bold text-fg">
                {v("Time")}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {t("BookingReminder.timezone")}
                </span>
                {v("TimeZone")}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {t("BookingReminder.service_name")}
                </span>
                {v("ServiceName")}
              </Text>
              <Text className="text-sm">
                <span className="font-semibold">
                  {t("BookingReminder.location")}
                </span>
                {v("Location")}
              </Text>
            </Section>

            <Section className="mb-8 text-center">
              <Button
                href={v("ModifyLink")}
                className="bg-primary px-4 py-3 text-center text-[14px]
                  font-medium text-fg-on-primary"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                {t("BookingReminder.primary_button")}
              </Button>
            </Section>

            <Text className="mb-3 text-sm">
              {t("BookingReminder.arrive_on_time_note")}
            </Text>

            <Text className="mb-6 text-xs text-fg-muted">
              {t("BookingReminder.contact_us_note")}
            </Text>

            <Hr className="mt-4 border-border" />

            <Footer />
          </Container>
        </Body>
      </Html>
    </EmailTailwind>
  );
}
