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
  Tailwind,
  Text,
} from "@react-email/components";
import React from "react";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function TrialWelcome() {
  return (
    <Html lang="hu" dir="ltr">
      <Head />
      <Preview>
        Az ingyenes próbaidőszakod most elindult, nézz körül bátran!
      </Preview>
      <Tailwind>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section>
              <Heading className="my-6 text-[22px] font-bold">
                Üdvözlünk a company name-nél
              </Heading>

              <Text className="mb-6 text-[16px] text-gray-700">
                Köszönjük, hogy kipróbálod a szolgáltatásunkat! Az ingyenes
                próbaidőszakod most elkezdődött, kattints az alábbi gombra, és
                kezdj el felfedezni minden új lehetőséget!
              </Text>

              <Section className="my-8 text-center">
                <Button
                  className="bg-blue-600 px-6 py-3 text-center font-medium text-white"
                  href="https://app.example.com/dashboard"
                  style={{ boxSizing: "border-box", borderRadius: "6px" }}
                >
                  Felfedezés
                </Button>
              </Section>

              <Text className="mb-6 text-gray-700">
                Ha segítségre van szüksége az új funkciók használatával
                kapcsolatban, tekintse meg{" "}
                <Link
                  href="https://app.example.com/tutorials"
                  className="font-medium text-blue-600"
                >
                  oktatóanyagainkat
                </Link>{" "}
                vagy vegye fel a kapcsolatot ügyfélszolgálatunkkal a
                support@example.com címen.
              </Text>

              <Hr className="my-6 border-gray-200" />
            </Section>
            <Footer />
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
}
