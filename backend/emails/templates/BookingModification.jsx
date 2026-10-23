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
import EmailTailwind from "../components/EmailTailwind";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function BookingModification() {
  return (
    <EmailTailwind>
      <Html>
        <Head />
        <Preview>{"{{ T .Lang `BookingModification.preview` . }}"}</Preview>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Heading as="h1" className="mb-4 text-2xl font-bold text-fg">
              {"{{ T .Lang `BookingModification.heading` . }}"}
            </Heading>

            <Text className="mb-5 text-[16px] text-fg-muted">
              {"{{ T .Lang `BookingModification.main_text` . }}"}
            </Text>

            <Section className="mb-6">
              <Row
                className="bg-bg-sunken p-3 border-b-2 border-b-border-strong"
                style={{
                  borderTopLeftRadius: "6px",
                  borderTopRightRadius: "6px",
                }}
              >
                <Column>
                  <Text className="m-0 text-[16px] font-semibold text-fg">
                    {
                      "{{ T .Lang `BookingModification.modification_details` . }}"
                    }
                  </Text>
                </Column>
              </Row>

              <Row className="p-3 border-b-2 border-b-border-strong">
                <Column className="w-[120px]">
                  <Text className="m-0 font-semibold text-fg-muted">
                    {"{{ T .Lang `BookingModification.original_text` . }}"}
                  </Text>
                </Column>
                <Column>
                  <Text
                    className="m-0 text-fg"
                    style={{ textDecoration: "line-through" }}
                  >
                    <span className="font-medium">{"{{ .OldDate }}"}</span>,{" "}
                    {"{{ .OldTime }}"}
                  </Text>
                </Column>
              </Row>

              <Row
                className="bg-success-subtle p-3"
                style={{
                  borderBottomLeftRadius: "6px",
                  borderBottomRightRadius: "6px",
                }}
              >
                <Column className="w-[120px]">
                  <Text className="m-0 font-semibold text-fg-muted">
                    {"{{ T .Lang `BookingModification.new_text` . }}"}
                  </Text>
                </Column>
                <Column>
                  <Text className="m-0 text-fg">
                    <span className="font-medium">{"{{ .Date }}"}</span>,{" "}
                    {"{{ .Time }}"}
                  </Text>
                </Column>
              </Row>
            </Section>

            <Section
              className="mb-6 bg-bg-sunken px-4 py-2 border-l-2 border-l-fg"
              style={{ borderRadius: "6px" }}
            >
              <Text className="mt-0 text-[16px] font-semibold text-fg">
                {"{{ T .Lang `BookingModification.unmodified_details` . }}"}
              </Text>
              <Text className="mt-0 mb-1 text-fg-muted">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingModification.timezone` . }}"}
                </span>{" "}
                {"{{ .TimeZone }}"}
              </Text>
              <Text className="mt-0 mb-1 text-fg-muted">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingModification.service_name` . }}"}
                </span>{" "}
                {"{{ .ServiceName }}"}
              </Text>
              <Text className="m-0 text-fg-muted">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingModification.location` . }}"}
                </span>{" "}
                {"{{ .Location }}"}
              </Text>
            </Section>

            <Text className="mb-6 text-[16px] text-fg-muted">
              {"{{ T .Lang `BookingModification.modification_note` . }}"}
            </Text>

            <Section className="mb-8 text-center">
              <Button
                href="{{ .ModifyLink }}"
                className="bg-primary px-6 py-3 text-center font-medium
                  text-fg-on-primary"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                {"{{ T .Lang `BookingModification.primary_button` . }}"}
              </Button>
            </Section>

            <Text className="mb-2 text-fg-muted">
              {"{{ T .Lang `BookingModification.extra_info` . }}"}
            </Text>

            <Text className="mb-6 text-fg-muted">
              {"{{ T .Lang `BookingModification.contact_us_note` . }}"}
            </Text>

            <Hr className="my-6 border-border" />

            <Footer />
          </Container>
        </Body>
      </Html>
    </EmailTailwind>
  );
}
