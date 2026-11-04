import React from "react";
import { mkdir, writeFile } from "node:fs/promises";
import { render } from "@react-email/render";
import Confirmation from "../resources/js/emails/confirmation";
import Campaign from "../resources/js/emails/campaign";
await mkdir("resources/emails", { recursive: true });
for (const [name, component] of [
  ["confirmation", <Confirmation />],
  ["campaign", <Campaign />],
] as const) {
  await writeFile(`resources/emails/${name}.html`, await render(component));
}
