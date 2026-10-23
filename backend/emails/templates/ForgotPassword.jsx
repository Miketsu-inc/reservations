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

export default function ForgotPassword() {
  return (
    <EmailTailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>{"{{ T .Lang `ForgotPassword.preview` . }}"}</Preview>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section className="my-4 px-2">
              <Heading className="mb-2 text-center text-2xl font-bold text-fg">
                {"{{ T .Lang `ForgotPassword.heading` . }}"}
              </Heading>

              <Text className="mb-8 text-center text-[16px] text-fg-muted">
                {"{{ T .Lang `ForgotPassword.main_text` . }}"}
              </Text>

              <Section className="mb-8 text-center">
                <Button
                  href={"{{ .PasswordLink }}"}
                  className="bg-primary px-5 py-3 font-semibold text-fg-on-primary"
                  style={{ borderRadius: "6px" }}
                >
                  {"{{ T .Lang `ForgotPassword.primary_button` . }}"}
                </Button>
              </Section>

              <Text className="mb-6 text-center text-fg-muted">
                {"{{ T .Lang `ForgotPassword.expiration_note` . }}"}
                <strong className="text-fg-primary">
                  {"{{ T .Lang `ForgotPassword.expiration_note2` . }}"}
                </strong>
                {"{{ T .Lang `ForgotPassword.expiration_note3` . }}"}
              </Text>

              <Text className="mt-2 text-center text-xs text-fg-muted">
                {"{{ T .Lang `ForgotPassword.ignore_email_note` . }}"}
              </Text>
              <Hr className="mt-2 border-border" />
            </Section>
            <Footer />
          </Container>
        </Body>
      </Html>
    </EmailTailwind>
  );
}
