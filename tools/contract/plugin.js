// Redocly plugin with the contract rules that configurable rules cannot express.
// Wired from redocly.yaml as `contract/<rule>`.
import { createHash } from 'node:crypto';

const FROZEN_KEY = 'x-frozen';
const PUBLIC_KEY = 'x-public';
const HASH_PREFIX = 'sha256:';
const HTTP_METHODS = ['get', 'put', 'post', 'delete', 'options', 'head', 'patch', 'trace'];

// Canonical JSON: object keys sorted, so key order in YAML does not change the hash.
function canonical(value) {
  if (Array.isArray(value)) {
    return `[${value.map(canonical).join(',')}]`;
  }
  if (value !== null && typeof value === 'object') {
    const keys = Object.keys(value).filter((key) => key !== FROZEN_KEY).sort();
    return `{${keys.map((key) => `${JSON.stringify(key)}:${canonical(value[key])}`).join(',')}}`;
  }
  return JSON.stringify(value);
}

export function frozenHash(node) {
  return HASH_PREFIX + createHash('sha256').update(canonical(node)).digest('hex');
}

function* frozenCandidates(root) {
  for (const [path, item] of Object.entries(root.paths ?? {})) {
    for (const method of HTTP_METHODS) {
      if (item?.[method]) {
        yield { node: item[method], pointer: ['paths', path, method] };
      }
    }
  }
  for (const [name, schema] of Object.entries(root.components?.schemas ?? {})) {
    yield { node: schema, pointer: ['components', 'schemas', name] };
  }
}

// frozen-unchanged: a node marked `x-frozen: sha256:<hex>` must still hash to that value.
// Changing a frozen node is a breaking-change decision (ADR + version bump, docs/api/CONTRACT.md);
// the new hash is printed so the author can update the marker deliberately.
function FrozenUnchanged() {
  return {
    Root: {
      leave(root, ctx) {
        for (const { node, pointer } of frozenCandidates(root)) {
          if (!node || !(FROZEN_KEY in node)) {
            continue;
          }
          const declared = node[FROZEN_KEY];
          const actual = frozenHash(node);
          if (declared !== actual) {
            ctx.report({
              message: `Frozen contract node changed (declared ${declared}, actual ${actual}). See docs/api/CONTRACT.md before updating the marker.`,
              location: ctx.location.child([...pointer, FROZEN_KEY]),
            });
          }
        }
      },
    },
  };
}

// public-operation-explicit: `security: []` (no auth) only with `x-public: true`, and vice versa,
// so opting an operation out of authentication is always a visible, deliberate decision.
function PublicOperationExplicit() {
  return {
    Operation(operation, ctx) {
      const optsOut = Array.isArray(operation.security) && operation.security.length === 0;
      const markedPublic = operation[PUBLIC_KEY] === true;
      if (optsOut && !markedPublic) {
        ctx.report({
          message: 'Operation disables security (`security: []`) but is not marked `x-public: true`.',
          location: ctx.location.child(['security']),
        });
      }
      if (markedPublic && !optsOut) {
        ctx.report({
          message: 'Operation is marked `x-public: true` but does not declare `security: []`.',
          location: ctx.location.child([PUBLIC_KEY]),
        });
      }
    },
  };
}

function hasExample(node) {
  return node !== null && typeof node === 'object' && ('example' in node || 'examples' in node);
}

// body-example: every request/response body carries an example, either on the media type
// or on the schema it references (one example per schema keeps them DRY and validated by
// no-invalid-schema-examples).
function BodyExample() {
  return {
    MediaType(mediaType, ctx) {
      if (hasExample(mediaType)) {
        return;
      }
      const schema = mediaType.schema ? ctx.resolve(mediaType.schema).node : undefined;
      if (hasExample(schema)) {
        return;
      }
      ctx.report({
        message: 'Body has no example: add `example`/`examples` to the media type or to the referenced schema.',
        location: ctx.location,
      });
    },
  };
}

export default function contractPlugin() {
  return {
    id: 'contract',
    rules: {
      oas3: {
        'frozen-unchanged': FrozenUnchanged,
        'public-operation-explicit': PublicOperationExplicit,
        'body-example': BodyExample,
      },
    },
  };
}
