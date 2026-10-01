import { describe, expect, it } from "vitest";
import { parseDotEnv } from "./parseDotEnv";

describe("parseDotEnv", () => {
    it("parses simple KEY=value lines", () => {
        expect(parseDotEnv("A=1\nB=hello world")).toEqual([
            { key: "A", value: "1" },
            { key: "B", value: "hello world" },
        ]);
    });

    it("trims whitespace around key and value", () => {
        expect(parseDotEnv("  A =  spaced value  ")).toEqual([
            { key: "A", value: "spaced value" },
        ]);
    });

    it("handles export prefix", () => {
        expect(parseDotEnv("export DATABASE_URL=postgres://db")).toEqual([
            { key: "DATABASE_URL", value: "postgres://db" },
        ]);
    });

    it("skips blank lines and # comments", () => {
        const out = parseDotEnv("# header\n\n\nA=1\n# trailing");
        expect(out).toEqual([{ key: "A", value: "1" }]);
    });

    it("ignores lines without an equals sign", () => {
        expect(parseDotEnv("JUST_A_FLAG\nA=1")).toEqual([
            { key: "A", value: "1" },
        ]);
    });

    it("unquotes single and double quoted values", () => {
        expect(parseDotEnv('A="double quoted"\nB=\'single quoted\'')).toEqual([
            { key: "A", value: "double quoted" },
            { key: "B", value: "single quoted" },
        ]);
    });

    it("honors escapes inside double quotes", () => {
        const out = parseDotEnv('A="line1\\nline2\\t\\"quoted\\"\\\\"');
        expect(out[0].value).toBe('line1\nline2\t"quoted"\\');
    });

    it("strips inline comments only on unquoted values", () => {
        expect(parseDotEnv("A=value # comment\nB=a#b#c\nC='keep # this'")).toEqual([
            { key: "A", value: "value" },
            { key: "B", value: "a#b#c" },
            { key: "C", value: "keep # this" },
        ]);
    });

    it("handles CRLF line endings", () => {
        expect(parseDotEnv("A=1\r\nB=2\r\n")).toEqual([
            { key: "A", value: "1" },
            { key: "B", value: "2" },
        ]);
    });

    it("preserves duplicate keys (caller dedupes keeping last)", () => {
        expect(parseDotEnv("A=1\nA=2")).toEqual([
            { key: "A", value: "1" },
            { key: "A", value: "2" },
        ]);
    });
});