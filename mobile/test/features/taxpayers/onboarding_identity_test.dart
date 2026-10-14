import 'dart:ui' show Tristate;

import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump_onboarding.dart';

const _addTaxpayer = 'Registrar contribuyente';
const _question = '¿Persona física o persona moral?';
const _fisica = 'Persona física';
const _moral = 'Persona moral';
const _name = 'Nombre para identificarlo';
const _rfc = 'RFC';
const _next = 'Siguiente';
const _required = 'Escribe este dato.';

void main() {
  late OnboardingHarness harness;

  setUp(() => harness = OnboardingHarness());

  group('flag module_taxpayers', () {
    testWidgets('off: home has no way in', (tester) async {
      await harness.pump(tester, enabled: false, location: null);

      expect(find.text(_addTaxpayer), findsNothing);
    });

    testWidgets('off: the route is not found', (tester) async {
      await harness.pump(tester, enabled: false);

      expect(find.text('Página no encontrada'), findsOneWidget);
      expect(harness.repository.catalogCalls, isEmpty);
    });

    testWidgets('on: home opens the onboarding at step 1', (tester) async {
      await harness.pump(tester, location: null);

      await tapText(tester, _addTaxpayer);

      expect(find.byType(OnboardingScreen), findsOneWidget);
      expect(find.text('Paso 1 de 4'), findsOneWidget);
      expect(find.text(_question), findsOneWidget);
    });
  });

  group('step 1: person type first (OD-3)', () {
    testWidgets('two large choices, each explained, and nothing else asked', (
      tester,
    ) async {
      await harness.pump(tester);

      expect(find.textContaining('13 caracteres'), findsOneWidget);
      expect(find.textContaining('12 caracteres'), findsOneWidget);
      expect(find.byType(TextField), findsNothing);
      for (final card in [_fisica, _moral]) {
        final size = tester.getSize(
          find.ancestor(of: find.text(card), matching: find.byType(InkWell)),
        );
        expect(size.height, greaterThanOrEqualTo(AppSizes.minTarget * 2));
      }
    });

    testWidgets('choosing one moves to the RFC, step 2', (tester) async {
      await harness.pump(tester);

      await tapText(tester, _moral);

      expect(find.text('Paso 2 de 4'), findsOneWidget);
      expect(find.text('RFC y nombre'), findsOneWidget);
      expect(find.textContaining('12 caracteres, como en'), findsOneWidget);
    });

    testWidgets('the chosen card reads as selected when coming back', (
      tester,
    ) async {
      await harness.pump(tester);
      await tapText(tester, _fisica);

      await tester.tap(find.byTooltip('Atrás'));
      await tester.pumpAndSettle();

      final semantics = tester.getSemantics(
        find.ancestor(of: find.text(_fisica), matching: find.byType(InkWell)),
      );
      expect(semantics.flagsCollection.isSelected, Tristate.isTrue);
    });

    testWidgets('back from step 1 closes the onboarding', (tester) async {
      await harness.pump(tester);

      await tester.tap(find.byTooltip('Atrás'));
      await tester.pumpAndSettle();

      expect(find.byType(OnboardingScreen), findsNothing);
      expect(find.text(_addTaxpayer), findsOneWidget);
    });
  });

  group('step 2: RFC and display name', () {
    Future<void> openStep2(WidgetTester tester, {String card = _fisica}) async {
      await harness.pump(tester);
      await tapText(tester, card);
    }

    testWidgets('the RFC field capitalizes and never autocorrects', (
      tester,
    ) async {
      await openStep2(tester);

      final input = tester.widget<TextField>(inputLabeled(_rfc));
      expect(input.textCapitalization, TextCapitalization.characters);
      expect(input.autocorrect, isFalse);
      expect(input.enableSuggestions, isFalse);
    });

    testWidgets('an RFC typed with spaces, dashes and lower case is '
        'upper-cased on blur', (tester) async {
      await openStep2(tester);

      await typeInto(tester, _rfc, ' goma-800101 ab6 ');
      await tester.tap(inputLabeled(_name));
      await tester.pumpAndSettle();

      expect(find.text('GOMA800101AB6'), findsOneWidget);
    });

    for (final (card, rfc, message) in [
      (
        _fisica,
        'ABC680524P76',
        'El RFC de una persona física tiene 13 caracteres.',
      ),
      (
        _moral,
        'GOMA800101AB6',
        'El RFC de una persona moral tiene 12 caracteres.',
      ),
      (_fisica, 'GOM4800101AB6', 'Revisa el RFC: primero las letras'),
      (_fisica, 'ABCD991332XX1', 'Los 6 números del RFC no forman una fecha'),
      (_fisica, 'XAXX010101000', 'Ese es un RFC genérico del SAT.'),
    ]) {
      testWidgets('$rfc as $card: format error on blur', (tester) async {
        await openStep2(tester, card: card);

        await typeInto(tester, _rfc, rfc);
        expect(find.textContaining(message), findsNothing, reason: 'typing');
        await tester.tap(inputLabeled(_name));
        await tester.pumpAndSettle();

        expect(find.textContaining(message), findsOneWidget);
      });
    }

    testWidgets('required-but-empty errors show only on submit', (
      tester,
    ) async {
      await openStep2(tester);
      expect(find.text(_required), findsNothing);

      await tapText(tester, _next);

      expect(find.text(_required), findsNWidgets(2));
      expect(find.text('Paso 2 de 4'), findsOneWidget, reason: 'stays');
    });

    testWidgets('a name over 80 characters is refused', (tester) async {
      await openStep2(tester);

      await typeInto(tester, _name, 'A' * 81);
      await typeInto(tester, _rfc, 'GOMA800101AB6');
      await tapText(tester, _next);

      expect(
        find.text('Escribe un nombre de hasta 80 caracteres.'),
        findsOneWidget,
      );
    });

    testWidgets('a valid RFC and name move to the regimes, step 3', (
      tester,
    ) async {
      await harness.pump(tester);

      await passIdentity(tester);

      expect(find.text('Paso 3 de 4'), findsOneWidget);
      expect(harness.repository.catalogCalls, [PersonType.fisica]);
    });

    testWidgets('back keeps what was typed, both ways', (tester) async {
      await harness.pump(tester);
      await passIdentity(tester);

      await tester.tap(find.byTooltip('Atrás'));
      await tester.pumpAndSettle();
      expect(find.text('GOMA800101AB6'), findsOneWidget);
      expect(find.text('Ana'), findsOneWidget);

      await typeInto(tester, _name, 'Ana López');
      await tester.tap(find.byTooltip('Atrás'));
      await tester.pumpAndSettle();
      await tapText(tester, _fisica);

      expect(find.text('Ana López'), findsOneWidget);
    });

    testWidgets('the system back button also goes back one step', (
      tester,
    ) async {
      await harness.pump(tester);
      await tapText(tester, _fisica);

      await tester.binding.handlePopRoute();
      await tester.pumpAndSettle();

      expect(find.text(_question), findsOneWidget);
    });
  });

  for (final brightness in Brightness.values) {
    testWidgets('steps 1 and 2 meet the tap-target, label and contrast '
        'guidelines (${brightness.name})', (tester) async {
      final semantics = tester.ensureSemantics();
      await harness.pump(tester, brightness: brightness);
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));

      await tapText(tester, _fisica);
      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
      semantics.dispose();
    });
  }

  testWidgets('with the tax engine on there are 5 steps', (tester) async {
    await harness.pump(tester, alsoEnabled: {AppModule.taxEngine});

    expect(find.text('Paso 1 de 5'), findsOneWidget);
  });
}
