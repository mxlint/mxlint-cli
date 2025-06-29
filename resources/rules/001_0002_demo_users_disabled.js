const metadata = {
    scope: "package",
    title: "Business apps should disable demo users",
    description: "No demo users ",
    authors: ["Xiwen Cheng <x@cinaq.com>"],
    custom: {
        category: "Security",
        rulename: "DemoUsersDisabled",
        severity: "HIGH",
        rulenumber: "001_0002",
        remediation: "Disable demo users in Project Security",
        input: ".*Security\\$ProjectSecurity\\.yaml$"
    }
};


function rule(input = {}) {

    const errors = [];
    
    if (input?.EnableDemoUsers === true) {
        const errorMessage = `[${metadata.custom.severity}, ${metadata.custom.category}, ${metadata.custom.rulenumber}] ${metadata.title}`;
        errors.push(errorMessage);
    }

    // Determine final authorization decision
    const allow = errors.length === 0;
    
    return {
        allow,
        errors
    };
}
