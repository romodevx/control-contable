import 'dart:async';

import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/api/generated/models/annual_section_kind.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// The computed kind of [kind], or `null` for a section with no trace (an
/// explanation, or a kind this app version does not know).
AnnualComputedSectionKind? computedKindOf(AnnualSectionKind kind) {
  final json = kind.json;
  if (json == null) return null;
  final computed = AnnualComputedSectionKind.fromJson(json);
  return computed == AnnualComputedSectionKind.$unknown ? null : computed;
}

/// "Ver cómo se calculó": opens [section]'s breakdown. Nothing for a
/// section with no trace.
class AnnualBreakdownButton extends StatelessWidget {
  /// The link to [section]'s breakdown in [taxpayerId]'s [fiscalYear].
  const new({
    required this.section,
    required this.taxpayerId,
    required this.fiscalYear,
    this.color,
    super.key,
  });

  /// The section whose trace opens.
  final AnnualSectionKind section;

  /// Whose estimate it is.
  final String taxpayerId;

  /// The fiscal year estimated.
  final int fiscalYear;

  /// The label's colour (the theme's by default).
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final computed = computedKindOf(section);
    if (computed == null) return const SizedBox.shrink();
    return Align(
      alignment: AlignmentDirectional.centerEnd,
      child: TextButton(
        style: TextButton.styleFrom(foregroundColor: color),
        onPressed: () => unawaited(
          context.push<void>(
            AnnualRoutes.breakdown(taxpayerId, fiscalYear, computed),
          ),
        ),
        child: Text(AppLocalizations.of(context).annualViewBreakdown),
      ),
    );
  }
}
