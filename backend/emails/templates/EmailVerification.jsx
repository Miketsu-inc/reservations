import {
  Body,
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

export default function EmailVerification() {
  const code = "141592";

  return (
    <Tailwind>
      <Html lang="hu" dir="ltr">
        <Head />
        <Preview>Érvényesítsd az email címed</Preview>
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
                  <Text className="m-0 text-[16px] font-medium text-[#333333]">
                    Company Name
                  </Text>
                </Column>
              </Row>
            </Section>
            <Section className="py-4">
              <Heading className="mb-2 text-center text-2xl font-bold text-gray-800">
                Erősitsd meg az email címed
              </Heading>

              <Text className="mb-6 text-center text-[16px] text-gray-700">
                Az email címed megerősítéséhez nincs más dolgod mint hogy beírod
                az alábbi códot az arra kijelölt helyre.
              </Text>

              <Section className="mb-6 w-auto text-center">
                <Row>
                  <Column align="center">
                    <Section className="mx-auto">
                      <Text
                        className="bg-blue-50 px-14 py-2 font-mono text-2xl font-bold tracking-widest text-blue-700"
                        style={{
                          border: "solid 1px #1447e6",
                          borderRadius: "5px",
                        }}
                      >
                        {code}
                      </Text>
                    </Section>
                  </Column>
                </Row>
              </Section>

              <Text className="mb-16 text-center text-gray-600">
                Ez a kód <strong className="text-blue-600">10 percig</strong>{" "}
                érvényes. Ha nem te kezdeményezted a regisztrációt, egyszerűen
                figyelmen kívül hagyhatod ezt az emailt.
              </Text>
              <Text className="text-center text-[14px] text-gray-600">
                Kérdésed van? Írj nekünk a{" "}
                <Link
                  href="mailto:support@company.com"
                  className="text-blue-600"
                >
                  <u>support@company.com</u>
                </Link>{" "}
                címre.
              </Text>
              <Hr className="mt-2" style={{ border: "1px solid #e5e7b" }} />
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
                  <u>Privacy Policy</u>
                </Link>
                {" • "}
                <Link
                  href="https://company.com/terms"
                  className="text-gray-500"
                >
                  <u>Terms of Service</u>
                </Link>
                {" • "}
                <Link
                  href="https://company.com/unsubscribe"
                  className="text-gray-500"
                >
                  <u>Unsubscribe</u>
                </Link>
              </Text>
            </Section>
          </Container>
        </Body>
      </Html>
    </Tailwind>
  );
}
