const metadata = {
    scope: "package",
    title: "TypeScript example rule",
    description: "Validates that the Name field is set",
    authors: ["Test <test@example.com>"],
    custom: {
        category: "Example",
        rulename: "TypescriptExampleRule",
        severity: "LOW",
        rulenumber: "001_0005",
        remediation: "No action required",
        input: ".*\\$Microflow\\.yaml"
    }
};
type RuleInput = {
    Name?: string;
};

function rule(input: RuleInput = {}) {
    const errors: string[] = [];
    const name = input.Name ?? "";

    if (name === "") {
        errors.push("Name must be set");
    }

    return {
        allow: errors.length === 0,
        errors
    };
}
