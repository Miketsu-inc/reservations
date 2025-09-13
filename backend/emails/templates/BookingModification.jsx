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

export default function BookingModification() {
  return (
    <Tailwind>
      <Html>
        <Head />
        <Preview>{"{{ T .Lang `BookingModification.preview` . }}"}</Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Heading as="h1" className="mb-4 text-2xl font-bold text-[#111111]">
              {"{{ T .Lang `BookingModification.heading` . }}"}
            </Heading>

            <Text className="mb-5 text-[16px] text-gray-700">
              {"{{ T .Lang `BookingModification.main_text` . }}"}
            </Text>

            <Section className="mb-6">
              <Row
                className="bg-gray-50 p-3"
                style={{
                  borderBottom: "solid 2px #d1d5dc",
                  borderTopLeftRadius: "6px",
                  borderTopRightRadius: "6px",
                }}
              >
                <Column>
                  <Text className="m-0 text-[16px] font-semibold text-black">
                    {
                      "{{ T .Lang `BookingModification.modification_details` . }}"
                    }
                  </Text>
                </Column>
              </Row>

              <Row
                className="p-3"
                style={{ borderBottom: "solid 2px #d1d5dc" }}
              >
                <Column className="w-[120px]">
                  <Text className="m-0 font-semibold text-gray-600">
                    {"{{ T .Lang `BookingModification.original_text` . }}"}
                  </Text>
                </Column>
                <Column>
                  <Text
                    className="m-0 text-gray-800"
                    style={{ textDecoration: "line-through" }}
                  >
                    <span className="font-medium">{"{{ .OldDate }}"}</span>,{" "}
                    {"{{ .OldTime }}"}
                  </Text>
                </Column>
              </Row>

              <Row
                className="bg-green-100 p-3"
                style={{
                  borderBottomLeftRadius: "6px",
                  borderBottomRightRadius: "6px",
                }}
              >
                <Column className="w-[120px]">
                  <Text className="m-0 font-semibold text-gray-600">
                    {"{{ T .Lang `BookingModification.new_text` . }}"}
                  </Text>
                </Column>
                <Column>
                  <Text className="m-0 text-gray-800">
                    <span className="font-medium">{"{{ .Date }}"}</span>,{" "}
                    {"{{ .Time }}"}
                  </Text>
                </Column>
              </Row>
            </Section>

            <Section
              className="mb-6 bg-gray-50 px-4 py-2"
              style={{ borderLeft: "solid 2px #000000", borderRadius: "6px" }}
            >
              <Text className="mt-0 text-[16px] font-semibold text-black">
                {"{{ T .Lang `BookingModification.unmodified_details` . }}"}
              </Text>
              <Text className="mt-0 mb-1 text-gray-700">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingModification.timezone` . }}"}
                </span>{" "}
                {"{{ .TimeZone }}"}
              </Text>
              <Text className="mt-0 mb-1 text-gray-700">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingModification.service_name` . }}"}
                </span>{" "}
                {"{{ .ServiceName }}"}
              </Text>
              <Text className="m-0 text-gray-700">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingModification.location` . }}"}
                </span>{" "}
                {"{{ .Location }}"}
              </Text>
            </Section>

            <Text className="mb-6 text-[16px] text-gray-700">
              {"{{ T .Lang `BookingModification.modification_note` . }}"}
            </Text>

            <Section className="mb-8 text-center">
              <Button
                href="{{ .ModifyLink }}"
                className="bg-blue-600 px-6 py-3 text-center font-medium
                  text-white"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                {"{{ T .Lang `BookingModification.primary_button` . }}"}
              </Button>
            </Section>

            <Text className="mb-2 text-gray-700">
              {"{{ T .Lang `BookingModification.extra_info` . }}"}
            </Text>

            <Text className="mb-6 text-gray-700">
              {"{{ T .Lang `BookingModification.contact_us_note` . }}"}
            </Text>

            <Hr className="my-6" style={{ border: "1px solid #e5e7eb" }} />

            <Footer />
          </Container>
        </Body>
      </Html>
    </Tailwind>
  );
}
