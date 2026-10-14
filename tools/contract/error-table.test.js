// Tests for error-table.js, the generator of the per-module error tables in docs/modules/.
// Run: `make contract-lint-test` (node --test).
import assert from 'node:assert/strict';
import { test } from 'node:test';

import { errorTable, replaceGenerated } from './error-table.js';

const errorBody = (code) => ({ error: { code, message: 'x' } });

// A dereferenced bundle, as `redocly bundle --dereferenced --ext json` writes it.
const spec = {
  paths: {
    '/v1/things': {
      post: {
        operationId: 'createThing',
        'x-module': 'module_things',
        responses: {
          201: { description: 'Created.' },
          400: {
            description: 'Invalid input.',
            content: { 'application/json': { examples: {
              a: { value: errorBody('validation_failed') },
              b: { value: errorBody('validation_failed') },
            } } },
          },
          429: {
            description: 'Rate limit exceeded.',
            content: { 'application/json': { example: errorBody('rate_limited') } },
          },
        },
      },
      get: {
        operationId: 'listThings',
        'x-module': 'module_things',
        responses: {
          200: { description: 'OK.' },
          403: {
            description: 'Wrong | password.',
            content: { 'application/json': { examples: {
              a: { value: errorBody('invalid_password') },
              b: { value: errorBody('unauthorized') },
            } } },
          },
        },
      },
    },
    '/v1/other': {
      get: {
        operationId: 'other',
        'x-module': 'module_other',
        responses: { 404: { description: 'Nope.', content: { 'application/json': { example: errorBody('not_found') } } } },
      },
    },
  },
};

test('lists every non-2xx response of the module, in contract order, with its error codes', () => {
  assert.equal(errorTable(spec, 'module_things'), [
    '| Operation | Endpoint | HTTP | `error.code` | When |',
    '|---|---|---|---|---|',
    '| `createThing` | `POST /v1/things` | 400 | `validation_failed` | Invalid input. |',
    '| `createThing` | `POST /v1/things` | 429 | `rate_limited` | Rate limit exceeded. |',
    '| `listThings` | `GET /v1/things` | 403 | `invalid_password`, `unauthorized` | Wrong \\| password. |',
    '',
  ].join('\n'));
});

test('a module without operations is an error, so a typo never empties the table', () => {
  assert.throws(() => errorTable(spec, 'module_typo'), /no operation/);
});

const marked = (body) => [
  'intro',
  '<!-- BEGIN GENERATED error-table module_things -->',
  body,
  '<!-- END GENERATED error-table module_things -->',
  'outro',
].join('\n');

test('replaces only the text between the markers', () => {
  assert.equal(replaceGenerated(marked('stale'), 'module_things', 'fresh\n'), marked('fresh'));
});

test('a document without the markers is an error', () => {
  assert.throws(() => replaceGenerated('no markers', 'module_things', 'x'), /markers/);
});
