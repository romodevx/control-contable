// Guard of `make mobile-client` (ADR-0003 condition 3).
//
// swagger_parser names inline enums after their property and writes them to
// the same folder as the components, so an inline enum can silently replace
// a component of the same name (it happened: `Period`). Everything still
// compiles; values then decode as `$unknown`. This compares what each
// component schema should become in Dart with what was generated.

/// What a schema becomes in the generated Dart code.
enum DartKind {
  /// A Dart `enum` (schema with `enum`).
  enumeration('enum'),

  /// A Freezed data class (schema with `properties` or `allOf`).
  dataClass('class'),

  /// A `typedef` alias (scalars, arrays, maps, `const`).
  alias('typedef');

  new(this.keyword);

  /// The Dart keyword that declares this kind.
  final String keyword;
}

/// The kind swagger_parser must generate for a component [schema].
DartKind expectedKindOf(Map<String, Object?> schema) {
  if (schema.containsKey('enum')) return DartKind.enumeration;
  if (schema.containsKey('properties') || schema.containsKey('allOf')) {
    return DartKind.dataClass;
  }
  return DartKind.alias;
}

/// The Dart type name for a component schema name (PascalCase).
String dartNameOf(String schemaName) => schemaName
    .split(RegExp('[^A-Za-z0-9]+'))
    .where((part) => part.isNotEmpty)
    .map((part) => part[0].toUpperCase() + part.substring(1))
    .join();

final _declaration = RegExp(
  r'^(?:(?:abstract|sealed|final|base|interface)\s+)*(class|enum|typedef)\s+([A-Za-z]\w*)',
  multiLine: true,
);
final _partFile = RegExp(r'\.(g|freezed)\.dart$');

typedef _Declaration = ({DartKind kind, String path});

/// Returns one human-readable finding per problem; empty means clean.
///
/// [spec] is the (normalized) contract; [sources] maps each generated file
/// path to its content. Part files (`*.g.dart`, `*.freezed.dart`) are ignored.
List<String> checkGeneratedKinds({
  required Map<String, Object?> spec,
  required Map<String, String> sources,
}) {
  final schemas = _componentSchemas(spec);
  final declarations = _declarationsIn(sources);
  return [
    ..._nameCollisions(schemas.keys),
    ..._duplicateDeclarations(declarations),
    for (final entry in schemas.entries)
      ?_kindMismatch(entry.key, entry.value, declarations),
  ];
}

Map<String, Map<String, Object?>> _componentSchemas(Map<String, Object?> spec) {
  final components = spec['components'];
  final schemas = components is Map<String, Object?>
      ? components['schemas']
      : null;
  if (schemas is! Map<String, Object?>) return const {};
  return {
    for (final e in schemas.entries)
      if (e.value case final Map<String, Object?> schema) e.key: schema,
  };
}

Map<String, List<_Declaration>> _declarationsIn(Map<String, String> sources) {
  final found = <String, List<_Declaration>>{};
  for (final MapEntry(key: path, value: content) in sources.entries) {
    if (_partFile.hasMatch(path)) continue;
    for (final match in _declaration.allMatches(content)) {
      final kind = DartKind.values.firstWhere((k) => k.keyword == match[1]);
      (found[match[2]!] ??= []).add((kind: kind, path: path));
    }
  }
  return found;
}

Iterable<String> _nameCollisions(Iterable<String> schemaNames) sync* {
  final byDartName = <String, List<String>>{};
  for (final name in schemaNames) {
    (byDartName[dartNameOf(name)] ??= []).add(name);
  }
  for (final MapEntry(key: dartName, value: names) in byDartName.entries) {
    if (names.length > 1) {
      yield '${names.join(', ')}: all map to the Dart name $dartName';
    }
  }
}

Iterable<String> _duplicateDeclarations(
  Map<String, List<_Declaration>> declarations,
) sync* {
  for (final MapEntry(key: name, value: found) in declarations.entries) {
    if (found.length > 1) {
      final where = found.map((d) => '${d.kind.keyword} in ${d.path}');
      yield '$name: declared ${found.length} times (${where.join('; ')})';
    }
  }
}

String? _kindMismatch(
  String schemaName,
  Map<String, Object?> schema,
  Map<String, List<_Declaration>> declarations,
) {
  final dartName = dartNameOf(schemaName);
  final expected = expectedKindOf(schema);
  final found = declarations[dartName];
  if (found == null) {
    return '$schemaName: no generated declaration of $dartName '
        '(expected ${expected.keyword})';
  }
  final wrong = found.where((d) => d.kind != expected);
  if (wrong.isEmpty) return null;
  final actual = wrong.map((d) => '${d.kind.keyword} in ${d.path}').join('; ');
  return '$schemaName: the component is a ${expected.keyword} but $dartName '
      'was generated as $actual (an inline schema took over the name?)';
}
