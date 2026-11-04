import React from "react";
import {
  Body,
  Container,
  Head,
  Hr,
  Html,
  Link,
  Preview,
  Text,
} from "@react-email/components";
export default function Campaign() {
  return (
    <Html lang="pt-BR">
      <Head />
      <Preview>{"{{.Subject}}"}</Preview>
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
            maxWidth: "600px",
          }}
        >
          <Text style={{ fontSize: "14px", fontWeight: 700 }}>
            {"{{.Project.Name}}"} / newsletter
          </Text>
          <Hr style={{ borderColor: "#eee", margin: "24px 0" }} />
          <div>{"{{.Body}}"}</div>
          <Hr style={{ borderColor: "#eee", margin: "32px 0" }} />
          <Text style={{ fontSize: "12px", color: "#737373" }}>
            Você recebe este e-mail porque confirmou sua inscrição em{" "}
            {"{{.Project.Name}}"}.
          </Text>
          <Link
            href="{{.UnsubscribeURL}}"
            style={{ fontSize: "12px", color: "#737373" }}
          >
            Cancelar inscrição
          </Link>
        </Container>
      </Body>
    </Html>
  );
}
