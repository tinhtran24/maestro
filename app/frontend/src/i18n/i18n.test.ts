import { describe, expect, it } from "vitest";
import { createTranslator } from "./index";

describe("i18n translator", () => {
  it("resolves app copy and interpolates variables", () => {
    const t = createTranslator("en");

    expect(t("app.brand")).toBe("Thanos");
    expect(t("workspace.dataKey", { key: "demo" })).toBe("Data key: demo");
  });
});
