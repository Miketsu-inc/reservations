import {
  Body,
  Button,
  Column,
  Container,
  Head,
  Heading,
  Hr,
  Html,
  Img,
  Link,
  Preview,
  Row,
  Section,
  Tailwind,
  Text,
} from "@react-email/components";
import React from "react";
import ReactDom from "react-dom";

void (React, ReactDom);

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
            <Section>
              <Row className="m-0 mt-4">
                <Column className="w-12" align="left">
                  <Img
                    src="https://dummyimage.com/40x40/d156c3/000000.jpg"
                    alt="App Logo"
                    className="w-12"
                    style={{ borderRadius: "40px", filter: "none !important" }}
                  />
                </Column>
                <Column align="left" className="pl-3">
                  <Text className="m-0 text-[16px] font-medium text-[#333333]">
                    Company Name
                  </Text>
                </Column>
              </Row>
            </Section>

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
                  href="https://example.com/manage"
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
            <Section className="px-5 pt-5 text-gray-500">
              <Text className="m-0 text-center text-[12px]">
                © {new Date().getFullYear()} Cég Neve
              </Text>
              <Text className="m-0 text-center text-[12px]">
                123 Utca Neve, Város, IR 12345
              </Text>
              <Text className="mt-2 text-center text-[12px]">
                <Link
                  href="https://company.com/privacy"
                  className="text-gray-500"
                >
                  <u>Privacy Policy</u>
                </Link>
                {" • "}
                <Link
                  href="https://company.com/terms"
                  className="text-gray-500"
                >
                  <u>Terms of Service</u>
                </Link>
                {" • "}
                <Link
                  href="https://company.com/unsubscribe"
                  className="text-gray-500"
                >
                  <u>Unsubscribe</u>
                </Link>
              </Text>
            </Section>
          </Container>
        </Body>
      </Html>
    </Tailwind>
  );
}
