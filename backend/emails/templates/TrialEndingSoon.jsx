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

export default function TrialEndingSoon() {
  return (
    <Html lang="hu" dir="ltr">
      <Head />
      <Preview>{"{{ T .Lang `TrialEndingSoon.preview` . }}"}</Preview>
      <EmailTailwind>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section>
              <Heading className="my-6 text-left text-[22px] font-bold text-fg">
                ⏰ {"{{ T .Lang `TrialEndingSoon.heading` . }}"}
              </Heading>

              <Text className="mb-6 text-[16px] text-fg-muted">
                {"{{ T .Lang `TrialEndingSoon.main_text` . }}"}
                <strong className="text-fg-primary">
                  {"{{ T .Lang `TrialEndingSoon.days_left` . }}"}
                </strong>
                {"{{ T .Lang `TrialEndingSoon.main_text2` . }}"}
              </Text>

              <Section className="my-8 text-center">
                <Button
                  className="bg-primary px-6 py-3 text-center font-medium text-fg-on-primary"
                  href="{{ .ManageLink }}"
                  style={{ boxSizing: "border-box", borderRadius: "6px" }}
                >
                  {"{{ T .Lang `TrialEndingSoon.primary_button` . }}"}
                </Button>
              </Section>

              <Text className="mb-6 text-fg-muted">
                {"{{ T .Lang `TrialEndingSoon.contact_us_note` . }}"}
                <Link
                  href="mailto:support@example.com"
                  className="font-medium text-fg-primary"
                >
                  support@example.com
                </Link>
                {"{{ T .Lang `TrialEndingSoon.contact_us_note2` . }}"}
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
