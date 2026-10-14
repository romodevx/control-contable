import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/domain/rfc.dart';
import 'package:flutter_test/flutter_test.dart';

import 'taxpayers_fakes.dart';

const _clientId = '0192f0c4-0000-4000-8000-000000000001';
const _rfcMismatch = Warning(
  code: 'rfc_check_digit_mismatch',
  field: 'rfc',
  message: 'Revisa tu RFC.',
);
const _cpaReview = Warning(
  code: 'regime_combination_needs_cpa_review',
  field: 'regimes.1.regime_code',
  message: 'Combinación por confirmar con un contador.',
);

Regime _byCode(List<Regime> catalog, String code) =>
    catalog.firstWhere((regime) => regime.code == code);

void main() {
  late FakeTaxpayersRepository repository;
  late OnboardingFlow flow;
  var now = DateTime.utc(2027, 1, 1, 3);

  OnboardingFlow newFlow({bool taxEngineEnabled = false}) => OnboardingFlow(
    repository: repository,
    newId: () => _clientId,
    clock: () => now,
    taxEngineEnabled: taxEngineEnabled,
  );

  setUp(() {
    repository = FakeTaxpayersRepository();
    now = DateTime.utc(2027, 1, 1, 3);
    flow = newFlow();
  });
  tearDown(() => flow.dispose());

  void fillIdentity({
    PersonType type = PersonType.fisica,
    String rfc = ' goma-800101 ab6 ',
  }) {
    flow
      ..choosePersonType(type)
      ..submitIdentity(displayName: ' Ana ', rfc: rfc);
  }

  group('steps (OD-3: person type first)', () {
    test('starts at the person type, step 1 of 4', () {
      expect(flow.step, OnboardingStep.personType);
      expect(flow.stepNumber, 1);
      expect(flow.totalSteps, 4);
    });

    test('with the tax engine on there is a fifth, optional step', () {
      final withInputs = newFlow(taxEngineEnabled: true);
      addTearDown(withInputs.dispose);

      expect(withInputs.totalSteps, 5);
    });

    test('person type → RFC and name → regimes', () {
      flow.choosePersonType(PersonType.fisica);
      expect(flow.step, OnboardingStep.identity);
      expect(flow.stepNumber, 2);

      flow.submitIdentity(displayName: 'Ana', rfc: 'GOMA800101AB6');
      expect(flow.step, OnboardingStep.regimes);
      expect(flow.stepNumber, 3);
    });

    test('regimes with options → one options page per such regime', () async {
      fillIdentity();
      final catalog = fisicaCatalog;

      await flow.submitRegimes([
        _byCode(catalog, '606'),
        _byCode(catalog, '605'),
        _byCode(catalog, '625'),
      ]);

      expect(flow.step, OnboardingStep.options);
      expect(flow.stepNumber, 4);
      expect(flow.optionRegimes.map((r) => r.code), ['606', '625']);
      expect(flow.page, 0);
      expect(repository.creates, isEmpty);
    });

    test('regimes without options create the taxpayer right away', () async {
      fillIdentity();

      await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

      expect(repository.creates, hasLength(1));
      expect(flow.step, OnboardingStep.done);
    });
  });

  group('back keeps the data', () {
    test('from regimes to the RFC, then to the person type', () async {
      fillIdentity();

      expect(flow.back(), isTrue);
      expect(flow.step, OnboardingStep.identity);
      expect(flow.displayName, ' Ana ');
      expect(flow.rfc, ' goma-800101 ab6 ');

      expect(flow.back(), isTrue);
      expect(flow.step, OnboardingStep.personType);
      expect(flow.personType, PersonType.fisica);
    });

    test('the first step lets the route close', () {
      expect(flow.back(), isFalse);
    });

    test('options pages go back one at a time, then to regimes', () async {
      fillIdentity();
      final chosen = [
        _byCode(fisicaCatalog, '606'),
        _byCode(fisicaCatalog, '612'),
      ];
      await flow.submitRegimes(chosen);
      await flow.submitOptions('606', {'deduction_mode': 'actual'});
      expect(flow.page, 1);

      expect(flow.back(), isTrue);
      expect(flow.page, 0);
      expect(flow.optionsFor(chosen.first), {'deduction_mode': 'actual'});

      expect(flow.back(), isTrue);
      expect(flow.step, OnboardingStep.regimes);
      expect(flow.regimes, chosen);
    });

    test('remembered values survive leaving a step without submitting', () {
      flow
        ..choosePersonType(PersonType.fisica)
        ..rememberIdentity(displayName: 'Ana', rfc: 'GOMA');

      expect(flow.back(), isTrue);
      flow.choosePersonType(PersonType.fisica);

      expect(flow.displayName, 'Ana');
      expect(flow.rfc, 'GOMA');
    });

    test(
      'changing the person type drops the regimes of the other one',
      () async {
        fillIdentity();
        await flow.submitRegimes([_byCode(fisicaCatalog, '606')]);
        flow
          ..back()
          ..back()
          ..back()
          ..choosePersonType(PersonType.moral);

        expect(flow.regimes, isEmpty);
        expect(flow.rfc, ' goma-800101 ab6 ', reason: 'the typed RFC stays');
      },
    );

    test('choosing the same person type again keeps the regimes', () async {
      fillIdentity();
      final chosen = [_byCode(fisicaCatalog, '606')];
      await flow.submitRegimes(chosen);
      flow
        ..back()
        ..back()
        ..back()
        ..choosePersonType(PersonType.fisica);

      expect(flow.regimes, chosen);
    });
  });

  group('options', () {
    test('start from the catalog defaults; required ones without a default '
        'start empty', () {
      final rentals = _byCode(fisicaCatalog, '606');
      final platforms = _byCode(fisicaCatalog, '625');

      expect(flow.optionsFor(rentals), {'deduction_mode': 'blind_35'});
      expect(flow.optionsFor(platforms), {
        'definitive_option': false,
        'rfc_given_to_platform': true,
      });
    });

    test('remembered options replace the defaults', () {
      final rentals = _byCode(fisicaCatalog, '606');

      flow.rememberOptions('606', {'deduction_mode': 'actual'});

      expect(flow.optionsFor(rentals), {'deduction_mode': 'actual'});
    });
  });

  group('create (one POST: taxpayer + regimes + options)', () {
    test('sends the normalized RFC, the trimmed name, every regime with its '
        'options and a client id for retries', () async {
      fillIdentity();
      await flow.submitRegimes([
        _byCode(fisicaCatalog, '606'),
        _byCode(fisicaCatalog, '625'),
      ]);
      await flow.submitOptions('606', {'deduction_mode': 'actual'});
      await flow.submitOptions('625', {
        'activities': ['lodging'],
        'definitive_option': false,
        'rfc_given_to_platform': true,
      });

      final request = repository.creates.single;
      expect(request.personType, PersonType.fisica);
      expect(request.rfc, 'GOMA800101AB6');
      expect(request.displayName, 'Ana');
      expect(request.id, _clientId);
      expect(request.regimes.map((r) => r.regimeCode), ['606', '625']);
      expect(request.regimes.first.options, {'deduction_mode': 'actual'});
      expect(request.regimes.last.options?['activities'], ['lodging']);
      expect(request.regimes.first.startPeriod, isNull);
      expect(flow.step, OnboardingStep.done);
    });

    test('a regime without options sends none', () async {
      fillIdentity();

      await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

      expect(repository.creates.single.regimes.single.options, isNull);
    });

    test('a retry after a failure sends the same client id', () async {
      fillIdentity();
      repository.createAnswers.add(const ApiUnreachable());
      final resico = _byCode(fisicaCatalog, '626');
      await flow.submitRegimes([resico]);
      expect(flow.notice, OnboardingNotice.unreachable);

      await flow.submitRegimes([resico]);

      expect(repository.creates.map((r) => r.id), [_clientId, _clientId]);
      expect(flow.notice, isNull);
    });

    test('402: the read-only user sees the paywall, nothing else', () async {
      fillIdentity();
      repository.createAnswers.add(const ApiEntitlementRequired());

      await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

      expect(flow.step, OnboardingStep.readOnly);
      expect(flow.back(), isFalse);
    });

    test(
      '409 rfc_already_registered: back to the RFC with its error',
      () async {
        fillIdentity();
        repository.createAnswers.add(
          const ApiRejected(code: ErrorCode.rfcAlreadyRegistered),
        );

        await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

        expect(flow.step, OnboardingStep.identity);
        expect(flow.rfcRejection, RfcRejection.duplicate);
      },
    );

    for (final (code, rejection) in [
      ('rfc_length_mismatch', RfcRejection.length),
      ('rfc_format', RfcRejection.format),
      ('rfc_date', RfcRejection.date),
      ('rfc_generic', RfcRejection.generic),
      ('pattern', RfcRejection.format),
    ]) {
      test('400 rfc/$code: back to the RFC with its error', () async {
        fillIdentity();
        repository.createAnswers.add(
          ApiValidationFailure([detail('rfc', code)]),
        );

        await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

        expect(flow.step, OnboardingStep.identity);
        expect(flow.rfcRejection, rejection);
      });
    }

    test('400 display_name: back to the name with its error', () async {
      fillIdentity();
      repository.createAnswers.add(
        ApiValidationFailure([detail('display_name', 'maxLength')]),
      );

      await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

      expect(flow.step, OnboardingStep.identity);
      expect(flow.displayNameRejected, isTrue);
      expect(flow.rfcRejection, isNull);
    });

    test(
      '400 on an option: back to that regime page, the option marked',
      () async {
        fillIdentity();
        await flow.submitRegimes([
          _byCode(fisicaCatalog, '606'),
          _byCode(fisicaCatalog, '621'),
        ]);
        await flow.submitOptions('606', {'deduction_mode': 'blind_35'});
        repository.createAnswers.add(
          ApiValidationFailure([
            detail('regimes.1.options.rif_year', 'required'),
          ]),
        );

        await flow.submitOptions('621', {'rif_year': '2'});

        expect(flow.step, OnboardingStep.options);
        expect(flow.page, 1);
        expect(flow.optionRejections, {
          '621': {'rif_year'},
        });
      },
    );

    test('400 on a regime code: back to regimes with a notice', () async {
      fillIdentity();
      repository.createAnswers.add(
        ApiValidationFailure([
          detail('regimes.0.regime_code', 'not_for_person_type'),
        ]),
      );

      await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

      expect(flow.step, OnboardingStep.regimes);
      expect(flow.notice, OnboardingNotice.regimeRejected);
    });

    test(
      '422 invalid_regime_combination: back to regimes with a notice',
      () async {
        fillIdentity();
        repository.createAnswers.add(
          const ApiRejected(code: ErrorCode.invalidRegimeCombination),
        );

        await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

        expect(flow.step, OnboardingStep.regimes);
        expect(flow.notice, OnboardingNotice.combinationRejected);
      },
    );

    for (final (failure, notice) in [
      (
        const ApiRejected(code: ErrorCode.taxpayerLimitReached),
        OnboardingNotice.limitReached,
      ),
      (const ApiRateLimited(null), OnboardingNotice.rateLimited),
      (const ApiUnreachable(), OnboardingNotice.unreachable),
      (const ApiUnexpected(), OnboardingNotice.unexpected),
      (
        const ApiRejected(code: ErrorCode.conflict),
        OnboardingNotice.unexpected,
      ),
      (
        ApiValidationFailure([detail('id', 'format')]),
        OnboardingNotice.unexpected,
      ),
    ]) {
      test(
        '${failure.runtimeType} stays on the screen with a notice',
        () async {
          fillIdentity();
          repository.createAnswers.add(failure);

          await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

          expect(flow.step, OnboardingStep.regimes);
          expect(flow.notice, notice);
        },
      );
    }

    test('a step change clears the notice', () async {
      fillIdentity();
      repository.createAnswers.add(const ApiUnreachable());
      await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);

      flow.back();

      expect(flow.notice, isNull);
    });
  });

  group('after create: warnings (non-blocking)', () {
    Future<void> createWith(List<Warning> warnings) async {
      fillIdentity();
      repository.createAnswers.add(taxpayer(warnings: warnings));
      await flow.submitRegimes([_byCode(fisicaCatalog, '626')]);
    }

    test('a check-digit mismatch asks to review before continuing', () async {
      await createWith([_rfcMismatch]);

      expect(flow.step, OnboardingStep.review);
      expect(flow.rfcWarning, isTrue);
      expect(flow.created?.rfcMasked, 'GOMA******AB6');
      expect(flow.back(), isFalse, reason: 'the taxpayer already exists');
    });

    test('a CPA-review warning is listed with the server reason', () async {
      await createWith([_cpaReview]);

      expect(flow.step, OnboardingStep.review);
      expect(flow.rfcWarning, isFalse);
      expect(flow.cpaReviewReasons, [_cpaReview.message]);
    });

    test('confirming the RFC as typed continues without a call', () async {
      await createWith([_rfcMismatch]);

      await flow.confirmReview(rfc: 'goma800101ab6');

      expect(repository.corrections, isEmpty);
      expect(flow.step, OnboardingStep.done);
    });

    test('a corrected RFC is sent normalized, then continues', () async {
      await createWith([_rfcMismatch, _cpaReview]);
      repository.correctAnswers.add(taxpayer(rfcMasked: 'GOMA******AB7'));

      await flow.confirmReview(rfc: 'goma800101ab7');

      expect(repository.corrections.single.rfc, 'GOMA800101AB7');
      expect(repository.corrections.single.taxpayerId, taxpayer().id);
      expect(flow.step, OnboardingStep.done);
    });

    test('a correction that still mismatches asks again', () async {
      await createWith([_rfcMismatch]);
      repository.correctAnswers.add(taxpayer(warnings: [_rfcMismatch]));

      await flow.confirmReview(rfc: 'GOMA800101AB8');

      expect(flow.step, OnboardingStep.review);
      expect(flow.rfc, 'GOMA800101AB8');
      expect(flow.created?.warnings, [_rfcMismatch]);
    });

    test('a correction the server refuses shows on the RFC', () async {
      await createWith([_rfcMismatch]);
      repository.correctAnswers.add(
        const ApiRejected(code: ErrorCode.rfcAlreadyRegistered),
      );

      await flow.confirmReview(rfc: 'GOMA800101AB7');

      expect(flow.step, OnboardingStep.review);
      expect(flow.rfcRejection, RfcRejection.duplicate);
    });

    test('a correction that cannot be sent shows a notice', () async {
      await createWith([_rfcMismatch]);
      repository.correctAnswers.add(const ApiUnreachable());

      await flow.confirmReview(rfc: 'GOMA800101AB7');

      expect(flow.step, OnboardingStep.review);
      expect(flow.notice, OnboardingNotice.unreachable);
    });
  });

  group('tax inputs (module_tax_engine on, optional)', () {
    setUp(() => flow = newFlow(taxEngineEnabled: true));

    Future<void> createMoral() async {
      fillIdentity(type: PersonType.moral, rfc: 'ABC680524P76');
      await flow.submitRegimes([_byCode(moralCatalog, '601')]);
      await flow.submitOptions('601', {'iva_fully_taxable': true});
    }

    test(
      'after create, regimes with tax inputs get the optional step 5',
      () async {
        await createMoral();

        expect(flow.step, OnboardingStep.taxInputs);
        expect(flow.stepNumber, 5);
        expect(flow.inputRegimes.map((r) => r.code), ['601']);
        expect(flow.fiscalYear, 2026, reason: 'Mexico City is still in 2026');
      },
    );

    test(
      'saving sends the values for the fiscal year, then finishes',
      () async {
        await createMoral();
        const values = [
          TaxInputValue(key: 'profit_coefficient', valueRate: '0.2'),
        ];

        await flow.saveTaxInputs(values);

        final save = repository.taxInputSaves.single;
        expect(save.regimeCode, '601');
        expect(save.fiscalYear, 2026);
        expect(save.values, values);
        expect(flow.step, OnboardingStep.done);
      },
    );

    test('nothing typed is a skip: no call', () async {
      await createMoral();

      await flow.saveTaxInputs(const []);

      expect(repository.taxInputSaves, isEmpty);
      expect(flow.step, OnboardingStep.done);
    });

    test('a failed save stays with a notice', () async {
      await createMoral();
      repository.taxInputsFailure = const ApiUnreachable();

      await flow.saveTaxInputs(const [TaxInputValue(key: 'k', valueRate: '1')]);

      expect(flow.step, OnboardingStep.taxInputs);
      expect(flow.notice, OnboardingNotice.unreachable);
    });

    test('with the tax engine off the step never shows', () async {
      flow = newFlow();
      await createMoral();

      expect(flow.inputRegimes, isEmpty);
      expect(flow.step, OnboardingStep.done);
    });
  });

  test('normalizeRfc is what the flow sends', () {
    expect(normalizeRfc('abc-680524 p76'), 'ABC680524P76');
  });
}
