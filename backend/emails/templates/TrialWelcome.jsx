import {
  Body,
  Button,
  Container,
  Head,
  Heading,
  Hr,
  Html,
  Link,
  Preview,
  Section,
  Tailwind,
  Text,
} from "@react-email/components";
import React from "react";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function TrialWelcome() {
  return (
    <Html lang="hu" dir="ltr">
      <Head />
      <Preview>{"{{ T .Lang `TrialWelcome.preview` . }}"}</Preview>
      <Tailwind>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section>
              <Heading className="my-6 text-[22px] font-bold">
                {"{{ T .Lang `TrialWelcome.heading` . }}"}
              </Heading>

              <Text className="mb-6 text-[16px] text-gray-700">
                {"{{ T .Lang `TrialWelcome.main_text` . }}"}
              </Text>

              <Section className="my-8 text-center">
                <Button
                  className="bg-blue-600 px-6 py-3 text-center font-medium text-white"
                  href="https://app.example.com/dashboard"
                  style={{ boxSizing: "border-box", borderRadius: "6px" }}
                >
                  {"{{ T .Lang `TrialWelcome.primary_button` . }}"}
                </Button>
              </Section>

              <Text className="mb-6 text-gray-700">
                {"{{ T .Lang `TrialWelcome.contact_us_note` . }}"}
                <Link
                  href="https://app.example.com/tutorials"
                  className="font-medium text-blue-600"
                >
                  {"{{ T .Lang `TrialWelcome.contact_us_note2` . }}"}
                </Link>
                {"{{ T .Lang `TrialWelcome.contact_us_note3` . }}"}
              </Text>

              <Hr className="my-6 border-gray-200" />
            </Section>
            <Footer />
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
}
