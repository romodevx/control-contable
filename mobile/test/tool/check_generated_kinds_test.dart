import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:swagger_parser/swagger_parser.dart';
import 'package:yaml/yaml.dart';

import '../../tool/src/generated_kinds.dart';

const _fixtures = 'test/tool/fixtures';

Map<String, Object?> _specOf(String fixture) {
  final yaml = loadYaml(File('$_fixtures/$fixture').readAsStringSync());
  return _plain(yaml)! as Map<String, Object?>;
}

Object? _plain(Object? node) => switch (node) {
  final Map<Object?, Object?> map => <String, Object?>{
    for (final e in map.entries) e.key.toString(): _plain(e.value),
  },
  final List<Object?> list => [for (final item in list) _plain(item)],
  _ => node,
};

/// Runs the real swagger_parser, with the project's own options
/// (swagger_parser.yaml), over [fixture]. Files are keyed by path and a later
/// file with the same path replaces the earlier one, as on disk.
Future<Map<String, String>> _generate(String fixture) async {
  final options = loadYaml(File('swagger_parser.yaml').readAsStringSync());
  final config = SWPConfig.fromYaml(
    (options as YamlMap)['swagger_parser'] as YamlMap,
  );
  final files = await GenProcessor(config).generateContent((
    fileContent: File('$_fixtures/$fixture').readAsStringSync(),
    isJson: false,
  ));
  return {for (final f in files) f.name: f.content};
}

void main() {
  test('an inline enum that takes over a component name fails the guard '
      '(the Period collision from ADR-0003)', () async {
    const fixture = 'inline_enum_collision.yaml';

    final findings = checkGeneratedKinds(
      spec: _specOf(fixture),
      sources: await _generate(fixture),
    );

    expect(
      findings,
      contains(
        allOf(contains('Period'), contains('typedef'), contains('enum')),
      ),
    );
  });

  test('a contract that names its inline objects passes the guard', () async {
    const fixture = 'named_plan.yaml';

    final findings = checkGeneratedKinds(
      spec: _specOf(fixture),
      sources: await _generate(fixture),
    );

    expect(findings, isEmpty);
  });

  test('two component schemas that map to one Dart name fail the guard', () {
    final findings = checkGeneratedKinds(
      spec: {
        'components': {
          'schemas': {
            'sub_plan': {'type': 'string'},
            'SubPlan': {'type': 'string'},
          },
        },
      },
      sources: {'models/sub_plan.dart': 'typedef SubPlan = String;'},
    );

    expect(
      findings,
      contains(allOf(contains('sub_plan'), contains('SubPlan'))),
    );
  });

  test('one Dart name declared by two generated files fails the guard', () {
    final findings = checkGeneratedKinds(
      spec: {
        'components': {
          'schemas': {
            'Error': {
              'type': 'object',
              'properties': {
                'code': {'type': 'string'},
              },
            },
          },
        },
      },
      sources: {
        'models/error.dart': r'abstract class Error with _$Error {}',
        'models/error_2.dart': 'enum Error { a, b }',
      },
    );

    expect(findings, contains(contains('declared 2 times')));
  });

  test('a component without a generated declaration fails the guard', () {
    final findings = checkGeneratedKinds(
      spec: {
        'components': {
          'schemas': {
            'Uuid': {'type': 'string', 'format': 'uuid'},
          },
        },
      },
      sources: const {},
    );

    expect(findings, [contains('Uuid')]);
  });

  test('part files (*.g.dart, *.freezed.dart) are not declarations', () {
    final findings = checkGeneratedKinds(
      spec: {
        'components': {
          'schemas': {
            'PersonType': {
              'type': 'string',
              'enum': ['fisica', 'moral'],
            },
          },
        },
      },
      sources: {
        'models/person_type.dart': 'enum PersonType { fisica, moral }',
        'models/person_type.g.dart': 'class PersonType {}',
      },
    );

    expect(findings, isEmpty);
  });
}
