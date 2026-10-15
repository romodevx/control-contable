import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Room under the last row so an extended floating button never covers it.
const double _fabClearance = 96;

/// The body of an annual input screen: [value]'s loading state, its failure
/// with a retry, or [data] once it arrives.
Widget annualInputBody<T>(
  AsyncValue<T> value, {
  required String loading,
  required String failed,
  required VoidCallback onRetry,
  required Widget Function(T data) data,
}) {
  // A value shows as soon as it exists, also while it refreshes.
  if (value.value case final T current?) return data(current);
  return Padding(
    padding: const EdgeInsets.all(AppSpacing.gutter),
    child: value.hasError && !value.isLoading
        ? LoadFailed(message: failed, onRetry: onRetry)
        : LoadStatus(message: loading, busy: true),
  );
}

/// The scrolling column of an annual input screen: the read-only notice
/// first when the session knows it, then [children] spaced evenly, with
/// room for a floating button under the last one.
class AnnualInputList extends ConsumerWidget {
  /// The list of [children].
  const new({
    required this.children,
    this.hasFloatingAction = false,
    super.key,
  });

  /// The screen's sections, in reading order.
  final List<Widget> children;

  /// Whether a floating button sits over the list's end.
  final bool hasFloatingAction;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final readOnly = ref.watch(readOnlyAccessProvider);
    final l10n = AppLocalizations.of(context);
    final items = [
      if (readOnly)
        Text(
          l10n.annualInputsReadOnlyNotice,
          key: const ValueKey('annual-input-read-only'),
          style: Theme.of(context).textTheme.bodyMedium,
        ),
      ...children,
    ];
    return ListView.separated(
      padding: EdgeInsets.fromLTRB(
        AppSpacing.gutter,
        AppSpacing.gutter,
        AppSpacing.gutter,
        AppSpacing.gutter + (hasFloatingAction ? _fabClearance : 0),
      ),
      itemCount: items.length,
      separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.lg),
      itemBuilder: (_, index) => items[index],
    );
  }
}
