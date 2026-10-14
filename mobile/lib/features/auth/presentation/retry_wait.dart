import 'dart:async';

import 'package:control_contable/core/time/countdown.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

/// The live `m:ss` left before another try after a `429` (`Retry-After`).
/// Not a live region: it changes every second; the form notice is announced
/// once, and a tap during the wait announces what is left
/// ([announceRetryWait]).
class RetryWait extends StatelessWidget {
  /// Shows [countdown] while it runs; nothing otherwise.
  const new({required this.countdown, super.key});

  /// The `Retry-After` wait being counted down.
  final Countdown countdown;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: countdown,
    builder: (context, _) => countdown.isRunning
        ? Text(
            retryWaitText(context, countdown),
            style: Theme.of(context).textTheme.bodyMedium,
          )
        : const SizedBox.shrink(),
  );
}

/// "Podrás intentar de nuevo en m:ss." for what is left of [countdown].
String retryWaitText(BuildContext context, Countdown countdown) =>
    AppLocalizations.of(context)
        .authRetryIn(formatCountdown(countdown.remaining));

/// Announces, assertively, how long is left of [countdown]: a tap during
/// the wait sends nothing, so the screen reader must hear why.
void announceRetryWait(BuildContext context, Countdown countdown) => unawaited(
  SemanticsService.sendAnnouncement(
    View.of(context),
    retryWaitText(context, countdown),
    Directionality.of(context),
    assertiveness: Assertiveness.assertive,
  ),
);
