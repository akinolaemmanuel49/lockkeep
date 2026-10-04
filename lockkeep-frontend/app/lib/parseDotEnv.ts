export interface DotEnvEntry {
    key: string;
    value: string;
}

const ESCAPES: Record<string, string> = {
    n: "\n",
    t: "\t",
    r: "\r",
    '"': '"',
    "'": "'",
    "\\": "\\",
};

/**
 * Parses `.env`-style content into `KEY=VALUE` entries. UI-only helper used to
 * bulk-fill workspace secrets; NOT a shell evaluator.
 *
 * Supported: `KEY=value`, `KEY = value`, `export KEY=value`, blank lines,
 * `#` comments, single/double-quoted values (double quotes honor `\n`, `\t`,
 * `\"`, `\\`), and inline ` # comment` on unquoted values. Multi-line values
 * are intentionally not supported. Duplicate keys are preserved (last wins in
 * the caller).
 */
export function parseDotEnv(text: string): DotEnvEntry[] {
    const entries: DotEnvEntry[] = [];

    for (const rawLine of text.split(/\r?\n/)) {
        let line = rawLine.trim();
        if (!line || line.startsWith("#")) continue;
        if (line.startsWith("export ")) {
            line = line.slice("export ".length).trim();
        }

        const eq = line.indexOf("=");
        if (eq === -1) continue;

        const key = line.slice(0, eq).trim();
        if (!key) continue;

        entries.push({ key, value: unquoteValue(line.slice(eq + 1).trim()) });
    }

    return entries;
}

function unquoteValue(value: string): string {
    if (value.length >= 2) {
        if (value.startsWith('"') && value.endsWith('"')) {
            return value
                .slice(1, -1)
                .replace(/\\(.)/g, (_m, char: string) => ESCAPES[char] ?? char);
        }
        if (value.startsWith("'") && value.endsWith("'")) {
            return value.slice(1, -1);
        }
    }

    // Unquoted: strip a trailing inline comment (space-hash).
    const commentAt = value.indexOf(" #");
    if (commentAt !== -1) value = value.slice(0, commentAt).trimEnd();
    return value;
}