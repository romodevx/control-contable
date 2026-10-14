// Generates a module's error table (docs/modules/<module>.md) from the contract, so the table
// is never maintained by hand. Input is the dereferenced JSON bundle that
// `redocly bundle --dereferenced --ext json` writes. Run through the root Makefile:
//   make api-error-tables        rewrite the tables in place
//   make api-error-tables-check  fail if a committed table is stale (CI)
import { readFileSync, writeFileSync } from 'node:fs';
import { argv, exit } from 'node:process';
import { fileURLToPath } from 'node:url';

const HTTP_METHODS = ['get', 'put', 'post', 'delete', 'options', 'head', 'patch', 'trace'];
const JSON_MEDIA_TYPE = 'application/json';

const isError = (status) => !String(status).startsWith('2');
const cell = (text) => String(text).replaceAll('|', '\\|').replaceAll(/\s+/g, ' ').trim();

function errorCodes(response) {
  const media = response.content?.[JSON_MEDIA_TYPE];
  if (!media) return [];
  const bodies = media.example ? [media.example] : Object.values(media.examples ?? {}).map((e) => e.value);
  return [...new Set(bodies.map((body) => body?.error?.code).filter(Boolean))].sort();
}

function moduleOperations(spec, module) {
  return Object.entries(spec.paths).flatMap(([path, item]) =>
    Object.keys(item).filter((key) => HTTP_METHODS.includes(key) && item[key]['x-module'] === module)
      .map((method) => ({ path, method, operation: item[method] })));
}

/** Markdown table of every non-2xx response the module's operations declare, in contract order. */
export function errorTable(spec, module) {
  const operations = moduleOperations(spec, module);
  if (operations.length === 0) throw new Error(`no operation declares x-module: ${module}`);
  const rows = operations.flatMap(({ path, method, operation }) =>
    Object.entries(operation.responses).filter(([status]) => isError(status)).map(([status, response]) => {
      const codes = errorCodes(response).map((code) => `\`${code}\``).join(', ') || '—';
      return `| \`${operation.operationId}\` | \`${method.toUpperCase()} ${path}\` | ${status} | ${codes} | ${cell(response.description)} |`;
    }));
  return ['| Operation | Endpoint | HTTP | `error.code` | When |', '|---|---|---|---|---|', ...rows, ''].join('\n');
}

/** Replaces the text between the module's BEGIN/END GENERATED markers in doc. */
export function replaceGenerated(doc, module, generated) {
  const begin = `<!-- BEGIN GENERATED error-table ${module} -->\n`;
  const end = `<!-- END GENERATED error-table ${module} -->`;
  const from = doc.indexOf(begin);
  const to = doc.indexOf(end, from);
  if (from < 0 || to < 0) throw new Error(`${module}: BEGIN/END GENERATED markers not found`);
  return doc.slice(0, from + begin.length) + generated + doc.slice(to);
}

// CLI: node error-table.js <bundle.json> [--check] <module>=<doc.md>...
function main(args) {
  const [bundle, ...rest] = args;
  const check = rest.includes('--check');
  const spec = JSON.parse(readFileSync(bundle, 'utf8'));
  let stale = 0;
  for (const target of rest.filter((arg) => arg !== '--check')) {
    const [module, doc] = target.split('=');
    const current = readFileSync(doc, 'utf8');
    const next = replaceGenerated(current, module, errorTable(spec, module));
    if (next === current) continue;
    if (check) {
      console.error(`${doc}: the ${module} error table is stale; run make api-error-tables`);
      stale += 1;
    } else {
      writeFileSync(doc, next);
      console.log(`updated ${doc}`);
    }
  }
  if (stale > 0) exit(1);
}

if (argv[1] === fileURLToPath(import.meta.url)) main(argv.slice(2));
