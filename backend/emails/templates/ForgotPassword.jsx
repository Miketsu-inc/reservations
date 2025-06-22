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

export default function ForgotPassword() {
  return (
    <Tailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>{"{{ T .Lang `ForgotPassword.preview` . }}"}</Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section className="my-4 px-2">
              <Heading className="mb-2 text-center text-2xl font-bold text-gray-800">
                {"{{ T .Lang `ForgotPassword.heading` . }}"}
              </Heading>

              <Text className="mb-8 text-center text-[16px] text-gray-700">
                {"{{ T .Lang `ForgotPassword.main_text` . }}"}
              </Text>

              <Section className="mb-8 text-center">
                <Button
                  href="{{ .PasswordLink }}"
                  className="bg-blue-600 px-5 py-3 font-semibold text-white"
                  style={{ borderRadius: "6px" }}
                >
                  {"{{ T .Lang `ForgotPassword.primary_button` . }}"}
                </Button>
              </Section>

              <Text className="mb-6 text-center text-gray-600">
                {"{{ T .Lang `ForgotPassword.expiration_note` . }}"}
                <strong className="text-blue-600">
                  {"{{ T .Lang `ForgotPassword.expiration_note2` . }}"}
                </strong>
                {"{{ T .Lang `ForgotPassword.expiration_note3` . }}"}
              </Text>

              <Text className="mt-2 text-center text-xs text-gray-500">
                {"{{ T .Lang `ForgotPassword.ignore_email_note` . }}"}
              </Text>
              <Hr className="mt-2" style={{ border: "1px solid #e5e7b" }} />
            </Section>
            <Footer />
          </Container>
        </Body>
      </Html>
    </Tailwind>
  );
}
