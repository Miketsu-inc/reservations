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
        <Preview>jelszó visszaállítási kérelem</Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section className="my-4 px-2">
              <Heading className="mb-2 text-center text-2xl font-bold text-gray-800">
                Elfelejtetted a jelszavadat?
              </Heading>

              <Text className="mb-8 text-center text-[16px] text-gray-700">
                Semmi gond, előfordul! Kattints az alábbi gombra az új jelszó
                beállitásához.
              </Text>

              <Section className="mb-8 text-center">
                <Button
                  href="{{ .PasswordLink }}"
                  className="bg-blue-600 px-5 py-3 font-semibold text-white"
                  style={{ borderRadius: "6px" }}
                >
                  Új jelszó beállitása
                </Button>
              </Section>

              <Text className="mb-6 text-center text-gray-600">
                Ez a link <strong className="text-blue-600">30 percig</strong>{" "}
                érvényes biztonsági okokból.
              </Text>

              <Text className="mt-2 text-center text-xs text-gray-500">
                Ha nem te kérted a jelszó visszaállítását, figyelmen kívül
                hagyhatod ezt az e-mailt. A fiókod biztonságban van.
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
