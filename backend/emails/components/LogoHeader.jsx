import { Column, Img, Row, Section, Text } from "@react-email/components";
import React from "react";

void React;

export default function LogoHeader() {
  return (
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
  );
}
