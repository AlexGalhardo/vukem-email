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
    <Html lang="pt-BR">
      <Head />
      <Preview>Confirme seu e-mail para receber novidades.</Preview>
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
          <Heading style={{ fontSize: "28px" }}>Só falta confirmar.</Heading>
          <Text style={{ color: "#525252", lineHeight: "26px" }}>
            Você está a um clique de receber as novidades de{" "}
            {"{{.Project.Name}}"}. Confirme que este e-mail é seu para entrar na
            lista.
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
            Confirmar minha inscrição →
          </Button>
          <Hr style={{ borderColor: "#eee", margin: "32px 0" }} />
          <Text style={{ fontSize: "12px", color: "#737373" }}>
            O link expira em 24 horas. Se você não solicitou esta inscrição,
            ignore este e-mail.
          </Text>
        </Container>
      </Body>
    </Html>
  );
}
