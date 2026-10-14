import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Warns that the last sign-out could not remove the session token from
/// this device (`Session.tokenLeftOnDevice`), so it never stays there
/// silently. Takes no space otherwise.
class SessionWipeNotice extends ConsumerWidget {
  /// Creates the notice, surrounded by [padding] when it shows.
  const new({this.padding = EdgeInsets.zero, super.key});

  /// Space around the notice when it shows.
  final EdgeInsetsGeometry padding;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(sessionProvider);
    return ListenableBuilder(
      listenable: session,
      builder: (context, _) => session.tokenLeftOnDevice
          ? Padding(
              padding: padding,
              child: FormNotice(
                message: AppLocalizations.of(context).sessionTokenLeftOnDevice,
              ),
            )
          : const SizedBox.shrink(),
    );
  }
}
