// Redocly plugin with the contract rules that configurable rules cannot express.
// Wired from redocly.yaml as `contract/<rule>`.
import { createHash } from 'node:crypto';

const FROZEN_KEY = 'x-frozen';
const PUBLIC_KEY = 'x-public';
const HASH_PREFIX = 'sha256:';
const ERROR_STATUS = /^([45]\d\d|[45]XX|default)$/;

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

// Every object in the document that carries an x-frozen marker, wherever it sits
// (operations, schemas, responses, parameters, path items...). A marker the rule cannot
// see would be a freeze that guards nothing.
function* frozenNodes(node, pointer = []) {
  if (node === null || typeof node !== 'object') {
    return;
  }
  if (!Array.isArray(node) && FROZEN_KEY in node) {
    yield { node, pointer };
  }
  for (const [key, child] of Object.entries(node)) {
    yield* frozenNodes(child, [...pointer, Array.isArray(node) ? Number(key) : key]);
  }
}

// frozen-unchanged: a node marked `x-frozen: sha256:<hex>` must still hash to that value.
// Changing a frozen node is a breaking-change decision (ADR + version bump, docs/api/CONTRACT.md);
// the new hash is printed so the author can update the marker deliberately.
function FrozenUnchanged() {
  return {
    Root: {
      leave(root, ctx) {
        for (const { node, pointer } of frozenNodes(root)) {
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

// error-response-has-body: every 4xx/5xx/default response, inline or referenced, declares at
// least one media type. `content: {}` is not a body; without this the single-envelope rule
// (rule/error-response-envelope, per media type) would have nothing to check.
function ErrorResponseHasBody() {
  return {
    Responses(responses, ctx) {
      for (const [status, response] of Object.entries(responses)) {
        if (!ERROR_STATUS.test(status)) {
          continue;
        }
        const content = ctx.resolve(response).node?.content;
        const hasMediaType = content !== null && typeof content === 'object' && Object.keys(content).length > 0;
        if (!hasMediaType) {
          ctx.report({
            message: '4xx/5xx responses must carry the shared ErrorResponse envelope (no body declared).',
            location: ctx.location.child([status]),
          });
        }
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
        'error-response-has-body': ErrorResponseHasBody,
      },
    },
  };
}
