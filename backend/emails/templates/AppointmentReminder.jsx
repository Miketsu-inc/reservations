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

export default function AppointmentReminder() {
  return (
    <Tailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>Emlékeztető a közelgő időpontjáról</Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <LogoHeader />
            <Heading
              as="h1"
              className="mb-4 text-[22px] font-bold text-[#111111]"
            >
              Emlékeztető a közelgő időpontjáról!
            </Heading>
            <Text className="mb-6 text-sm text-black">
              Szeretnénk emlékeztetni, hogy hamarosan esedékes a foglalása. Íme
              az időpontjával kapcsolatos információk:
            </Text>

            <Section
              className="mb-6 bg-gray-50 pl-4 text-black"
              style={{
                borderLeft: "solid 2px #000000",
                borderRadius: "6px",
              }}
            >
              <Text className="text-xs font-medium tracking-wide text-black uppercase">
                {"{{ .Date }}"}
              </Text>
              <Text className="mb-4 text-2xl font-bold text-black">
                {"{{ .Time }}"}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">Időzóna: </span>
                {"{{ .TimeZone }}"}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">Szolgáltatás: </span>
                {"{{ .ServiceType }}"}
              </Text>
              <Text className="text-sm">
                <span className="font-semibold">Helyszín: </span>
                {"{{ .Location }}"}
              </Text>
            </Section>

            <Section className="mb-8 text-center">
              <Button
                href="{{ .ModifyLink }}"
                className="bg-blue-600 px-4 py-3 text-center text-[14px] font-medium text-white"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                Időpont kezelése
              </Button>
            </Section>

            <Text className="mb-3 text-sm">
              Kérjük, érkezzen pontosan a foglalt időpontra. Ha bármilyen
              kérdése van, vagy módosítaná időpontját, kérjük, vegye fel velünk
              a kapcsolatot.
            </Text>

            <Text className="mb-6 text-xs text-gray-600">
              Ha bármilyen változtatást szeretne eszközölni az időpontjával
              kapcsolatban, kérjük, lépjen kapcsolatba velünk legalább 24 órával
              a tervezett időpont előtt.
            </Text>

            <Hr className="mt-4" style={{ border: "1px solid #e5e7eb" }} />

            <Footer />
          </Container>
        </Body>
      </Html>
    </Tailwind>
  );
}
