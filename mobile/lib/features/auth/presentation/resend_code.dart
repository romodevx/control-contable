import 'package:control_contable/features/auth/application/resend_cooldown.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// "Enviar otro código", offered only once [cooldown] is over; until then
/// the remaining wait is shown as text (no disabled button to puzzle over).
/// The countdown is not a live region: a screen reader would otherwise
/// read it every second.
class ResendCode extends StatelessWidget {
  /// The resend action, gated by [cooldown].
  const new({required this.cooldown, required this.onResend, super.key});

  /// The wait before another code.
  final ResendCooldown cooldown;

  /// Asks for a new code.
  final VoidCallback onResend;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ListenableBuilder(
      listenable: cooldown,
      builder: (context, _) => Align(
        alignment: Alignment.centerLeft,
        child: cooldown.isRunning
            ? Text(
                l10n.authResendWait(formatCountdown(cooldown.remaining)),
                style: Theme.of(context).textTheme.bodyMedium,
              )
            : TextButton(onPressed: onResend, child: Text(l10n.authResend)),
      ),
    );
  }
}
