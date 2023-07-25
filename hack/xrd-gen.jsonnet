local k8s = import 'functions.jsonnet';

local s = {
    config: std.parseJson(std.extVar('config')),
    crd: std.parseJson(std.extVar('crd')),
    data: std.parseJson(std.extVar('data')),
    readinessChecks: std.extVar('readinessChecks'),
};

local plural = k8s.NameToPlural(s.config);
local fqdn = k8s.FQDN(plural, s.config.group);
local version = k8s.GetVersion(s.crd);

local uidFieldName = 'uid';

local definitionSpec = k8s.GenerateSchema(
    version.schema.openAPIV3Schema.properties.spec,
    s.config,
    ['spec'],
);

local definitionStatus = k8s.GenerateSchema(
    version.schema.openAPIV3Schema.properties.status,
    s.config,
    ['status'],
);

{
    definition: {
        apiVersion: 'apiextensions.crossplane.io/v1',
        kind: 'CompositeResourceDefinition',
        metadata: {
            name: "x"+fqdn,
        },
        spec: {
            claimNames: {
                kind: s.config.kind,
                plural: plural,
            },
            [if std.objectHas(s.config, "connectionSecretKeys") then "connectionSecretKeys"]:
                s.config.connectionSecretKeys,
            group: s.config.group,
            names: {
                kind: s.crd.spec.names.kind,
                plural: s.crd.spec.names.plural,
                categories: k8s.GenerateCategories(s.config.group),
            },
            versions: [
                {
                    name: s.config.version,
                    referenceable: version.storage,
                    served: version.served,
                    schema: {
                        openAPIV3Schema: {
                            properties: {
                                spec: definitionSpec,
                                status: definitionStatus
                                {
                                    properties+: {
                                        [uidFieldName]: {
                                            description: 'The unique ID of this %s resource reported by the provider' % [s.config.kind],
                                            type: 'string',
                                        },
                                        observed: {
                                            description: 'Freeform field containing information about the observed status.',
                                            type: 'object',
                                            "x-kubernetes-preserve-unknown-fields": true,
                                        },
                                    },
                                },
                            },
                        },
                    },
                    [if std.objectHas(version, "additionalPrinterColumns") then "additionalPrinterColumns"]: k8s.FilterPrinterColumns(version.additionalPrinterColumns),
                },
            ],
        },
    },
}