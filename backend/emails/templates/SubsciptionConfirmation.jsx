import React from "react";
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
  Text,
} from "react-email";
import EmailTailwind from "../components/EmailTailwind";
import Footer from "../components/Footer";
import LogoHeader from "../components/LogoHeader";

void React;

export default function SubscriptionConfirmation() {
  return (
    <Html lang="hu" dir="ltr">
      <Head />
      <Preview>Az előfizetése sikeresen aktiválva</Preview>
      <EmailTailwind>
        <Body className="bg-canvas font-sans text-fg">
          <Container
            className="mx-auto max-w-md bg-bg-raised p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Section>
              <Heading className="my-6 text-[22px] font-bold">
                Köszönjük az előfizetést!
              </Heading>

              <Text className="mb-6 text-[16px] text-fg-muted">
                Örömmel értesítünk, hogy sikeresen frissítetted fiókodat a
                <span className="font-bold text-fg-primary"> Pro</span>{" "}
                csomagra. Mostantól hozzáférhet a csomag prémium funkcióihoz,
                amelyek segítségével vállalkozása foglalási rendszere még
                hatékonyabb lesz.
              </Text>

              <Section
                className="my-8 bg-primary-subtle px-4 py-1"
                style={{ borderRadius: "6px" }}
              >
                <Text className="mb-4 text-[18px] font-bold text-fg">
                  Új funkciók, amelyekhez most hozzáfér:
                </Text>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-fg-muted">
                    • Korlátlan kereskedői profilok
                  </Text>
                  <Text className="m-0 ml-4 text-fg-muted">
                    Hozzon létre bármennyi kereskedői profilt különböző
                    helyszínekhez
                  </Text>
                </Section>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-fg-muted">
                    • Testreszabható foglalási oldal
                  </Text>
                  <Text className="m-0 ml-4 text-fg-muted">
                    Egyedi színek, logó és dizájn a vállalkozása arculatához
                    igazítva
                  </Text>
                </Section>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-fg-muted">
                    • Fejlett analitika
                  </Text>
                  <Text className="m-0 ml-4 text-fg-muted">
                    Részletes kimutatások a foglalásokról és az ügyfelek
                    szokásairól
                  </Text>
                </Section>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-fg-muted">
                    • Automatikus emlékeztetők
                  </Text>
                  <Text className="m-0 ml-4 text-fg-muted">
                    Email és SMS értesítések az ügyfelek és a személyzet számára
                  </Text>
                </Section>

                <Section className="mb-4">
                  <Text className="m-0 mb-2 text-[16px] font-medium text-fg-muted">
                    • API hozzáférés
                  </Text>
                  <Text className="m-0 ml-4 text-fg-muted">
                    Integrálja a foglalási rendszert a meglévő webhelyével
                  </Text>
                </Section>
              </Section>

              <Text className="mb-6 text-fg-muted">
                Most már minden készen áll! Kattints az alábbi gombra, és kezdj
                el felfedezni minden új lehetőséget!
              </Text>

              <Section className="my-8 text-center">
                <Button
                  className="bg-primary px-6 py-3 text-center font-medium text-fg-on-primary"
                  href="https://app.example.com/dashboard"
                  style={{ boxSizing: "border-box", borderRadius: "6px" }}
                >
                  Felfedezés
                </Button>
              </Section>

              <Text className="mb-6 text-fg-muted">
                Ha segítségre van szüksége az új funkciók használatával
                kapcsolatban, tekintse meg{" "}
                <Link
                  href="https://app.example.com/tutorials"
                  className="font-medium text-fg-primary"
                >
                  oktatóanyagainkat
                </Link>{" "}
                vagy vegye fel a kapcsolatot ügyfélszolgálatunkkal a
                support@example.com címen.
              </Text>

              <Hr className="my-6 border-border" />
            </Section>
            <Footer />
          </Container>
        </Body>
      </EmailTailwind>
    </Html>
  );
}
