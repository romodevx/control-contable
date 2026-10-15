import 'dart:io';

import 'package:control_contable/core/widgets/captioned_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/spoken_labels.dart';

const String _caption = 'Ejercicio fiscal';
const String _home = 'lib/core/widgets/captioned_picker.dart';

/// A caption wrapped in `Semantics(label:)` straight around a dropdown: the
/// pattern [CaptionedPicker] owns.
final RegExp _handRolledLabel = RegExp(
  r'Semantics\(\s*label:[^;]*?child:\s*DropdownButtonFormField',
);

void main() {
  testWidgets('shows the caption and announces it once', (tester) async {
    final semantics = tester.ensureSemantics();
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: CaptionedPicker(
            caption: _caption,
            picker: DropdownButtonFormField<int>(
              initialValue: 2026,
              items: const [DropdownMenuItem(value: 2026, child: Text('2026'))],
              onChanged: (_) {},
            ),
          ),
        ),
      ),
    );

    expect(find.text(_caption), findsOneWidget);
    expect(timesSpoken(tester, _caption), 1);
    semantics.dispose();
  });

  // Review of #174: the doubled label was fixed by hand in five copies of
  // the same caption + Semantics pattern; one widget owns it now.
  test('no screen hand-rolls a labeled dropdown outside CaptionedPicker', () {
    final offenders = [
      for (final file in Directory('lib').listSync(recursive: true))
        if (file is File &&
            file.path.endsWith('.dart') &&
            file.path != _home &&
            _handRolledLabel.hasMatch(file.readAsStringSync()))
          file.path,
    ];
    expect(offenders, isEmpty);
  });
}
