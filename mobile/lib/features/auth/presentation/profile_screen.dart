import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/auth/application/profile.dart';
import 'package:control_contable/features/auth/presentation/profile_form.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// *Mi cuenta* (US-A3, US-A4): the profile (name editable, email
/// read-only), the legal documents and the account deletion.
class ProfileScreen extends ConsumerWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final profile = ref.watch(profileProvider);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.accountTitle)),
      body: switch (profile) {
        AsyncData(:final value) => ProfileForm(user: value.profile),
        AsyncError() => _Padded(
          LoadFailed(
            message: l10n.accountLoadFailed,
            onRetry: () => ref.invalidate(profileProvider),
          ),
        ),
        _ => _Padded(LoadStatus(message: l10n.accountLoading, busy: true)),
      },
    );
  }
}

class _Padded extends StatelessWidget {
  const new(this.child);

  final Widget child;

  @override
  Widget build(BuildContext context) =>
      Padding(padding: const EdgeInsets.all(AppSpacing.gutter), child: child);
}
