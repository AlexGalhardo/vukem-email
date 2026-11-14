import { renderContent } from "./email-renderer";
let input = "";
for await (const chunk of process.stdin) {
  input += chunk;
  if (input.length > 200000) throw new Error("Input too large");
}
try {
  const { mode, source } = JSON.parse(input);
  process.stdout.write(await renderContent(mode, source));
} catch {
  process.stderr.write("Invalid email content\n");
  // 65 tells the Go caller the content was rejected; any other exit is an infrastructure failure.
  process.exitCode = 65;
}
