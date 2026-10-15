import 'package:control_contable/api/generated/models/reminder_settings.dart';
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

/// The reminder details while reminders are on. Each change saves on its
/// own (MODULE-reminders §7: autosave, no primary button): the control
/// moves at once, a polite "Guardado" confirms every save, and a failure
/// puts the control back and says so assertively.
class ReminderDetailsForm extends ConsumerStatefulWidget {
  /// The details showing [settings] (the latest change included).
  const new({required this.settings, super.key});

  /// The settings on screen.
  final ReminderSettings settings;

  @override
  ConsumerState<ReminderDetailsForm> createState() =>
      _ReminderDetailsFormState();
}

/// The last word on the autosave, shown under the details.
sealed class _Status {
  const new();
}

final class _Saved extends _Status {
  const new(this.count);

  /// Which save this confirms: a new key per save re-announces it.
  final int count;
}

final class _Failed extends _Status {
  const new(this.message);

  final String message;
}

class _ReminderDetailsFormState extends ConsumerState<ReminderDetailsForm> {
  _Status? _status;
  int _saves = 0;

  Future<void> _change(ReminderSettings next) async {
    final l10n = AppLocalizations.of(context);
    try {
      final outcome = await ref
          .read(reminderSettingsProvider.notifier)
          .change(next);
      switch (outcome) {
        case ChangeOutcome.saved:
          _show(_Saved(++_saves));
        case ChangeOutcome.noReminderSelected:
          _show(_Failed(l10n.remindersNoneSelected));
        case ChangeOutcome.superseded:
          break;
      }
    } on ApiRejected catch (rejected) {
      final noneSelected = rejected.details.any(
        (detail) => detail.code == _noReminderSelected,
      );
      _show(
        _Failed(
          noneSelected ? l10n.remindersNoneSelected : l10n.remindersSaveFailed,
        ),
      );
    } on Exception {
      _show(_Failed(l10n.remindersSaveFailed));
    }
  }

  void _show(_Status status) {
    if (mounted) setState(() => _status = status);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final settings = widget.settings;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        EarlyDayField(
          value: settings.earlyDay,
          onChanged: (day) => _change(settings.copyWith(earlyDay: day)),
        ),
        SwitchListTile(
          title: Text(l10n.remindersOnDueDateLabel),
          value: settings.onDueDate,
          onChanged: (on) => _change(settings.copyWith(onDueDate: on)),
        ),
        SendTimeField(
          value: settings.sendTime,
          onChanged: (time) => _change(settings.copyWith(sendTime: time)),
        ),
        ShowAmountsField(
          value: settings.showAmounts,
          onChanged: (on) => _change(settings.copyWith(showAmounts: on)),
        ),
        if (_status case final status?)
          Padding(
            padding: const EdgeInsets.fromLTRB(
              AppSpacing.gutter,
              AppSpacing.md,
              AppSpacing.gutter,
              0,
            ),
            child: switch (status) {
              _Saved(:final count) => FormNotice(
                key: ValueKey(count),
                message: l10n.remindersSaved,
                tone: FormNoticeTone.info,
              ),
              _Failed(:final message) => FormNotice(message: message),
            },
          ),
      ],
    );
  }
}
