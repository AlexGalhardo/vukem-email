import { mkdir, writeFile } from "node:fs/promises";
import { render } from "@react-email/render";
import React from "react";
import Campaign from "../resources/js/emails/campaign";
import Confirmation from "../resources/js/emails/confirmation";

await mkdir("resources/emails", { recursive: true });
for (const [name, Component] of Object.entries({
  confirmation: Confirmation,
  campaign: Campaign,
})) {
  await writeFile(`resources/emails/${name}.html`, await render(<Component />));
}
