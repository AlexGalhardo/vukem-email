export type DocVersion = { id: string; label: string; prefix: string; contentDir: string };
export const versions: readonly DocVersion[] = [{ id: "0.1", label: "v0.1", prefix: "", contentDir: "docs" }];
export const currentVersion: DocVersion = versions[0];
export function docsBase(v: DocVersion): string { return v.prefix ? `/docs/${v.prefix}` : "/docs"; }
export function versionOfPath(_pathname: string): DocVersion { return currentVersion; }
export function versionHref(href: string, _v: DocVersion): string { return href; }
