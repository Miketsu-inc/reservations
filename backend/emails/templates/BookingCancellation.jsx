import React from "react";
import {
  Body,
  Button,
  Column,
  Container,
  Head,
  Heading,
  Hr,
  Html,
  Preview,
  Row,
  Section,
  Text,
} from "react-email";
import { If, t, v } from "../template";
import EmailTailwind from "../components/EmailTailwind";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function BookingCancellation() {
  return (
    <EmailTailwind>
      <Html>
        <Head />
        <Preview>{t("BookingCancellation.preview")}</Preview>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Heading
              as="h1"
              className="mb-[16px] text-[22px] font-bold text-fg"
            >
              {t("BookingCancellation.heading")}
            </Heading>

            <Text className="mb-6 text-sm">
              {t("BookingCancellation.main_text")}
            </Text>

            <Section
              className="mb-6 bg-bg-sunken pt-3 pr-4 pb-4 pl-4 text-fg border-l-2 border-l-danger"
              style={{
                borderRadius: "6px",
              }}
            >
              <Row>
                <Column>
                  <Text
                    className="m-0 text-xs font-medium tracking-wide
                      text-fg-muted uppercase"
                  >
                    {v("Date")}
                  </Text>
                </Column>
                <Column className="w-[100px]" align="right">
                  <Text
                    className="m-0 inline-block border-2 border-danger px-1.5
                      py-0.5 text-[14px] font-medium text-fg-danger"
                    style={{ borderRadius: "6px" }}
                  >
                    {t("BookingCancellation.cancelled")}
                  </Text>
                </Column>
              </Row>

              <Text className="mb-4 text-2xl font-bold text-fg">
                {v("Time")}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {t("BookingCancellation.timezone")}
                </span>
                {v("TimeZone")}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {t("BookingCancellation.service_name")}
                </span>
                {v("ServiceName")}
              </Text>
              <Text className="text-sm">
                <span className="font-semibold">
                  {t("BookingCancellation.location")}
                </span>
                {v("Location")}
              </Text>
            </Section>

            <If field="Reason">
              {/* Cancellation reason section */}
              <Section
                className="mb-6 bg-bg-sunken p-[16px]"
                style={{
                  borderRadius: "6px",
                }}
              >
                <Text className="m-0 mb-[8px] text-sm font-semibold">
                  {t("BookingCancellation.cancellation_reason")}
                </Text>
                <Text className="m-0 text-sm">{v("Reason")}</Text>
              </Section>
            </If>
            <Text className="mb-6 text-sm">
              {t("BookingCancellation.cancellation_reason_note")}
            </Text>

            <Section className="my-8 text-center">
              <Button
                href={v("NewBookingLink")}
                className="bg-primary px-4 py-3 text-center text-[14px]
                  font-medium text-fg-on-primary"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                {t("BookingCancellation.primary_button")}
              </Button>
            </Section>

            <Text className="mb-6 text-sm">
              {t("BookingCancellation.contact_us_note")}
            </Text>

            <Text className="mb-6 text-xs text-fg-muted">
              {t("BookingCancellation.apology")}
            </Text>

            <Hr className="mt-4 border-border" />

            <Footer />
          </Container>
        </Body>
      </Html>
    </EmailTailwind>
  );
}
