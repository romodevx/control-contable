import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The days an early reminder can fall on (`early_day` 1–16).
const int _firstEarlyDay = 1;
const int _lastEarlyDay = 16;
const int _timeDigits = 2;
const String _timeSeparator = ':';

/// A field with its visible label above it, read together by screen
/// readers.
class LabeledField extends StatelessWidget {
  /// [child] under [label].
  const new({required this.label, required this.child, super.key});

  /// The visible label.
  final String label;

  /// The control.
  final Widget child;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(
      horizontal: AppSpacing.gutter,
      vertical: AppSpacing.xs,
    ),
    child: MergeSemantics(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(label, style: Theme.of(context).textTheme.titleSmall),
          const SizedBox(height: AppSpacing.xxs),
          child,
        ],
      ),
    ),
  );
}

/// "Recordatorio anticipado": a day 1–16 of the due month, or none.
class EarlyDayField extends StatelessWidget {
  /// The select showing [value] (`null`: none).
  const new({required this.value, required this.onChanged, super.key});

  /// The chosen day, or `null`.
  final int? value;

  /// The user chose a day (or none).
  final ValueChanged<int?> onChanged;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return LabeledField(
      label: l10n.remindersEarlyDayLabel,
      child: DropdownButtonFormField<int?>(
        initialValue: value,
        isExpanded: true,
        onChanged: onChanged,
        items: [
          DropdownMenuItem(child: Text(l10n.remindersEarlyDayNone)),
          for (var day = _firstEarlyDay; day <= _lastEarlyDay; day++)
            DropdownMenuItem(
              value: day,
              child: Text(l10n.remindersEarlyDayOption(day)),
            ),
        ],
      ),
    );
  }
}

/// "Hora del aviso": `HH:MM` in Mexico City, chosen with the platform
/// time picker.
class SendTimeField extends StatelessWidget {
  /// The field showing [value] (`HH:MM`).
  const new({required this.value, required this.onChanged, super.key});

  /// The send time, `HH:MM`.
  final String value;

  /// The user picked another time (`HH:MM`).
  final ValueChanged<String> onChanged;

  Future<void> _pick(BuildContext context) async {
    final picked = await showTimePicker(
      context: context,
      initialTime: parseSendTime(value),
    );
    if (picked != null) onChanged(formatSendTime(picked));
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return LabeledField(
      label: l10n.remindersSendTimeLabel,
      child: OutlinedButton(
        style: OutlinedButton.styleFrom(alignment: Alignment.centerLeft),
        onPressed: () => unawaited(_pick(context)),
        child: Text(l10n.remindersSendTimeValue(value)),
      ),
    );
  }
}

/// The time of a `HH:MM` send time.
TimeOfDay parseSendTime(String value) {
  final [hour, minute] = value.split(_timeSeparator).map(int.parse).toList();
  return TimeOfDay(hour: hour, minute: minute);
}

/// [time] as the contract's `HH:MM`.
String formatSendTime(TimeOfDay time) =>
    '${time.hour.toString().padLeft(_timeDigits, '0')}$_timeSeparator'
    '${time.minute.toString().padLeft(_timeDigits, '0')}';
