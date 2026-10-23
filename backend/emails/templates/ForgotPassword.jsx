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
import { t, v } from "../template";
import EmailTailwind from "../components/EmailTailwind";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function ForgotPassword() {
  return (
    <EmailTailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>{t("ForgotPassword.preview")}</Preview>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section className="my-4 px-2">
              <Heading className="mb-2 text-center text-2xl font-bold text-fg">
                {t("ForgotPassword.heading")}
              </Heading>

              <Text className="mb-8 text-center text-[16px] text-fg-muted">
                {t("ForgotPassword.main_text")}
              </Text>

              <Section className="mb-8 text-center">
                <Button
                  href={v("PasswordLink")}
                  className="bg-primary px-5 py-3 font-semibold text-fg-on-primary"
                  style={{ borderRadius: "6px" }}
                >
                  {t("ForgotPassword.primary_button")}
                </Button>
              </Section>

              <Text className="mb-6 text-center text-fg-muted">
                {t("ForgotPassword.expiration_note")}
                <strong className="text-fg-primary">
                  {t("ForgotPassword.expiration_note2")}
                </strong>
                {t("ForgotPassword.expiration_note3")}
              </Text>

              <Text className="mt-2 text-center text-xs text-fg-muted">
                {t("ForgotPassword.ignore_email_note")}
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
