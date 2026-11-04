export async function api<T = { message: string }>(
  url: string,
  data: Record<string, unknown>,
): Promise<T> {
  const response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    credentials: "same-origin",
    body: JSON.stringify(data),
  });
  const result = await response
    .json()
    .catch(() => ({ detail: "The request could not be completed." }));
  if (!response.ok)
    throw new Error(
      result.detail ||
        result.message ||
        result.title ||
        "The request could not be completed.",
    );
  return result;
}
