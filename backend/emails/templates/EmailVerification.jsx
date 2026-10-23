import React from "react";
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
  Text,
} from "react-email";
import EmailTailwind from "../components/EmailTailwind";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function EmailVerification() {
  return (
    <EmailTailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>{"{{ T .Lang `EmailVerification.preview` . }}"}</Preview>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section className="py-4">
              <Heading className="mb-2 text-center text-2xl font-bold text-fg">
                {"{{ T .Lang `EmailVerification.heading` . }}"}
              </Heading>

              <Text className="mb-6 text-center text-[16px] text-fg-muted">
                {"{{ T .Lang `EmailVerification.main_text` . }}"}
              </Text>

              <Section className="mb-6 w-auto text-center">
                <Row>
                  <Column align="center">
                    <Section className="mx-auto">
                      <Text
                        className="bg-primary-subtle px-14 py-2 font-mono text-2xl font-bold tracking-widest text-fg-primary border border-border-primary"
                        style={{
                          borderRadius: "5px",
                        }}
                      >
                        {"{{ .Code }}"}
                      </Text>
                    </Section>
                  </Column>
                </Row>
              </Section>

              <Text className="mb-16 text-center text-fg-muted">
                {"{{ T .Lang `EmailVerification.expiration_note` . }}"}
                <strong className="text-fg-primary">
                  {"{{ T .Lang `EmailVerification.expiration_note2` . }}"}
                </strong>
                {"{{ T .Lang `EmailVerification.expiration_note3` . }}"}
              </Text>
              <Text className="text-center text-[14px] text-fg-muted">
                {"{{ T .Lang `EmailVerification.contact_us_note` . }}"}
                <Link
                  href="mailto:support@company.com"
                  className="text-fg-primary"
                >
                  <u>support@company.com</u>
                </Link>
                {"{{ T .Lang `EmailVerification.contact_us_note2` . }}"}
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
