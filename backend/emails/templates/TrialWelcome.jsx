import React from "react";
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
  Text,
} from "react-email";
import EmailTailwind from "../components/EmailTailwind";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function TrialWelcome() {
  return (
    <Html lang="hu" dir="ltr">
      <Head />
      <Preview>{"{{ T .Lang `TrialWelcome.preview` . }}"}</Preview>
      <EmailTailwind>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section>
              <Heading className="my-6 text-[22px] font-bold">
                {"{{ T .Lang `TrialWelcome.heading` . }}"}
              </Heading>

              <Text className="mb-6 text-[16px] text-fg-muted">
                {"{{ T .Lang `TrialWelcome.main_text` . }}"}
              </Text>

              <Section className="my-8 text-center">
                <Button
                  className="bg-primary px-6 py-3 text-center font-medium text-fg-on-primary"
                  href="https://app.example.com/dashboard"
                  style={{ boxSizing: "border-box", borderRadius: "6px" }}
                >
                  {"{{ T .Lang `TrialWelcome.primary_button` . }}"}
                </Button>
              </Section>

              <Text className="mb-6 text-fg-muted">
                {"{{ T .Lang `TrialWelcome.contact_us_note` . }}"}
                <Link
                  href="https://app.example.com/tutorials"
                  className="font-medium text-fg-primary"
                >
                  {"{{ T .Lang `TrialWelcome.contact_us_note2` . }}"}
                </Link>
                {"{{ T .Lang `TrialWelcome.contact_us_note3` . }}"}
              </Text>

              <Hr className="my-6 border-border" />
            </Section>
            <Footer />
          </Container>
        </Body>
      </EmailTailwind>
    </Html>
  );
}
