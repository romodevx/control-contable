import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/features/taxpayers/domain/taxpayer_cap.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_screen.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_onboarding.dart';
import 'taxpayers_fakes.dart';

const String _limitTitle = 'Llegaste al máximo de contribuyentes';
const String _personTypeQuestion = '¿Persona física o persona moral?';

/// [count] taxpayers with distinct ids; the last one archived (it still
/// counts toward the cap, OQ-9).
List<Taxpayer> _held(int count) => [
  for (var i = 0; i < count; i++)
    taxpayer(
      id: '0192f0c4-7b1e-7c3a-9d2e-${i.toString().padLeft(12, '0')}',
      displayName: 'Cliente $i',
      archivedAt: i == count - 1 ? DateTime.utc(2026, 10, 2) : null,
    ),
];

/// QA M2-12 D1: the cap of 10 (OD-2, archived included) is said before
/// the onboarding asks anything; the server's `422` stays the authority.
void main() {
  late OnboardingHarness harness;

  setUp(() => harness = OnboardingHarness());

  test('the cap is 10 taxpayers, archived included', () {
    expect(taxpayerCap, 10);
    expect(isAtTaxpayerCap(9), isFalse);
    expect(isAtTaxpayerCap(10), isTrue);
  });

  testWidgets('at the cap: the limit instead of step 1, nothing loaded', (
    tester,
  ) async {
    await harness.pump(tester, taxpayers: _held(taxpayerCap));

    expect(find.text(_limitTitle), findsOneWidget);
    expect(find.textContaining('archivados también cuentan'), findsOneWidget);
    expect(find.text(_personTypeQuestion), findsNothing);
    expect(harness.repository.catalogCalls, isEmpty);

    await tapText(tester, 'Volver al inicio');

    expect(find.byType(OnboardingScreen), findsNothing);
    expect(harness.repository.creates, isEmpty);
  });

  testWidgets('below the cap: step 1 as usual', (tester) async {
    await harness.pump(tester, taxpayers: _held(taxpayerCap - 1));

    expect(find.text(_personTypeQuestion), findsOneWidget);
    expect(find.text(_limitTitle), findsNothing);
  });
}
