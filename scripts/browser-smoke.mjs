import { chromium, expect } from "@playwright/test";
import { tmpdir } from "node:os";
import { join } from "node:path";
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
    page.getByRole("heading", { name: "Every connection matters." }),
  ).toBeVisible();
  await expect(page.getByText("ana@example.com")).toBeVisible();
  await page.screenshot({
    path: join(tmpdir(), "vukem-email-dashboard.png"),
    fullPage: true,
  });
  await page.getByRole("link", { name: "Subscribers", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Your audience." }),
  ).toBeVisible();
  await page.getByRole("textbox", { name: "Search email" }).fill("ana@");
  await page.getByRole("textbox", { name: "Search email" }).press("Enter");
  await expect(page.getByText("ana@example.com")).toBeVisible();
  await expect(page.getByText("bruno@example.com")).toHaveCount(0);
  await page.getByRole("link", { name: "New campaign" }).click();
  await expect(
    page.getByRole("heading", { name: "A new conversation." }),
  ).toBeVisible();
  await page.getByLabel("Email subject").fill("An update for you");
  await page.getByRole("button", { name: "Refresh preview" }).click();
  await expect(page.getByTitle("Campaign preview")).toBeVisible({
    timeout: 20000,
  });
  await expect(
    page.frameLocator("iframe").getByText("We have news."),
  ).toBeVisible();
  await page.getByRole("button", { name: "Save draft" }).click();
  await expect(page.getByText("Draft saved.")).toBeVisible();
  await page.getByRole("button", { name: "React Email", exact: true }).click();
  await page.getByRole("button", { name: "Refresh preview" }).click();
  await expect(
    page.frameLocator("iframe").getByText("Welcome aboard."),
  ).toBeVisible({ timeout: 20000 });
  await page.screenshot({
    path: join(tmpdir(), "vukem-email-editor.png"),
    fullPage: true,
  });
  await page.getByRole("button", { name: "Review send" }).click();
  await expect(
    page.getByRole("heading", {
      name: "Ready for the inbox?",
    }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Confirm and send campaign" })
    .click();
  await expect(
    page.getByRole("heading", { name: "Conversations that connect." }),
  ).toBeVisible();
  await expect(
    page.getByText("An update for you", { exact: true }),
  ).toBeVisible();
  await page
    .getByRole("link", { name: "Projects", exact: false })
    .first()
    .click();
  await expect(
    page.getByRole("heading", { name: "A place for every project." }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Edit project" }).click();
  await expect(
    page.getByRole("heading", { name: "Configure project" }),
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
    publicPage.getByRole("link", { name: "Continue with Google" }),
  ).toBeVisible();
  await publicPage.goto(root + "/subscribe/hexascore");
  await expect(
    publicPage.getByRole("heading", { name: /The next chapter/ }),
  ).toBeVisible();
  await publicPage.getByLabel("Your email").fill("browser@example.com");
  await publicPage.getByRole("button", { name: "Subscribe" }).click();
  await expect(publicPage.getByRole("status")).toContainText(
    "confirmation link",
  );
  await publicPage.screenshot({
    path: join(tmpdir(), "vukem-email-subscribe-mobile.png"),
    fullPage: true,
  });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(root + "/");
  await expect(
    page.getByRole("heading", { name: "Every connection matters." }),
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
