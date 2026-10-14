import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

const double _spinnerSize = 20;
const double _spinnerStroke = 2;

/// A one-line load status, announced as it changes: a spinner and
/// [message] while [busy], else [message] alone.
class LoadStatus extends StatelessWidget {
  /// A status reading [message].
  const new({required this.message, this.busy = false, super.key});

  /// What the user reads.
  final String message;

  /// Whether the load is still running.
  final bool busy;

  @override
  Widget build(BuildContext context) => Semantics(
    liveRegion: true,
    child: Row(
      children: [
        if (busy) ...[
          const SizedBox.square(
            dimension: _spinnerSize,
            child: ExcludeSemantics(
              child: CircularProgressIndicator(strokeWidth: _spinnerStroke),
            ),
          ),
          const SizedBox(width: AppSpacing.xs),
        ],
        Expanded(
          child: Text(message, style: Theme.of(context).textTheme.bodyMedium),
        ),
      ],
    ),
  );
}

/// A failed load: [message] and a *Reintentar* button running [onRetry].
class LoadFailed extends StatelessWidget {
  /// The failure reading [message].
  const new({required this.message, required this.onRetry, super.key});

  /// What failed.
  final String message;

  /// Loads again.
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) => Row(
    children: [
      Expanded(child: LoadStatus(message: message)),
      TextButton(
        onPressed: onRetry,
        child: Text(AppLocalizations.of(context).authRetry),
      ),
    ],
  );
}
