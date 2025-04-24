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
            <Section>
              <Row className="m-0 mt-3">
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
