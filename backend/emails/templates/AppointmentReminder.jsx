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
  Tailwind,
  Text,
} from "@react-email/components";
import React from "react";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function AppointmentReminder() {
  return (
    <Tailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>{"{{ T .Lang `AppointmentReminder.preview` . }}"}</Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Heading
              as="h1"
              className="mb-4 text-[22px] font-bold text-[#111111]"
            >
              {"{{ T .Lang `AppointmentReminder.heading` . }}"}
            </Heading>
            <Text className="mb-6 text-sm text-black">
              {"{{ T .Lang `AppointmentReminder.main_text` . }}"}
            </Text>

            <Section
              className="mb-6 bg-gray-50 pl-4 text-black"
              style={{
                borderLeft: "solid 2px #000000",
                borderRadius: "6px",
              }}
            >
              <Text className="text-xs font-medium tracking-wide text-black uppercase">
                {"{{ .Date }}"}
              </Text>
              <Text className="mb-4 text-2xl font-bold text-black">
                {"{{ .Time }}"}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `AppointmentReminder.timezone` . }}"}
                </span>
                {"{{ .TimeZone }}"}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `AppointmentReminder.service_name` . }}"}
                </span>
                {"{{ .ServiceName }}"}
              </Text>
              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `AppointmentReminder.location` . }}"}
                </span>
                {"{{ .Location }}"}
              </Text>
            </Section>

            <Section className="mb-8 text-center">
              <Button
                href="{{ .ModifyLink }}"
                className="bg-blue-600 px-4 py-3 text-center text-[14px] font-medium text-white"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                {"{{ T .Lang `AppointmentReminder.primary_button` . }}"}
              </Button>
            </Section>

            <Text className="mb-3 text-sm">
              {"{{ T .Lang `AppointmentReminder.arrive_on_time_note` . }}"}
            </Text>

            <Text className="mb-6 text-xs text-gray-600">
              {"{{ T .Lang `AppointmentReminder.contact_us_note` . }}"}
            </Text>

            <Hr className="mt-4" style={{ border: "1px solid #e5e7eb" }} />

            <Footer />
          </Container>
        </Body>
      </Html>
    </Tailwind>
  );
}
