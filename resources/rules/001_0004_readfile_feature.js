const metadata = {
    scope: "package",
    title: "Files in directory must not be empty",
    description: "Validates that all files in the current directory contain content and are not empty",
    authors: ["Test <test@example.com>"],
    custom: {
        category: "Integrity",
        rulename: "NoEmptyFiles",
        severity: "LOW",
        rulenumber: "001_0004",
        remediation: "Ensure all files in the directory have content. Remove or populate empty files.",
        input: ".*Microflows\\$Microflow\\.yaml"
    }
};


function rule(input = {}) {
    const errors = [];

    // Use mxlint.io.listdir and mxlint.io.readfile to read file in another directory
    try {

        const items = mxlint.io.listdir(".");
        if (items.length === 0) {
            errors.push("No items found in the current directory");
        }
        for (const item of items) {
            const itemPath = item;
            if (!mxlint.io.isdir(itemPath)) {
                const content = mxlint.io.readfile(itemPath);
                if (content === "") {
                    errors.push("item " + itemPath + " is empty");
                }
            }
        }
    } catch (e) {
        errors.push("Failed to read items in the current directory: " + e.message);
    }

    // Determine final authorization decision
    const allow = errors.length === 0;

    return {
        allow,
        errors
    };
}

