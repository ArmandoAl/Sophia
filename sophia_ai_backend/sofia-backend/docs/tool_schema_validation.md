# Tool Schema Validation

Sprint 13 adds real validation for `ToolDefinition.input_schema` before any AI action proposal is accepted.

## Where It Runs

Validation runs in two places:

- AI Runtime safety gate, before a planned action becomes visible as a proposal candidate.
- AI Actions service, before `POST /ai/action-proposals` persists a proposal.

This means both future model-generated actions and manually created proposals use the same schema gate.

## Supported JSON Schema Subset

The local validator supports the subset currently needed by Sofía tools:

- `type`
- `required`
- `properties`
- `items`
- `enum`
- `additionalProperties: false`

Supported JSON types:

- `object`
- `array`
- `string`
- `boolean`
- `number`
- `integer`

The implementation lives in `internal/platform/jsonschema`.

## Default Tool Schemas

Default tool definitions now declare explicit allowed fields and use `additionalProperties:false`.

This intentionally blocks fields such as:

- `user_id`
- `role`
- `owner`
- arbitrary nested payloads

Ownership must continue to come from JWT/context, never from model output.

## Limitations

This is not a complete JSON Schema draft implementation. Missing features include:

- `format`
- `minLength`
- `maxLength`
- numeric ranges
- `oneOf` / `anyOf` / `allOf`
- nested advanced constraints

Those can be added later or replaced by a vetted schema library if the backend needs full draft compatibility.
