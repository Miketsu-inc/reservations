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
                  <u>Adatvédelmi irányelvek</u>
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
