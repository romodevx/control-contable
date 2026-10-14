import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/presentation/taxpayer_switcher_sheet.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The app bar title of a main screen: the active taxpayer's name, which
/// opens the switcher sheet. Shows [fallback] while `module_taxpayers` is
/// off (UX only: the server answers `404`) or no taxpayer is active.
class ActiveTaxpayerSwitcher extends ConsumerWidget {
  /// The switcher, or [fallback].
  const new({required this.fallback, super.key});

  /// The title shown when there is nothing to switch.
  final Widget fallback;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (!ref.useFlag(AppModule.taxpayers)) return fallback;
    final active = ref.watch(activeTaxpayerProvider).value;
    if (active == null) return fallback;
    void open() => unawaited(showTaxpayerSwitcher(context));
    // excludeSemantics drops the InkWell's tap action, so the merged node
    // declares it again (TalkBack/VoiceOver activate through it).
    return Semantics(
      button: true,
      label: AppLocalizations.of(context).switcherLabel(active.displayName),
      onTap: open,
      excludeSemantics: true,
      child: InkWell(
        borderRadius: BorderRadius.circular(AppSpacing.xs),
        onTap: open,
        child: ConstrainedBox(
          constraints: const BoxConstraints(
            minHeight: kMinInteractiveDimension,
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Flexible(
                child: Text(
                  active.displayName,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              const Icon(Icons.expand_more),
            ],
          ),
        ),
      ),
    );
  }
}
