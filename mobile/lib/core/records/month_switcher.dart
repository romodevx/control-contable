import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Horizontal swipe speed (logical px/s) that turns the month.
const double _swipeVelocity = 300;

/// The month a record list shows: back, the month's name, forward. Never
/// goes past [latest] (records cannot be in the future). A horizontal
/// swipe on [child] turns the month too.
class MonthSwitcher extends StatelessWidget {
  /// The switcher on [month].
  const new({
    required this.month,
    required this.latest,
    required this.onChanged,
    required this.child,
    super.key,
  });

  /// The month shown.
  final RecordMonth month;

  /// The last month the user may reach (the current one).
  final RecordMonth latest;

  /// Called with the month to show.
  final ValueChanged<RecordMonth> onChanged;

  /// The month's content; swiping it turns the month.
  final Widget child;

  bool get _canGoForward => month.isBefore(latest);

  void _handleSwipe(DragEndDetails details) {
    final velocity = details.primaryVelocity ?? 0;
    if (velocity > _swipeVelocity) onChanged(month.previous);
    if (velocity < -_swipeVelocity && _canGoForward) onChanged(month.next);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final label = MaterialLocalizations.of(context)
        .formatMonthYear(month.firstDay);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: AppSpacing.xs),
          child: Row(
            children: [
              IconButton(
                tooltip: l10n.recordsPreviousMonth,
                icon: const Icon(Icons.chevron_left),
                onPressed: () => onChanged(month.previous),
              ),
              Expanded(
                child: Semantics(
                  header: true,
                  child: Text(
                    label,
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ),
              ),
              IconButton(
                tooltip: l10n.recordsNextMonth,
                icon: const Icon(Icons.chevron_right),
                onPressed: _canGoForward ? () => onChanged(month.next) : null,
              ),
            ],
          ),
        ),
        Expanded(
          child: GestureDetector(
            onHorizontalDragEnd: _handleSwipe,
            child: child,
          ),
        ),
      ],
    );
  }
}
