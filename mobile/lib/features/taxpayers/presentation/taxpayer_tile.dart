import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/taxpayers/domain/masked_rfc.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Characters shown before the hidden part of a masked RFC, by length.
const int _fisicaPrefix = 4;
const int _moralPrefix = 3;

/// Characters shown after it (homoclave and check digit).
const int _visibleSuffix = 3;

/// How a taxpayer reads in a list: its name, the masked RFC
/// (`GOMA••••••AB6`) and its person type as a text badge (and "Archivado"
/// when archived), never by color alone.
class TaxpayerSummary extends StatelessWidget {
  /// The summary of [taxpayer].
  const new({required this.taxpayer, super.key});

  /// The taxpayer.
  final Taxpayer taxpayer;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final masked = taxpayer.rfcMasked;
    final prefix = taxpayer.personType == PersonType.moral
        ? _moralPrefix
        : _fisicaPrefix;
    final readable = masked.length > prefix + _visibleSuffix
        ? l10n.taxpayerRfcSemantics(
            masked.substring(0, prefix),
            masked.substring(masked.length - _visibleSuffix),
          )
        : null;
    return Wrap(
      spacing: AppSpacing.xs,
      runSpacing: AppSpacing.xxs,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        Text(displayRfc(masked), semanticsLabel: readable),
        TextBadge(label: personTypeLabel(taxpayer.personType, l10n)),
        if (taxpayer.archivedAt != null)
          TextBadge(label: l10n.taxpayerArchivedBadge),
      ],
    );
  }
}

/// "Persona física" or "Persona moral".
String personTypeLabel(PersonType type, AppLocalizations l10n) =>
    type == PersonType.moral
    ? l10n.onboardingMoralTitle
    : l10n.onboardingFisicaTitle;

/// A small outlined label. The words carry the meaning.
class TextBadge extends StatelessWidget {
  /// A badge reading [label].
  const new({required this.label, super.key});

  /// What it says.
  final String label;

  @override
  Widget build(BuildContext context) {
    final colors = context.tokens.colors;
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: colors.border),
        borderRadius: BorderRadius.circular(AppSpacing.xxs),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.xs,
          vertical: AppSpacing.xxs / 2,
        ),
        child: Text(label, style: Theme.of(context).textTheme.bodySmall),
      ),
    );
  }
}
