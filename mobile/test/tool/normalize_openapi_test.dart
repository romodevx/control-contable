import 'package:flutter_test/flutter_test.dart';
import 'package:yaml/yaml.dart';

import '../../tool/src/openapi_normalizer.dart';

const _spec = r'''
openapi: 3.1.0
components:
  parameters:
    TaxpayerId:
      name: taxpayerId
      in: path
      required: true
      schema: { type: string }
paths:
  /v1/flags:
    get:
      operationId: listEnabledFlags
  /v1/taxpayers/{taxpayerId}/incomes:
    parameters:
      - $ref: '#/components/parameters/TaxpayerId'
      - name: X-Trace
        in: header
        schema: { type: string }
    summary: Incomes of one taxpayer
    get:
      operationId: listIncomes
      parameters:
        - name: period
          in: query
          schema: { type: string }
    post:
      operationId: createIncome
      parameters:
        - name: X-Trace
          in: header
          required: true
          schema: { type: string, maxLength: 8 }
''';

Map<String, Object?> _normalized() =>
    normalizeOpenApi(loadYaml(_spec) as Map<Object?, Object?>);

Map<String, Object?> _path(Map<String, Object?> spec, String path) =>
    (spec['paths']! as Map<String, Object?>)[path]! as Map<String, Object?>;

List<Object?> _paramsOf(Map<String, Object?> pathItem, String method) =>
    (pathItem[method]! as Map<String, Object?>)['parameters']! as List<Object?>;

void main() {
  test('path-level parameters move into every operation of the path', () {
    final item = _path(_normalized(), '/v1/taxpayers/{taxpayerId}/incomes');

    expect(item.containsKey('parameters'), isFalse);
    expect(_paramsOf(item, 'get'), [
      {r'$ref': '#/components/parameters/TaxpayerId'},
      {
        'name': 'X-Trace',
        'in': 'header',
        'schema': {'type': 'string'},
      },
      {
        'name': 'period',
        'in': 'query',
        'schema': {'type': 'string'},
      },
    ]);
  });

  test('an operation parameter overrides the path one with the same name '
      'and location (OAS allows overriding path-level parameters)', () {
    final item = _path(_normalized(), '/v1/taxpayers/{taxpayerId}/incomes');

    expect(_paramsOf(item, 'post'), [
      {r'$ref': '#/components/parameters/TaxpayerId'},
      {
        'name': 'X-Trace',
        'in': 'header',
        'required': true,
        'schema': {'type': 'string', 'maxLength': 8},
      },
    ]);
  });

  test('non-operation keys and paths without parameters are untouched', () {
    final spec = _normalized();

    expect(
      _path(spec, '/v1/taxpayers/{taxpayerId}/incomes')['summary'],
      'Incomes of one taxpayer',
    );
    expect(_path(spec, '/v1/flags'), {
      'get': {'operationId': 'listEnabledFlags'},
    });
  });

  test('the result is plain JSON-encodable maps, not YAML nodes', () {
    final spec = _normalized();

    expect(spec, isNot(isA<YamlMap>()));
    expect(spec['components'], isNot(isA<YamlMap>()));
  });

  // swagger_parser maps `format: date` to `DateTime`, whose
  // `toIso8601String()` sends `2026-09-10T00:00:00.000`: not a `date`, so
  // the server and Prism refuse it. Without the format the value stays the
  // `YYYY-MM-DD` string the contract defines.
  test('a calendar date (format: date) stays a string; a date-time keeps '
      'its format', () {
    const dates = '''
components:
  schemas:
    BusinessDate: { type: string, format: date }
    Stamp: { type: string, format: date-time }
    Holder:
      type: object
      properties:
        on: { type: string, format: date }
        list: { type: array, items: { type: string, format: date } }
paths:
  /x:
    get:
      parameters:
        - { name: on, in: query, schema: { type: string, format: date } }
''';
    final spec = normalizeOpenApi(loadYaml(dates) as Map<Object?, Object?>);
    final schemas =
        (spec['components']! as Map<String, Object?>)['schemas']!
            as Map<String, Object?>;
    final holder =
        (schemas['Holder']! as Map<String, Object?>)['properties']!
            as Map<String, Object?>;
    final param =
        _paramsOf(_path(spec, '/x'), 'get').single! as Map<String, Object?>;

    expect(schemas['BusinessDate'], {'type': 'string'});
    expect(schemas['Stamp'], {'type': 'string', 'format': 'date-time'});
    expect(holder['on'], {'type': 'string'});
    expect((holder['list']! as Map<String, Object?>)['items'], {
      'type': 'string',
    });
    expect(param['schema'], {'type': 'string'});
  });

  test(r'an unresolvable parameter $ref fails loudly', () {
    const broken = r'''
paths:
  /x/{id}:
    parameters:
      - $ref: '#/components/parameters/Missing'
    get: {}
''';

    expect(
      () => normalizeOpenApi(loadYaml(broken) as Map<Object?, Object?>),
      throwsA(isA<OpenApiNormalizeError>()),
    );
  });
}
