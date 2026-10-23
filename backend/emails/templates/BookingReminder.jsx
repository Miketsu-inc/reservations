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
import EmailTailwind from "../components/EmailTailwind";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function BookingReminder() {
  return (
    <EmailTailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>{"{{ T .Lang `BookingReminder.preview` . }}"}</Preview>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Heading as="h1" className="mb-4 text-[22px] font-bold text-fg">
              {"{{ T .Lang `BookingReminder.heading` . }}"}
            </Heading>
            <Text className="mb-6 text-sm text-fg">
              {"{{ T .Lang `BookingReminder.main_text` . }}"}
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
                {"{{ .Date }}"}
              </Text>
              <Text className="mb-4 text-2xl font-bold text-fg">
                {"{{ .Time }}"}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingReminder.timezone` . }}"}
                </span>
                {"{{ .TimeZone }}"}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingReminder.service_name` . }}"}
                </span>
                {"{{ .ServiceName }}"}
              </Text>
              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingReminder.location` . }}"}
                </span>
                {"{{ .Location }}"}
              </Text>
            </Section>

            <Section className="mb-8 text-center">
              <Button
                href="{{ .ModifyLink }}"
                className="bg-primary px-4 py-3 text-center text-[14px]
                  font-medium text-fg-on-primary"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                {"{{ T .Lang `BookingReminder.primary_button` . }}"}
              </Button>
            </Section>

            <Text className="mb-3 text-sm">
              {"{{ T .Lang `BookingReminder.arrive_on_time_note` . }}"}
            </Text>

            <Text className="mb-6 text-xs text-fg-muted">
              {"{{ T .Lang `BookingReminder.contact_us_note` . }}"}
            </Text>

            <Hr className="mt-4 border-border" />

            <Footer />
          </Container>
        </Body>
      </Html>
    </EmailTailwind>
  );
}
