import 'package:control_contable/core/forms/form_notice.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_form.dart';

/// A host that shows a notice reading [message], or none while it is null.
class _Host extends StatelessWidget {
  const new({required this.message, this.tone = FormNoticeTone.error});

  final String? message;
  final FormNoticeTone tone;

  @override
  Widget build(BuildContext context) => Column(
    children: [
      if (message case final message?) FormNotice(message: message, tone: tone),
    ],
  );
}

void main() {
  group('FormNotice announces itself to screen readers', () {
    // A live region alone is not enough: VoiceOver ignores it, and TalkBack
    // stays silent when the text repeats (tech-lead review, PR #26 / #39).
    testWidgets('when it appears, assertively for an error', (tester) async {
      await pumpForm(tester, const _Host(message: null));
      tester.takeAnnouncements();

      await pumpForm(tester, const _Host(message: 'Sin conexión.'));

      final announced = tester.takeAnnouncements();
      expect(announced.map((a) => a.message), ['Sin conexión.']);
      expect(announced.single.assertiveness, Assertiveness.assertive);
    });

    testWidgets('politely for a confirmation', (tester) async {
      await pumpForm(
        tester,
        const _Host(
          message: 'Te enviamos otro código.',
          tone: FormNoticeTone.info,
        ),
      );

      final announced = tester.takeAnnouncements();
      expect(announced.map((a) => a.message), ['Te enviamos otro código.']);
      expect(announced.single.assertiveness, Assertiveness.polite);
    });

    testWidgets('again when its message changes, not on a plain rebuild', (
      tester,
    ) async {
      await pumpForm(tester, const _Host(message: 'Sin conexión.'));
      tester.takeAnnouncements();

      await pumpForm(tester, const _Host(message: 'Sin conexión.'));
      await pumpForm(tester, const _Host(message: 'Demasiados intentos.'));

      expect(tester.takeAnnouncements().map((a) => a.message), [
        'Demasiados intentos.',
      ]);
    });
  });
}
