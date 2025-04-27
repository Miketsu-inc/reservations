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
            <LogoHeader />
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
            <Footer />
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
}
