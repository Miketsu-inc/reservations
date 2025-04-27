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

export default function SubscriptionConfirmation() {
  return (
    <Html lang="hu" dir="ltr">
      <Head />
      <Preview>Az előfizetése sikeresen aktiválva</Preview>
      <Tailwind>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section>
              <Heading className="my-6 text-[22px] font-bold">
                Köszönjük az előfizetést!
              </Heading>

              <Text className="mb-6 text-[16px] text-gray-700">
                Örömmel értesítünk, hogy sikeresen frissítetted fiókodat a
                <span className="font-bold text-blue-600"> Pro</span> csomagra.
                Mostantól hozzáférhet a csomag prémium funkcióihoz, amelyek
                segítségével vállalkozása foglalási rendszere még hatékonyabb
                lesz.
              </Text>

              <Section
                className="my-8 bg-blue-50 px-4 py-1"
                style={{ borderRadius: "6px" }}
              >
                <Text className="mb-4 text-[18px] font-bold text-gray-800">
                  Új funkciók, amelyekhez most hozzáfér:
                </Text>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-gray-700">
                    • Korlátlan kereskedői profilok
                  </Text>
                  <Text className="m-0 ml-4 text-gray-600">
                    Hozzon létre bármennyi kereskedői profilt különböző
                    helyszínekhez
                  </Text>
                </Section>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-gray-700">
                    • Testreszabható foglalási oldal
                  </Text>
                  <Text className="m-0 ml-4 text-gray-600">
                    Egyedi színek, logó és dizájn a vállalkozása arculatához
                    igazítva
                  </Text>
                </Section>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-gray-700">
                    • Fejlett analitika
                  </Text>
                  <Text className="m-0 ml-4 text-gray-600">
                    Részletes kimutatások a foglalásokról és az ügyfelek
                    szokásairól
                  </Text>
                </Section>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-gray-700">
                    • Automatikus emlékeztetők
                  </Text>
                  <Text className="m-0 ml-4 text-gray-600">
                    Email és SMS értesítések az ügyfelek és a személyzet számára
                  </Text>
                </Section>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-gray-700">
                    • API hozzáférés
                  </Text>
                  <Text className="m-0 ml-4 text-gray-600">
                    Integrálja a foglalási rendszert a meglévő webhelyével
                  </Text>
                </Section>
              </Section>

              <Text className="mb-6 text-gray-700">
                Most már minden készen áll! Kattints az alábbi gombra, és kezdj
                el felfedezni minden új lehetőséget!
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
