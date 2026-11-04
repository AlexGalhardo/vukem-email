import assert from "node:assert/strict";
import test from "node:test";
import { renderContent } from "./email-renderer";
test("Markdown renders GFM tables, links, code, lists and strikethrough", async () => {
  const result = await renderContent(
    "markdown",
    '# Hello\n\n**Bold** and ~~old~~.\n\n- One\n- Two\n\n| A | B |\n|---|---|\n| 1 | 2 |\n\n```go\nfmt.Println("hello")\n```\n\n[Link](https://example.com)',
  );
  for (const value of [
    "<h1>",
    "<strong>",
    "<del>",
    "<table>",
    "<li>",
    "<pre>",
    'href="https://example.com"',
  ])
    assert.ok(result.includes(value), value);
});
test("MDX supports variables, expressions, functions, imports, aliases and native React Email components", async () => {
  const result = await renderContent(
    "mdx",
    `import { Button as CTA, Section, Row, Column, Text } from '@react-email/components'
import React from 'react'

export const name = 'Alex'
export const items = ['One', 'Two']
export const Greeting = ({name}) => <Text>Hello {name}</Text>

# Hello {name.toUpperCase()}

<Greeting name={name}/>

<Section style={{padding: 12 * 2 + 'px'}}><Row><Column>{items.map(item => <Text key={item}>{item}</Text>)}</Column></Row><CTA href={'https://'+'example.com'}>Read more</CTA></Section>
`,
  );
  assert.match(result, /Hello ALEX/);
  assert.match(result, /Hello Alex/);
  assert.match(result, /One/);
  assert.match(result, /Two/);
  assert.match(result, /padding:24px/);
  assert.match(result, /href="https:\/\/example.com"/);
});
test("MDX namespace imports and components mode render correctly", async () => {
  const result = await renderContent(
    "components",
    `import * as Email from '@react-email/components'\n\n<Email.Section><Email.Text>Namespace works</Email.Text></Email.Section>`,
  );
  assert.match(result, /Namespace works/);
  assert.match(result, /<table/);
});
test("isolated MDX has no filesystem, network, environment, Node globals or host constructors", async () => {
  const bad = [
    'import fs from "node:fs"\n\n# Hi',
    "{process.env.SECRET}",
    '{require("node:fs")}',
    '{fetch("https://example.com")}',
    "{globalThis.process.env.SECRET}",
    '{({}).constructor.constructor("return process")().env.SECRET}',
    '<Text onClick="bad">hi</Text>',
    '<Text dangerouslySetInnerHTML={{__html:"bad"}} />',
    "<script>alert(1)</script>",
    "<constructor>bad</constructor>",
  ];
  for (const source of bad) await assert.rejects(renderContent("mdx", source));
  const result = await renderContent(
    "mdx",
    "{typeof process} / {typeof fetch} / {typeof require}",
  );
  assert.match(result, /undefined \/ undefined \/ undefined/);
});
test("isolated execution enforces CPU and memory limits", async () => {
  await assert.rejects(renderContent("mdx", "{(() => { while (true) {} })()}"));
  await assert.rejects(
    renderContent("mdx", '{new Array(100000000).fill("memory").join("")}'),
  );
});
test("sanitizes unsafe URL schemes and CSS without dropping safe email layouts", async () => {
  const result = await renderContent(
    "components",
    '<Section style={{ backgroundColor: "#fff" }}><Link href="javascript:alert(1)">Bad</Link><Img src="https://example.com/image.png" alt="Safe" width="100" /></Section>',
  );
  assert.doesNotMatch(result, /javascript:/);
  assert.match(result, /src="https:\/\/example.com\/image.png"/);
  assert.match(result, /background-color:#fff/);
});
test("rejects unknown formats, empty and oversized content", async () => {
  for (const [mode, source] of [
    ["html", "x"],
    ["mdx", ""],
    ["markdown", "x".repeat(100001)],
  ])
    await assert.rejects(renderContent(mode, source));
});

test("CodeBlock and Tailwind native components render in the isolated pipeline", async () => {
  const result = await renderContent(
    "components",
    '<Tailwind><Section className="bg-gray-100 p-6"><Text>Tailwind works</Text><CodeBlock code={"const hello = 42;"} language="javascript" /></Section></Tailwind>',
  );
  assert.match(result, /Tailwind works/);
  assert.match(result, /hello/);
  assert.match(result, /42/);
});
