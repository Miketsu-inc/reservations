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

export default function TrialEndingSoonEmail({ manageLink }) {
  return (
    <Html lang="hu" dir="ltr">
      <Head />
      <Preview>Az ingyenes próbaidőszakod hamarosan lejár!</Preview>
      <Tailwind>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <Section>
              <Row className="m-0 mt-3">
                <Column className="w-16" align="left">
                  <Img
                    src="https://dummyimage.com/40x40/d156c3/000000.jpg"
                    alt="App Logo"
                    className="w-14"
                    style={{ borderRadius: "40px" }}
                  />
                </Column>
                <Column align="left" className="pl-3">
                  <Text className="m-0 text-[16px] font-medium text-[#333333]">
                    Company Name
                  </Text>
                </Column>
              </Row>
            </Section>

            <Section>
              <Heading className="my-6 text-left text-[22px] font-bold text-gray-800">
                ⏰ A próbaidőszakod hamarosan lejár!
              </Heading>

              <Text className="mb-6 text-[16px] text-gray-700">
                Reméljük, élvezted az eddigi szolgáltatásunkat! Az ingyenes
                próbaidőszakod{" "}
                <strong className="text-blue-600">2 napon belül </strong> lejár.
                Ha nem mondod le az előfizetést a próbaidőszak vége előtt,
                automatikusan aktiváljuk számodra a havi előfizetést.
              </Text>

              <Section className="my-8 text-center">
                <Button
                  className="bg-blue-600 px-6 py-3 text-center font-medium text-white"
                  href={manageLink}
                  style={{ boxSizing: "border-box", borderRadius: "6px" }}
                >
                  Előfizetés kezelése
                </Button>
              </Section>

              <Text className="mb-6 text-gray-700">
                Ha bármilyen kérdésed van, vagy segítségre van szükséged, ne
                habozz kapcsolatba lépni velünk a{" "}
                <Link
                  href="mailto:support@example.com"
                  className="font-medium text-blue-600"
                >
                  support@example.com
                </Link>{" "}
                címen.
              </Text>

              <Hr className="my-6 border-gray-200" />
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
                  <u>Adatvédelem</u>
                </Link>
                {" • "}
                <Link
                  href="https://company.com/terms"
                  className="text-gray-500"
                >
                  <u>Felhasználási feltételek</u>
                </Link>
                {" • "}
                <Link
                  href="https://company.com/unsubscribe"
                  className="text-gray-500"
                >
                  <u>Leiratkozás</u>
                </Link>
              </Text>
            </Section>
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
}
