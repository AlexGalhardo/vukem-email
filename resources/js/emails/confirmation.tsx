import React from "react";
import {
  Body,
  Button,
  Container,
  Head,
  Heading,
  Hr,
  Html,
  Preview,
  Text,
} from "@react-email/components";
export default function Confirmation() {
  return (
    <Html lang="en">
      <Head />
      <Preview>Confirm your email to receive updates.</Preview>
      <Body
        style={{
          backgroundColor: "#f6f6f6",
          fontFamily: "Arial,sans-serif",
          padding: "40px 12px",
        }}
      >
        <Container
          style={{
            backgroundColor: "#fff",
            padding: "40px",
            borderRadius: "12px",
            maxWidth: "560px",
          }}
        >
          <Text style={{ fontSize: "14px", fontWeight: 700 }}>
            {"{{.Project.Name}}"} / newsletter
          </Text>
          <Heading style={{ fontSize: "28px" }}>One more step.</Heading>
          <Text style={{ color: "#525252", lineHeight: "26px" }}>
            You are one click away from updates from{" "}
            {"{{.Project.Name}}"}. Confirm that this email belongs to you to
            join the list.
          </Text>
          <Button
            href="{{.URL}}"
            style={{
              backgroundColor: "#171717",
              color: "#fff",
              padding: "14px 24px",
              borderRadius: "6px",
            }}
          >
            Confirm my subscription →
          </Button>
          <Hr style={{ borderColor: "#eee", margin: "32px 0" }} />
          <Text style={{ fontSize: "12px", color: "#737373" }}>
            This link expires in 24 hours. If you did not request this
            subscription, ignore this email.
          </Text>
        </Container>
      </Body>
    </Html>
  );
}
