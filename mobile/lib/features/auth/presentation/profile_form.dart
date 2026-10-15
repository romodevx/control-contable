import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/user.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/auth/data/profile_repository.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/features/auth/presentation/delete_account_section.dart';
import 'package:control_contable/features/auth/presentation/legal_documents_section.dart';
import 'package:control_contable/features/auth/presentation/read_only_value.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// `UserUpdate.name` length limit.
const int _nameMaxLength = 120;

/// The account screen's content for [user]: the name (the one editable
/// field), the email (read-only), the legal documents and the deletion.
class ProfileForm extends ConsumerStatefulWidget {
  /// The form prefilled from [user].
  const new({required this.user, super.key});

  /// The profile as loaded.
  final User user;

  @override
  ConsumerState<ProfileForm> createState() => _ProfileFormState();
}

class _ProfileFormState extends ConsumerState<ProfileForm> {
  final AppFormController _form = AppFormController();
  AppFieldController? _nameField;
  ({String message, FormNoticeTone tone})? _notice;

  AppFieldController get _name => _nameField!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // The label and validator need the localized copy (not in initState).
    final l10n = AppLocalizations.of(context);
    _nameField ??= _form.textField(
      label: l10n.authNameLabel,
      initialText: widget.user.name ?? '',
      validator: (value) => value.runes.length > _nameMaxLength
          ? CustomFieldError(l10n.authNameTooLong)
          : null,
    );
  }

  @override
  void dispose() {
    _form.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final l10n = AppLocalizations.of(context);
    _showNotice(null);
    try {
      await ref.read(profileRepositoryProvider).updateName(_name.value);
      _showNotice((message: l10n.accountSaved, tone: FormNoticeTone.info));
    } on ApiValidationFailure catch (failure) {
      _showServerVerdict(failure.details, l10n);
    } on ApiFailure catch (failure) {
      _showNotice((
        message: describeAuthFailure(failure, l10n),
        tone: FormNoticeTone.error,
      ));
    }
  }

  void _showServerVerdict(List<api.FieldError> details, AppLocalizations l10n) {
    final (:matched, :unmatched) = matchServerFieldErrors(
      details,
      fieldsByPath: {'name': _name},
      describe: (detail) => describeProfileNameDetail(detail, l10n),
    );
    if (!mounted) return;
    _form.showServerErrors(matched);
    if (unmatched.isNotEmpty || matched.isEmpty) {
      _showNotice((message: l10n.authFormRejected, tone: FormNoticeTone.error));
    }
  }

  void _showNotice(({String message, FormNoticeTone tone})? notice) {
    if (mounted) setState(() => _notice = notice);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final notice = _notice;
    return AppForm(
      controller: _form,
      onSave: _save,
      primaryLabel: l10n.accountSave,
      children: [
        if (notice != null)
          FormNotice(message: notice.message, tone: notice.tone),
        AppTextField(
          field: _name,
          keyboardType: TextInputType.name,
          textInputAction: TextInputAction.done,
          autofillHints: const [AutofillHints.name],
        ),
        ReadOnlyValue(
          label: l10n.authEmailLabel,
          value: widget.user.email,
          note: l10n.accountEmailReadOnly,
        ),
        const LegalDocumentsSection(),
        const DeleteAccountSection(),
      ],
    );
  }
}
