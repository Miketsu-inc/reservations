import {
  Body,
  Column,
  Container,
  Head,
  Heading,
  Hr,
  Html,
  Link,
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

export default function EmailVerification() {
  return (
    <Tailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>{"{{ T .Lang `EmailVerification.preview` . }}"}</Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section className="py-4">
              <Heading className="mb-2 text-center text-2xl font-bold text-gray-800">
                {"{{ T .Lang `EmailVerification.heading` . }}"}
              </Heading>

              <Text className="mb-6 text-center text-[16px] text-gray-700">
                {"{{ T .Lang `EmailVerification.main_text` . }}"}
              </Text>

              <Section className="mb-6 w-auto text-center">
                <Row>
                  <Column align="center">
                    <Section className="mx-auto">
                      <Text
                        className="bg-blue-50 px-14 py-2 font-mono text-2xl font-bold tracking-widest text-blue-700"
                        style={{
                          border: "solid 1px #1447e6",
                          borderRadius: "5px",
                        }}
                      >
                        {"{{ .Code }}"}
                      </Text>
                    </Section>
                  </Column>
                </Row>
              </Section>

              <Text className="mb-16 text-center text-gray-600">
                {"{{ T .Lang `EmailVerification.expiration_note` . }}"}
                <strong className="text-blue-600">
                  {"{{ T .Lang `EmailVerification.expiration_note2` . }}"}
                </strong>
                {"{{ T .Lang `EmailVerification.expiration_note3` . }}"}
              </Text>
              <Text className="text-center text-[14px] text-gray-600">
                {"{{ T .Lang `EmailVerification.contact_us_note` . }}"}
                <Link
                  href="mailto:support@company.com"
                  className="text-blue-600"
                >
                  <u>support@company.com</u>
                </Link>
                {"{{ T .Lang `EmailVerification.contact_us_note2` . }}"}
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
