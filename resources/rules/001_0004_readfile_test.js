const metadata = {
    scope: "package",
    title: "Test mxlint.readfile function",
    description: "Validates that mxlint.readfile can read files relative to the input file",
    authors: ["Test <test@example.com>"],
    custom: {
        category: "Testing",
        rulename: "ReadFileTest",
        severity: "LOW",
        rulenumber: "001_0004",
        remediation: "N/A",
        input: ".*Security\\$ProjectSecurity\\.yaml"
    }
};


function rule(input = {}) {
    const errors = [];

    // Use mxlint.readfile to read the Settings$ProjectSettings.yaml file
    // which should be in the same directory as the input file
    try {
        const settingsContent = mxlint.readfile("Settings$ProjectSettings.yaml");

        // Check if the settings file contains expected content
        if (!settingsContent.includes("$Type:")) {
            errors.push("Settings file does not contain expected $Type field");
        }

        // Verify we can parse and check content
        if (!settingsContent.includes("Settings$ProjectSettings")) {
            errors.push("Settings file does not contain Settings$ProjectSettings type");
        }
    } catch (e) {
        errors.push("Failed to read Settings$ProjectSettings.yaml: " + e.message);
    }

    // Determine final authorization decision
    const allow = errors.length === 0;

    return {
        allow,
        errors
    };
}

