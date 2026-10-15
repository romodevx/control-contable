import 'package:control_contable/api/generated/models/reminder_settings.dart';
import 'package:control_contable/core/forms/form_footer.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/reminders/application/reminder_settings.dart';
import 'package:control_contable/features/reminders/presentation/reminder_detail_fields.dart';
import 'package:control_contable/features/reminders/presentation/show_amounts_field.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// `details[].code` of the `422` for "on with nothing to send".
const String _noReminderSelected = 'no_reminder_selected';

/// The reminder details while reminders are on, saved together with
/// *Guardar cambios* (never disabled; repeat taps ignored while saving).
/// "Nothing selected" is checked on submit only.
class ReminderDetailsForm extends ConsumerStatefulWidget {
  /// The form over the [saved] settings.
  const new({required this.saved, super.key});

  /// What the server holds.
  final ReminderSettings saved;

  @override
  ConsumerState<ReminderDetailsForm> createState() =>
      _ReminderDetailsFormState();
}

class _ReminderDetailsFormState extends ConsumerState<ReminderDetailsForm> {
  late ReminderSettings _draft = widget.saved;
  final FocusNode _noticeFocus = FocusNode();
  bool _saving = false;
  String? _error;
  bool _saved = false;

  @override
  void dispose() {
    _noticeFocus.dispose();
    super.dispose();
  }

  void _edit(ReminderSettings next) => setState(() {
    _draft = next;
    _saved = false;
    // Once shown, the error clears the moment the draft is valid.
    if (_error != null && _hasAReminder(next)) _error = null;
  });

  static bool _hasAReminder(ReminderSettings s) =>
      s.earlyDay != null || s.onDueDate;

  Future<void> _save() async {
    if (_saving) return;
    final l10n = AppLocalizations.of(context);
    if (!_hasAReminder(_draft)) return _fail(l10n.remindersNoneSelected);
    setState(() => _saving = true);
    try {
      await ref.read(reminderSettingsProvider.notifier).save(_draft);
      if (!mounted) return;
      setState(() {
        _saved = true;
        _error = null;
      });
    } on ApiRejected catch (rejected) {
      final noneSelected = rejected.details.any(
        (detail) => detail.code == _noReminderSelected,
      );
      _fail(
        noneSelected ? l10n.remindersNoneSelected : l10n.remindersSaveFailed,
      );
    } on Exception {
      _fail(l10n.remindersSaveFailed);
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _fail(String message) {
    if (!mounted) return;
    setState(() {
      _error = message;
      _saved = false;
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_noticeFocus.context != null) _noticeFocus.requestFocus();
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    const padding = EdgeInsets.symmetric(horizontal: AppSpacing.gutter);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (_error case final error?)
          Padding(
            padding: padding,
            child: Focus(
              focusNode: _noticeFocus,
              child: FormNotice(message: error),
            ),
          ),
        EarlyDayField(
          value: _draft.earlyDay,
          onChanged: (day) => _edit(_draft.copyWith(earlyDay: day)),
        ),
        SwitchListTile(
          title: Text(l10n.remindersOnDueDateLabel),
          value: _draft.onDueDate,
          onChanged: (on) => _edit(_draft.copyWith(onDueDate: on)),
        ),
        SendTimeField(
          value: _draft.sendTime,
          onChanged: (time) => _edit(_draft.copyWith(sendTime: time)),
        ),
        ShowAmountsField(
          value: _draft.showAmounts,
          onChanged: (on) => _edit(_draft.copyWith(showAmounts: on)),
        ),
        const SizedBox(height: AppSpacing.md),
        if (_saved)
          Padding(
            padding: padding.copyWith(bottom: AppSpacing.xs),
            child: FormNotice(
              message: l10n.remindersSaved,
              tone: FormNoticeTone.info,
            ),
          ),
        Padding(
          padding: padding,
          child: FilledButton(
            onPressed: _save,
            child: _saving
                ? BusyLabel(
                    label: l10n.remindersSave,
                    announcement: l10n.formSaving,
                  )
                : Text(l10n.remindersSave),
          ),
        ),
      ],
    );
  }
}
