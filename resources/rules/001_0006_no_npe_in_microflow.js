const metadata = {
    scope: "package",
    title: "No NPE in microflow",
    description: "Disallow Non-persistent entity usage in persistent microflows",
    authors: ["Xiwen Cheng <x@cinaq.com>"],
    custom: {
        category: "Security",
        rulename: "NonPersistentEntityUsageInPersistentMicroflows",
        severity: "HIGH",
        rulenumber: "001_0006",
        remediation: "Remove the non-persistent entity from the persistent microflow",
        input: ".*\\$Microflow\\.yaml"
    }
};

function entityName(entityRef) {
    const parts = entityRef.split(".");
    return parts[parts.length - 1];
}

function isNPE(entity) {
    // read the parent domain model and check if the entity is NPE
    // entity: Module2.EntityNonPersist
    if (entity === undefined) {
        return false;
    }
    const moduleName = entity.split(".")[0];
    const domainModelPath = [moduleName, "DomainModels$DomainModel.yaml"].join("/");
    const domainModel = mxlint.io.readYaml(domainModelPath);
    if (domainModel === undefined || domainModel.Entities === undefined) {
        return false;
    }
    const name = entityName(entity);
    const found = domainModel.Entities.find(e => e.Name === name);
    if (found === undefined || found.MaybeGeneralization === undefined) {
        return false;
    }
    return found.MaybeGeneralization.Persistable === false;
}

function rule(input = {}) {
    const errors = [];

    // for each entity, check the entity type via the parent domain model if it's NPE or not
    for (const object of input.ObjectCollection.Objects) {
        if (object.$Type === "Microflows$ActionActivity") {
            const entity = object.Action.Entity;
            try {
                if (isNPE(entity)) {
                    errors.push("Non-persistent entity used in microflow: " + entity);
                }
            } catch (e) {
                errors.push("Failed to check if entity is NPE: " + e);
            }
        }
    }
    // Determine final authorization decision
    const allow = errors.length === 0;

    return {
        allow,
        errors
    };
}

