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
  Tailwind,
  Text,
} from "@react-email/components";
import React from "react";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function AppointmentCancellation() {
  return (
    <Tailwind>
      <Html>
        <Head />
        <Preview>{"{{ T .Lang `AppointmentCancellation.preview` . }}"}</Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Heading
              as="h1"
              className="mb-[16px] text-[22px] font-bold text-[#111111]"
            >
              {"{{ T .Lang `AppointmentCancellation.heading` . }}"}
            </Heading>

            <Text className="mb-6 text-sm">
              {"{{ T .Lang `AppointmentCancellation.main_text` . }}"}
            </Text>

            <Section
              className="mb-6 bg-gray-50 pt-3 pr-4 pb-4 pl-4 text-black"
              style={{
                borderLeft: "solid 2px #e53e3e",
                borderRadius: "6px",
              }}
            >
              <Row>
                <Column>
                  <Text className="m-0 text-xs font-medium tracking-wide text-gray-700 uppercase">
                    {"{{ .Date }}"}
                  </Text>
                </Column>
                <Column className="w-[100px]" align="right">
                  <Text
                    className="m-0 inline-block border-[2px] border-red-600 px-1.5 py-0.5 text-[14px]
                      font-medium text-red-600"
                    style={{ border: "solid 2px #dc2626", borderRadius: "6px" }}
                  >
                    {"{{ T .Lang `AppointmentCancellation.cancelled` . }}"}
                  </Text>
                </Column>
              </Row>

              <Text className="mb-4 text-2xl font-bold text-black">
                {"{{ .Time }}"}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `AppointmentCancellation.timezone` . }}"}
                </span>
                {"{{ .TimeZone }}"}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `AppointmentCancellation.service_name` . }}"}
                </span>
                {"{{ .ServiceName }}"}
              </Text>
              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `AppointmentCancellation.location` . }}"}
                </span>
                {"{{ .Location }}"}
              </Text>
            </Section>

            {"{{ if .Reason}}"}
            {/* Cancellation reason section */}
            <Section
              className="mb-6 bg-gray-50 p-[16px]"
              style={{
                borderRadius: "6px",
              }}
            >
              <Text className="m-0 mb-[8px] text-sm font-semibold">
                {
                  "{{ T .Lang `AppointmentCancellation.cancellation_reason` . }}"
                }
              </Text>
              <Text className="m-0 text-sm">{"{{ .Reason }}"}</Text>
            </Section>
            {"{{ end }}"}
            <Text className="mb-6 text-sm">
              {
                "{{ T .Lang `AppointmentCancellation.cancellation_reason_note` . }}"
              }
            </Text>

            <Section className="my-8 text-center">
              <Button
                href="{{ .NewAppointmentLink }}"
                className="bg-blue-600 px-4 py-3 text-center text-[14px] font-medium text-white"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                {"{{ T .Lang `AppointmentCancellation.primary_button` . }}"}
              </Button>
            </Section>

            <Text className="mb-6 text-sm">
              {"{{ T .Lang `AppointmentCancellation.contact_us_note` . }}"}
            </Text>

            <Text className="mb-6 text-xs text-gray-600">
              {"{{ T .Lang `AppointmentCancellation.apology` . }}"}
            </Text>

            <Hr className="mt-4" style={{ border: "1px solid #e5e7eb" }} />

            <Footer />
          </Container>
        </Body>
      </Html>
    </Tailwind>
  );
}
