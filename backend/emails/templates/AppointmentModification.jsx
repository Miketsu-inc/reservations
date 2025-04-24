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

export default function AppointmentModification() {
  const oldDate = "Kedd, Április 22";
  const oldTime = "12:00 - 12:45";
  const newDate = "Szerda, Április 23";
  const newTime = "14:30 - 15:15";
  const serviceName = "Hajvágás és styling";
  const location = "Szépség Szalon, Fő utca 45, Budapest";
  const modificationReason =
    "A szakember más időpontban tudja csak biztosítani a szolgáltatást.";
  const timeZone = "GMT +2 (Central European Summer Time)";

  return (
    <Tailwind>
      <Html>
        <Head />
        <Preview>
          Az időpontja módosításra került - {newDate}, {newTime}
        </Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            {/* Header Section */}
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
                  <Text className="m-0 text-[16px] font-medium text-[#333333]">
                    Company Name
                  </Text>
                </Column>
              </Row>
            </Section>

            <Heading as="h1" className="mb-4 text-2xl font-bold text-[#111111]">
              Az időpontja módosításra került
            </Heading>

            <Text className="mb-5 text-[16px] text-gray-700">
              Tisztelt Ügyfelünk! Tájékoztatjuk, hogy a korábban foglalt
              időpontját módosítottuk. Az új időpont részleteit alább találja.
            </Text>

            <Section className="mb-6">
              <Row
                className="bg-gray-50 p-3"
                style={{
                  borderBottom: "solid 2px #d1d5dc",
                  borderTopLeftRadius: "6px",
                  borderTopRightRadius: "6px",
                }}
              >
                <Column>
                  <Text className="m-0 text-[16px] font-semibold text-black">
                    Időpont módosítás részletei
                  </Text>
                </Column>
              </Row>

              <Row
                className="p-3"
                style={{ borderBottom: "solid 2px #d1d5dc" }}
              >
                <Column className="w-[120px]">
                  <Text className="m-0 font-semibold text-gray-600">
                    Eredeti időpont:
                  </Text>
                </Column>
                <Column>
                  <Text
                    className="m-0 text-gray-800"
                    style={{ textDecoration: "line-through" }}
                  >
                    <span className="font-medium">{oldDate}</span>, {oldTime}
                  </Text>
                </Column>
              </Row>

              <Row
                className="bg-green-100 p-3"
                style={{
                  borderBottomLeftRadius: "6px",
                  borderBottomRightRadius: "6px",
                }}
              >
                <Column className="w-[120px]">
                  <Text className="m-0 font-semibold text-gray-600">
                    Új időpont:
                  </Text>
                </Column>
                <Column>
                  <Text className="m-0 text-gray-800">
                    <span className="font-medium">{newDate}</span>, {newTime}
                  </Text>
                </Column>
              </Row>
            </Section>

            <Section
              className="mb-6 bg-gray-50 px-4 py-2"
              style={{ borderLeft: "solid 2px #000000", borderRadius: "6px" }}
            >
              <Text className="mt-0 text-[16px] font-semibold text-black">
                A foglalás változatlan részei
              </Text>
              <Text className="mt-0 mb-1 text-gray-700">
                <span className="font-semibold">Szolgáltatás:</span>{" "}
                {serviceName}
              </Text>
              <Text className="mt-0 mb-1 text-gray-700">
                <span className="font-semibold">Helyszín:</span> {location}
              </Text>
              <Text className="m-0 text-gray-700">
                <span className="font-semibold">Időzóna:</span> {timeZone}
              </Text>
            </Section>

            <Section
              className="mb-6 bg-gray-50 p-4"
              style={{ borderRadius: "6px" }}
            >
              <Text className="mt-0 mb-2 font-semibold">
                Miért történt a módosítás?
              </Text>
              <Text className="m-0 text-gray-700">{modificationReason}</Text>
            </Section>

            <Text className="mb-6 text-[16px] text-gray-700">
              Ha az új időpont nem megfelelő Önnek, kérjük, válasszon egy
              másikat, vagy módosítsa a foglalását az alábbi gombra kattintva.
            </Text>

            <Section className="mb-8 text-center">
              <Button
                href="https://example.com/manage"
                className="bg-blue-600 px-6 py-3 text-center font-medium text-white"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                Időpont kezelése
              </Button>
            </Section>

            <Text className="mb-2 text-gray-700">
              Az új időpont automatikusan bekerült a naptárába, amennyiben
              korábban elfogadta a naptárbejegyzést.
            </Text>

            <Text className="mb-6 text-gray-700">
              Ha kérdése van, vagy segítségre van szüksége, kérjük, vegye fel a
              kapcsolatot a +36 1 234 5678 telefonszámon, vagy válaszoljon erre
              az e-mailre.
            </Text>

            <Hr className="my-6" style={{ border: "1px solid #e5e7eb" }} />

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
