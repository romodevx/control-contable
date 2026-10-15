import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/api/generated/models/withholder.dart';
import 'package:control_contable/api/generated/models/withholding.dart';
import 'package:control_contable/api/generated/models/withholding_basis.dart';
import 'package:control_contable/api/generated/models/withholding_explanation.dart';
import 'package:control_contable/api/generated/models/withholding_preview.dart';
import 'package:control_contable/core/fiscal/withholding_preview_loader.dart';
import 'package:flutter_test/flutter_test.dart';

/// A derived withholding as the server answers it (ADR-0008).
Withholding withholding({
  required int netCents,
  int isrCents = 0,
  int ivaCents = 0,
  String isrRate = '0',
  WithholdingBasis ivaBasis = WithholdingBasis.none,
  Withholder withholder = Withholder.none,
  String? remitDueOn,
  Map<String, String> explanations = const {},
  bool cpaUnverified = false,
}) => Withholding(
  isrCents: isrCents,
  ivaCents: ivaCents,
  isrRate: isrRate,
  ivaBasis: ivaBasis,
  withholder: withholder,
  netCents: netCents,
  remitDueOn: remitDueOn,
  ruleIds: explanations.keys.toList(),
  explanations: [
    for (final MapEntry(:key, :value) in explanations.entries)
      WithholdingExplanation(ruleId: key, text: value),
  ],
  tableVersion: '2026.1',
  cpaUnverified: cpaUnverified,
);

/// A preview answer carrying [withholding] and [warnings].
WithholdingPreview preview(
  Withholding withholding, {
  List<WarningCode> warnings = const [],
}) => WithholdingPreview(
  withholding: withholding,
  warnings: [
    for (final code in warnings)
      Warning(code: code, field: 'withholding', message: code.json!),
  ],
);

/// A preview where nobody withholds anything: the net is the total.
WithholdingPreview nothingWithheld(int totalCents) => preview(
  withholding(
    netCents: totalCents,
    explanations: const {'ISR-0': 'Una persona física no te retiene ISR.'},
  ),
);

/// The owner's case (US-W1, WM WH-01): a RESICO persona física bills
/// 100,000.00 + IVA of professional services to a company.
final WithholdingPreview ownerExample = preview(
  withholding(
    isrCents: 125000,
    ivaCents: 1066667,
    isrRate: '0.0125',
    ivaBasis: WithholdingBasis.twoThirdsOfIva,
    withholder: Withholder.counterparty,
    netCents: 10408333,
    explanations: const {
      'ISR-4':
          'La empresa te retiene 1.25 % de ISR porque estás en RESICO '
          '(LISR art. 113-J).',
      'IVA-6':
          'Y 2/3 del IVA por servicios de persona física (LIVA art. 1-A, '
          'RLIVA art. 3).',
    },
  ),
);

/// US-W6 (WM WH-11): a 601 pays a 612 consultant 40,000.00 + IVA.
final WithholdingPreview moralPaysConsultant = preview(
  withholding(
    isrCents: 400000,
    ivaCents: 426667,
    isrRate: '0.10',
    ivaBasis: WithholdingBasis.twoThirdsOfIva,
    withholder: Withholder.you,
    netCents: 3813333,
    remitDueOn: '2026-07-17',
    explanations: const {
      'ISR-5':
          'Debes retener 10 % de ISR por honorarios de persona física '
          '(LISR art. 106).',
      'IVA-6':
          'Y 2/3 del IVA por servicios de persona física (LIVA art. 1-A, '
          'RLIVA art. 3).',
    },
  ),
  warnings: const [WarningCode.deductionRequiresRemittance],
);

/// Lets the debounced preview fire and its answer land.
Future<void> settlePreview(WidgetTester tester) async {
  await tester.pump(withholdingPreviewDebounce);
  await tester.pumpAndSettle();
}
