// Pre-step of `make mobile-client` (ADR-0003 condition 1).
//
// swagger_parser 1.45.0 crashes on path-level `parameters` (it casts every
// path-item value to a Map before skipping that key). This copies them into
// each operation, which is spec-equivalent, so the contract keeps its
// path-level parameters. Delete once a fixed swagger_parser is pinned.
//
// It also drops `format: date`: swagger_parser maps it to `DateTime`, which
// serializes as `2026-09-10T00:00:00.000`, not the `YYYY-MM-DD` the
// contract defines (`BusinessDate`). Kept as a string, the client sends
// exactly what the contract says. `date-time` is untouched.

/// HTTP methods that hold an Operation Object inside a Path Item.
const _operationKeys = {
  'get',
  'put',
  'post',
  'delete',
  'options',
  'head',
  'patch',
  'trace',
};

const _parametersKey = 'parameters';
const _refKey = r'$ref';
const _parameterRefPrefix = '#/components/parameters/';
const _formatKey = 'format';
const _calendarDateFormat = 'date';

/// Raised when the contract cannot be normalized.
class OpenApiNormalizeError extends Error {
  /// Creates the error with a [message] for the developer.
  new(this.message);

  /// What is wrong.
  final String message;

  @override
  String toString() => 'OpenApiNormalizeError: $message';
}

/// Returns a plain-map copy of [spec] in which every path-level `parameters`
/// list is merged into each operation of that path and then removed, and
/// every `format: date` is dropped (see the file header).
///
/// An operation parameter with the same `name` and `in` overrides the path
/// one, as the OpenAPI Specification defines.
Map<String, Object?> normalizeOpenApi(Map<Object?, Object?> spec) {
  final plain = _toPlain(spec)! as Map<String, Object?>;
  _dropCalendarDateFormat(plain);
  final paths = plain['paths'];
  if (paths is! Map<String, Object?>) return plain;
  final componentParameters = _componentParameters(plain);
  for (final pathItem in paths.values.whereType<Map<String, Object?>>()) {
    _inlinePathParameters(pathItem, componentParameters);
  }
  return plain;
}

void _inlinePathParameters(
  Map<String, Object?> pathItem,
  Map<String, Object?> componentParameters,
) {
  final shared = pathItem.remove(_parametersKey);
  if (shared is! List<Object?> || shared.isEmpty) return;
  for (final entry in pathItem.entries) {
    final operation = entry.value;
    if (!_operationKeys.contains(entry.key)) continue;
    if (operation is! Map<String, Object?>) continue;
    final own = (operation[_parametersKey] as List<Object?>?) ?? const [];
    operation[_parametersKey] = _merge(shared, own, componentParameters);
  }
}

List<Object?> _merge(
  List<Object?> shared,
  List<Object?> own,
  Map<String, Object?> componentParameters,
) {
  final ownKeys = {for (final p in own) _identity(p, componentParameters)};
  return [
    for (final p in shared)
      if (!ownKeys.contains(_identity(p, componentParameters))) _toPlain(p),
    ...own,
  ];
}

/// A parameter is identified by its `name` and `in` (OAS 3.1, Path Item).
String _identity(Object? parameter, Map<String, Object?> componentParameters) {
  final resolved = _resolve(parameter, componentParameters);
  return '${resolved['in']}:${resolved['name']}';
}

Map<String, Object?> _resolve(
  Object? parameter,
  Map<String, Object?> componentParameters,
) {
  if (parameter is! Map<String, Object?>) {
    throw OpenApiNormalizeError('parameter is not an object: $parameter');
  }
  final ref = parameter[_refKey];
  if (ref == null) return parameter;
  final target = ref is String && ref.startsWith(_parameterRefPrefix)
      ? componentParameters[ref.substring(_parameterRefPrefix.length)]
      : null;
  if (target is! Map<String, Object?>) {
    throw OpenApiNormalizeError('cannot resolve parameter $ref "$ref"');
  }
  return target;
}

Map<String, Object?> _componentParameters(Map<String, Object?> spec) {
  final components = spec['components'];
  if (components is! Map<String, Object?>) return const {};
  final parameters = components[_parametersKey];
  return parameters is Map<String, Object?> ? parameters : const {};
}

/// Removes `format: date` from every schema under [node], at any depth.
void _dropCalendarDateFormat(Object? node) {
  switch (node) {
    case final Map<String, Object?> map:
      if (map[_formatKey] == _calendarDateFormat) map.remove(_formatKey);
      map.values.forEach(_dropCalendarDateFormat);
    case final List<Object?> list:
      list.forEach(_dropCalendarDateFormat);
  }
}

/// Deep-copies YAML nodes (or any map/list tree) into plain Dart collections
/// with string keys, so the result is JSON-encodable.
Object? _toPlain(Object? node) => switch (node) {
  final Map<Object?, Object?> map => <String, Object?>{
    for (final e in map.entries) e.key.toString(): _toPlain(e.value),
  },
  final List<Object?> list => [for (final item in list) _toPlain(item)],
  _ => node,
};
