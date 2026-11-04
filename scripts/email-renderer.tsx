import React from "react";
import { isolatedMdx } from "./isolated-mdx";
import * as Email from "@react-email/components";
import { render } from "@react-email/render";
import sanitizeHtml from "sanitize-html";

export const components = {
  Button: Email.Button,
  Column: Email.Column,
  Container: Email.Container,
  Heading: Email.Heading,
  Hr: Email.Hr,
  Img: Email.Img,
  Link: Email.Link,
  Row: Email.Row,
  Section: Email.Section,
  Text: Email.Text,
  CodeInline: Email.CodeInline,
  Markdown: Email.Markdown,
  Tailwind: Email.Tailwind,
  CodeBlock: (props: any) =>
    React.createElement(Email.CodeBlock, {
      theme: Email.dracula,
      language: "javascript",
      ...props,
    }),
};
const htmlTags = new Set([
  "a",
  "p",
  "div",
  "span",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "ul",
  "ol",
  "li",
  "strong",
  "em",
  "del",
  "blockquote",
  "pre",
  "code",
  "br",
  "hr",
  "img",
  "table",
  "thead",
  "tbody",
  "tr",
  "th",
  "td",
  "sup",
  "sub",
]);
function hydrate(tree: any): React.ReactNode {
  let nodes = 0;
  function element(node: any, depth = 0): React.ReactNode {
    if (++nodes > 20000 || depth > 100) throw new Error("Content too complex");
    if (node === null || node === undefined || typeof node === "boolean")
      return null;
    if (typeof node === "string" || typeof node === "number") return node;
    if (Array.isArray(node))
      return node.map((child: any) => element(child, depth + 1));
    if (
      typeof node !== "object" ||
      typeof node.type !== "string" ||
      !node.props ||
      typeof node.props !== "object" ||
      Array.isArray(node.props)
    )
      throw new Error("Invalid element");
    const name = node.type;
    if (!Object.hasOwn(components, name) && !htmlTags.has(name))
      throw new Error("Unknown email component");
    const props: Record<string, unknown> = Object.create(null);
    props.key = nodes;
    for (const [key, value] of Object.entries(node.props)) {
      if (
        /^on/i.test(key) ||
        [
          "dangerouslySetInnerHTML",
          "ref",
          "srcDoc",
          "__proto__",
          "constructor",
          "prototype",
        ].includes(key)
      )
        throw new Error("Unsafe attribute");
      if (key !== "children") props[key] = value;
    }
    return React.createElement(
      Object.hasOwn(components, name)
        ? components[name as keyof typeof components]
        : name,
      props,
      element(node.props.children, depth + 1),
    );
  }
  return element(tree);
}
export async function renderContent(
  mode: string,
  source: string,
): Promise<string> {
  if (
    !["markdown", "mdx", "components"].includes(mode) ||
    !source.trim() ||
    Buffer.byteLength(source) > 100000
  )
    throw new Error("Invalid content");
  const tree = await isolatedMdx(mode, source, Object.keys(components));
  const html = await render(
    React.createElement(React.Fragment, null, hydrate(tree)),
  );
  return sanitizeHtml(html, {
    allowedTags: [...htmlTags],
    allowedAttributes: {
      "*": ["style", "class", "align", "valign", "width", "height", "role"],
      a: ["href", "target", "rel", "style"],
      img: ["src", "alt", "width", "height", "style"],
      table: ["border", "cellpadding", "cellspacing", "role", "width", "style"],
      td: ["colspan", "rowspan", "style", "align", "valign"],
      th: ["colspan", "rowspan", "style", "align"],
    },
    allowedSchemes: ["https", "http", "mailto"],
    allowProtocolRelative: false,
    allowedStyles: {
      "*": {
        color: [/^(?:#[0-9a-f]{3,8}|[a-z]+|rgba?\([\d,. %]+\))$/i],
        "background-color": [/^(?:#[0-9a-f]{3,8}|[a-z]+|rgba?\([\d,. %]+\))$/i],
        "font-size": [/^[\d.]+(?:px|em|rem|%)$/],
        "font-weight": [/^(?:[1-9]00|normal|bold)$/],
        "font-family": [/^[a-zA-Z\s,'"-]+$/],
        "line-height": [/^[\d.]+(?:px|em|rem|%)?$/],
        "text-align": [/^(?:left|center|right|justify)$/],
        "text-decoration": [/^(?:none|underline|line-through)$/],
        "border-radius": [/^[\d.]+(?:px|%)$/],
        border: [/^[\d.]+px (?:solid|dashed) #[0-9a-f]{3,8}$/i],
        padding: [/^[\d. px%]+$/],
        margin: [/^[\d. px%auto-]+$/],
        "padding-top": [/^[\d.]+px$/],
        "padding-bottom": [/^[\d.]+px$/],
        "padding-left": [/^[\d.]+px$/],
        "padding-right": [/^[\d.]+px$/],
        "margin-top": [/^[\d.]+px$/],
        "margin-bottom": [/^[\d.]+px$/],
        width: [/^[\d.]+(?:px|%)$/],
        "max-width": [/^[\d.]+(?:px|%)$/],
        display: [/^(?:block|inline|inline-block|table|table-cell|none)$/],
      },
    },
  });
}
