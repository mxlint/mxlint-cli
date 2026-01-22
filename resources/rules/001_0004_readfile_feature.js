const metadata = {
    scope: "package",
    title: "Project settings must have valid configuration",
    description: "Validates project settings by reading related configuration files",
    authors: ["Test <test@example.com>"],
    custom: {
        category: "Configuration",
        rulename: "ProjectSettingsValidation",
        severity: "LOW",
        rulenumber: "001_0004",
        remediation: "Ensure Settings$ProjectSettings.yaml exists and contains valid configuration",
        input: ".*Security\\$ProjectSecurity\\.yaml"
    }
};


function rule(input = {}) {
    const errors = [];

    // Use mxlint.readfile to read the Settings$ProjectSettings.yaml file
    // which should be in the same directory as the Security$ProjectSecurity.yaml input file
    try {
        const settingsContent = mxlint.io.readfile("Settings$ProjectSettings.yaml");

        // Check if the settings file contains expected content
        if (!settingsContent.includes("$Type:")) {
            errors.push("Settings file does not contain expected $Type field");
        }

        // Verify we can read the content correctly
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

