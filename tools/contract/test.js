// Tests for the contract lint ruleset (redocly.yaml + plugin.js).
// Each case mutates the valid fixture so it breaks exactly one rule, then asserts that
// `redocly lint` reports exactly that rule. Run: `make contract-lint-test` (node --test).
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const repoRoot = join(here, '..', '..');
const config = join(repoRoot, 'redocly.yaml');
const redocly = join(here, 'node_modules', '.bin', 'redocly');
const validFixture = readFileSync(join(here, 'testdata', 'valid.yaml'), 'utf8');
const workDir = mkdtempSync(join(tmpdir(), 'contract-lint-'));

function lint(source, name) {
  const file = join(workDir, `${name}.yaml`);
  writeFileSync(file, source);
  const run = spawnSync(redocly, ['lint', file, '--config', config, '--format', 'json'], {
    encoding: 'utf8',
    env: { ...process.env, REDOCLY_SUPPRESS_UPDATE_NOTICE: 'true', REDOCLY_TELEMETRY: 'off' },
  });
  const report = JSON.parse(run.stdout);
  return [...new Set(report.problems.map((problem) => problem.ruleId))].sort();
}

function mutate(search, replacement) {
  assert.ok(validFixture.includes(search), `fixture does not contain: ${search}`);
  return validFixture.replace(search, replacement);
}

const ERROR_400_BLOCK = `        '400':
          description: Invalid.
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ErrorResponse' }
              example: { error: { code: validation_failed, message: Invalid. } }
`;

const cases = [
  {
    name: 'missing x-module',
    source: () => mutate('      x-module: module_things\n', ''),
    expect: ['rule/operation-x-module'],
  },
  {
    name: 'x-module that is not a flag key',
    source: () => mutate('x-module: module_things', 'x-module: things'),
    expect: ['rule/operation-x-module'],
  },
  {
    name: 'missing operationId',
    source: () => mutate('      operationId: createThing\n', ''),
    expect: ['operation-operationId'],
  },
  {
    name: 'inline 4xx error schema instead of the shared envelope',
    source: () => mutate(ERROR_400_BLOCK, ERROR_400_BLOCK.replace('ErrorResponse', 'Status').replace('example: { error: { code: validation_failed, message: Invalid. } }', 'example: { status: ok }')),
    expect: ['rule/error-response-envelope'],
  },
  {
    name: 'shared 5xx response component not using the envelope',
    source: () => mutate(
      `          schema: { $ref: '#/components/schemas/ErrorResponse' }
          example: { error: { code: internal_error, message: Something went wrong. } }`,
      `          schema: { $ref: '#/components/schemas/Status' }
          example: { status: ok }`,
    ),
    expect: ['rule/error-response-envelope'],
  },
  {
    name: '4xx response without a body',
    source: () => mutate(ERROR_400_BLOCK, "        '400':\n          description: Invalid.\n"),
    expect: ['rule/error-response-has-body'],
  },
  {
    name: '4xx response with an empty content map',
    source: () => mutate(ERROR_400_BLOCK, "        '400':\n          description: Invalid.\n          content: {}\n"),
    expect: ['rule/error-response-has-body'],
  },
  {
    name: '4xx body without a schema',
    source: () => mutate(ERROR_400_BLOCK, ERROR_400_BLOCK.replace("              schema: { $ref: '#/components/schemas/ErrorResponse' }\n", '')),
    expect: ['rule/error-response-envelope'],
  },
  {
    name: 'body without any example',
    source: () => mutate('      examples:\n        - { status: ok }\n', ''),
    expect: ['contract/body-example', 'contract/frozen-unchanged'],
  },
  {
    name: 'security opt-out without x-public',
    source: () => mutate('      x-public: true\n', ''),
    expect: ['contract/public-operation-explicit'],
  },
  {
    name: 'x-public on an authenticated operation',
    source: () => mutate('      x-module: module_things\n', '      x-module: module_things\n      x-public: true\n'),
    expect: ['contract/public-operation-explicit'],
  },
  {
    name: 'operation without any security requirement',
    source: () => mutate('security:\n  - bearerAuth: []\n', ''),
    expect: ['no-unused-components', 'security-defined'],
  },
  {
    name: 'frozen schema changed',
    source: () => mutate('enum: [ok]', 'enum: [ok, degraded]'),
    expect: ['contract/frozen-unchanged'],
  },
  {
    name: 'frozen marker outside operations and schemas that does not match',
    source: () => mutate(
      '    TooManyRequests:\n      description: Rate limited.',
      '    TooManyRequests:\n      x-frozen: sha256:0000\n      description: Rate limited.',
    ),
    expect: ['contract/frozen-unchanged'],
  },
];

test('valid fixture passes every rule', () => {
  assert.deepEqual(lint(validFixture, 'valid'), []);
});

for (const { name, source, expect } of cases) {
  test(`rejects: ${name}`, () => {
    assert.deepEqual(lint(source(), name.replaceAll(/\W+/g, '-')), [...expect].sort());
  });
}
