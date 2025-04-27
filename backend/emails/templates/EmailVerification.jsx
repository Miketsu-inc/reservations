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
  const code = "141592";

  return (
    <Tailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>Érvényesítsd az email címed</Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section className="py-4">
              <Heading className="mb-2 text-center text-2xl font-bold text-gray-800">
                Erősitsd meg az email címed
              </Heading>

              <Text className="mb-6 text-center text-[16px] text-gray-700">
                Az email címed megerősítéséhez nincs más dolgod mint hogy beírod
                az alábbi códot az arra kijelölt helyre.
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
                        {code}
                      </Text>
                    </Section>
                  </Column>
                </Row>
              </Section>

              <Text className="mb-16 text-center text-gray-600">
                Ez a kód <strong className="text-blue-600">10 percig</strong>{" "}
                érvényes. Ha nem te kezdeményezted a regisztrációt, egyszerűen
                figyelmen kívül hagyhatod ezt az emailt.
              </Text>
              <Text className="text-center text-[14px] text-gray-600">
                Kérdésed van? Írj nekünk a{" "}
                <Link
                  href="mailto:support@company.com"
                  className="text-blue-600"
                >
                  <u>support@company.com</u>
                </Link>{" "}
                címre.
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
