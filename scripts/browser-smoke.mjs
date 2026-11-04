import { chromium, expect } from "@playwright/test";
const browser = await chromium.launch({
  executablePath: process.env.BROWSER_EXECUTABLE || "/usr/bin/chromium",
  headless: true,
  args: ["--no-sandbox"],
});
const context = await browser.newContext({
  viewport: { width: 1440, height: 1000 },
});
await context.addCookies(JSON.parse(process.env.BROWSER_COOKIES));
const page = await context.newPage();
const errors = [];
page.on("pageerror", (e) => errors.push(e.message));
try {
  const root = process.env.BROWSER_URL;
  await page.goto(root);
  await expect(
    page.getByRole("heading", { name: "Cada conexão importa." }),
  ).toBeVisible();
  await expect(page.getByText("ana@example.com")).toBeVisible();
  await page.screenshot({
    path: "/tmp/vukem-email-dashboard.png",
    fullPage: true,
  });
  await page.getByRole("link", { name: "Inscritos", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Sua audiência." }),
  ).toBeVisible();
  await page.getByRole("textbox", { name: "Buscar e-mail" }).fill("ana@");
  await page.getByRole("textbox", { name: "Buscar e-mail" }).press("Enter");
  await expect(page.getByText("ana@example.com")).toBeVisible();
  await expect(page.getByText("bruno@example.com")).toHaveCount(0);
  await page.getByRole("link", { name: "Nova campanha" }).click();
  await expect(
    page.getByRole("heading", { name: "Uma nova conversa." }),
  ).toBeVisible();
  await page.getByLabel("Assunto do e-mail").fill("Uma novidade para você");
  await page.getByRole("button", { name: "Atualizar prévia" }).click();
  await expect(page.getByTitle("Prévia da campanha")).toBeVisible({
    timeout: 20000,
  });
  await expect(
    page.frameLocator("iframe").getByText("Temos novidades."),
  ).toBeVisible();
  await page.getByRole("button", { name: "Salvar rascunho" }).click();
  await expect(page.getByText("Rascunho salvo.")).toBeVisible();
  await page.getByRole("button", { name: "React Email", exact: true }).click();
  await page.getByRole("button", { name: "Atualizar prévia" }).click();
  await expect(
    page.frameLocator("iframe").getByText("Que bom ter você por aqui."),
  ).toBeVisible({ timeout: 20000 });
  await page.screenshot({
    path: "/tmp/vukem-email-editor.png",
    fullPage: true,
  });
  await page.getByRole("button", { name: "Revisar envio" }).click();
  await expect(
    page.getByRole("heading", {
      name: "Pronto para chegar à caixa de entrada?",
    }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Confirmar e enviar campanha" })
    .click();
  await expect(
    page.getByRole("heading", { name: "Conversas que conectam." }),
  ).toBeVisible();
  await expect(
    page.getByText("Uma novidade para você", { exact: true }),
  ).toBeVisible();
  await page
    .getByRole("link", { name: "Projetos", exact: false })
    .first()
    .click();
  await expect(
    page.getByRole("heading", { name: "Um lugar para cada projeto." }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Editar projeto" }).click();
  await expect(
    page.getByRole("heading", { name: "Configurar projeto" }),
  ).toBeVisible();
  await page.keyboard.press("Escape");
  // A separate guest context validates the public flow and the single Google button.
  const guest = await browser.newContext({
    viewport: { width: 390, height: 844 },
  });
  const publicPage = await guest.newPage();
  publicPage.on("pageerror", (e) => errors.push(e.message));
  await publicPage.goto(root + "/");
  await expect(
    publicPage.getByRole("link", { name: "Continuar com Google" }),
  ).toBeVisible();
  await publicPage.goto(root + "/subscribe/hexascore");
  await expect(
    publicPage.getByRole("heading", { name: /O próximo capítulo/ }),
  ).toBeVisible();
  await publicPage.getByLabel("Seu e-mail").fill("browser@example.com");
  await publicPage.getByRole("button", { name: "Quero receber" }).click();
  await expect(publicPage.getByRole("status")).toContainText(
    "link de confirmação",
  );
  await publicPage.screenshot({
    path: "/tmp/vukem-email-subscribe-mobile.png",
    fullPage: true,
  });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(root + "/");
  await expect(
    page.getByRole("heading", { name: "Cada conexão importa." }),
  ).toBeVisible();
  if (
    await page.evaluate(
      () => document.documentElement.scrollWidth > innerWidth + 1,
    )
  )
    throw new Error("Dashboard overflows mobile viewport");
  if (errors.length) throw new Error(errors.join("\n"));
  console.log(
    "Browser smoke passed: dashboard, subscriber search, editor, preview, campaign send, projects, public subscription and mobile.",
  );
  await guest.close();
} finally {
  await browser.close();
}
