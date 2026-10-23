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

const JABULANI_URL = process.env.JABULANI_URL;

export default function BookingConfirmation() {
  return (
    <EmailTailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>{"{{ T .Lang `BookingConfirmation.preview` . }}"}</Preview>
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
              {"{{ T .Lang `BookingConfirmation.heading` . }}"}
            </Heading>
            <Text className="mb-6 text-sm">
              {"{{ T .Lang `BookingConfirmation.main_text` . }}"}
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
                  {"{{ T .Lang `BookingConfirmation.timezone` . }}"}
                </span>
                {"{{ .TimeZone }}"}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingConfirmation.service_name` . }}"}
                </span>
                {"{{ .ServiceName }}"}
              </Text>
              <Text className="text-sm">
                <span className="font-semibold">
                  {"{{ T .Lang `BookingConfirmation.location` . }}"}
                </span>
                {"{{ .Location }}"}
              </Text>
            </Section>

            <Section className="mb-8 text-left">
              <Button
                href={`${JABULANI_URL}/calendar`}
                className="mr-2 inline-block w-fit bg-primary px-4 py-3
                  text-[14px] font-medium text-fg-on-primary"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                {"{{ T .Lang `BookingConfirmation.primary_button` . }}"}
              </Button>
              <Button
                href="{{ .ModifyLink }}"
                className="ml-2 inline-block w-fit bg-primary-subtle px-4 py-3
                  text-[14px] font-semibold text-fg-primary"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                {"{{ T .Lang `BookingConfirmation.secondary_button` . }}"}
              </Button>
            </Section>

            <Text className="mb-6 text-xs text-fg-muted">
              {"{{ T .Lang `BookingConfirmation.modification_note` . }}"}
            </Text>

            <Hr className="mt-4 border-border" />

            <Footer />
          </Container>
        </Body>
      </Html>
    </EmailTailwind>
  );
}
