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

export default function AppointmentReminder() {
  const date = "Szerda, Április 23";
  const time = "14:30 - 15:15";
  const serviceName = "Hajvágás és styling";
  const location = "Szépség Szalon, Fő utca 45, Budapest";
  const timeZone = "GMT +2 (Central European Summer Time)";

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
            <Section>
              <Row className="m-0 mt-4">
                <Column className="w-12" align="left">
                  <Img
                    src="https://dummyimage.com/40x40/d156c3/000000.jpg"
                    alt="App Logo"
                    className="w-12"
                    style={{ borderRadius: "40px" }}
                  />
                </Column>
                <Column align="left" className="pl-3">
                  <Text className="text-[16px] font-medium text-[#333333]">
                    Company Name
                  </Text>
                </Column>
              </Row>
            </Section>

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
                {date}
              </Text>
              <Text className="mb-4 text-2xl font-bold text-black">{time}</Text>

              <Text className="text-sm">
                <span className="font-semibold">Időzóna:</span> {timeZone}
              </Text>

              <Text className="text-sm">
                <span className="font-semibold">Szolgáltatás:</span>{" "}
                {serviceName}
              </Text>
              <Text className="text-sm">
                <span className="font-semibold">Helyszín:</span> {location}
              </Text>
            </Section>

            <Section className="mb-8 text-center">
              <Button
                href="https://example.com/manage"
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
      </Html>
    </Tailwind>
  );
}
