import 'dart:async';

import 'package:control_contable/features/auth/application/sign_out.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// *Cerrar sesión*. Signing out is reversible (sign in again), so it acts
/// at once with no confirmation (house UI rules). Repeat taps while it runs
/// are ignored; the router then shows the login form.
class SignOutButton extends ConsumerStatefulWidget {
  /// Creates the button.
  const new({super.key});

  @override
  ConsumerState<SignOutButton> createState() => _SignOutButtonState();
}

class _SignOutButtonState extends ConsumerState<SignOutButton> {
  bool _isSigningOut = false;

  Future<void> _signOut() async {
    if (_isSigningOut) return;
    setState(() => _isSigningOut = true);
    try {
      await ref.read(signOutProvider).call();
    } finally {
      if (mounted) setState(() => _isSigningOut = false);
    }
  }

  @override
  Widget build(BuildContext context) => TextButton(
    onPressed: () => unawaited(_signOut()),
    child: Text(AppLocalizations.of(context).authSignOut),
  );
}
