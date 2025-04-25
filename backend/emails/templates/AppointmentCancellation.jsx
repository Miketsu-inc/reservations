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

export default function AppointmentCancellation() {
  const date = "Szerda, Április 23";
  const time = "14:30 - 15:15";
  const serviceName = "Hajvágás és styling";
  const location = "Szépség Szalon, Fő utca 45, Budapest";
  const timeZone = "GMT +2 (Central European Summer Time)";
  const cancellationReason =
    "A szakember váratlanul megbetegedett, ezért nem tudja ellátni a szolgáltatást a megadott időpontban.";

  return (
    <Tailwind>
      <Html>
        <Head />
        <Preview>Az időpontja lemondásra került</Preview>
        <Body className="bg-gray-100 font-sans text-black">
          <Container
            className="mx-auto max-w-md bg-white p-4"
            style={{ borderRadius: "6px" }}
          >
            <Section>
              <Row className="m-0 mt-4">
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

            <Heading
              as="h1"
              className="mb-[16px] text-[22px] font-bold text-[#111111]"
            >
              Az időpontja lemondásra került
            </Heading>

            <Text className="mb-6 text-sm">
              Sajnálattal tájékoztatjuk, hogy az alábbi időpontját előre nem
              látható körülmények miatt le kellett mondanunk:
            </Text>

            <Section
              className="mb-6 bg-gray-50 pt-3 pr-4 pb-4 pl-4 text-black"
              style={{
                borderLeft: "solid 2px #e53e3e",
                borderRadius: "6px",
              }}
            >
              <Row>
                <Column>
                  <Text className="m-0 text-xs font-medium tracking-wide text-gray-700 uppercase">
                    {date}
                  </Text>
                </Column>
                <Column className="w-[100px]" align="right">
                  <Text
                    className="m-0 inline-block border-[2px] border-red-600 px-1.5 py-0.5 text-[14px]
                      font-medium text-red-600"
                    style={{ border: "solid 2px #dc2626", borderRadius: "6px" }}
                  >
                    LEMONDVA
                  </Text>
                </Column>
              </Row>

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

            {/* Cancellation reason section */}
            <Section
              className="mb-6 bg-gray-50 p-[16px]"
              style={{
                borderRadius: "6px",
              }}
            >
              <Text className="m-0 mb-[8px] text-sm font-semibold">
                A lemondás oka:
              </Text>
              <Text className="m-0 text-sm">{cancellationReason}</Text>
            </Section>

            <Text className="mb-6 text-sm">
              Elnézést kérünk a kellemetlenségért. Értékeljük az Ön idejét, és
              szeretnénk lehetőséget biztosítani egy új időpont egyszerű
              foglalására.
            </Text>

            <Section className="my-8 text-center">
              <Button
                href="https://example.com/manage"
                className="bg-blue-600 px-4 py-3 text-center text-[14px] font-medium text-white"
                style={{
                  boxSizing: "border-box",
                  borderRadius: "6px",
                }}
              >
                Új időpont foglalása
              </Button>
            </Section>

            <Text className="mb-6 text-sm">
              Amennyiben kérdése lenne vagy segítségre van szüksége, kérjük,
              vegye fel velünk a kapcsolatot a +36 1 234 5678 telefonszámon vagy
              válaszoljon erre az e-mailre.
            </Text>

            <Text className="mb-6 text-xs text-gray-600">
              Köszönjük megértését és elnézést kérünk az esetleges
              kellemetlenségért.
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
