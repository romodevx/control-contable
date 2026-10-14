import 'package:control_contable/core/forms/form_scope.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:flutter/material.dart';

/// Confirms a saved record on [messenger], the same way for every record
/// form. [heading] alone without [notes]; with notes (the server's
/// non-blocking warnings, already in words) the toast reads
/// "[heading] [notes]" and stays until closed so they are not missed.
/// It waits behind a pending Undo of [undo], never closing it. [overForm]:
/// the form stays open (*Guardar y crear otro*), so the toast floats above
/// its footer.
void showRecordSavedToast(
  ScaffoldMessengerState messenger, {
  required String heading,
  required List<String> notes,
  required StackedUndo<Object?> undo,
  bool overForm = false,
}) {
  final hasNotes = notes.isNotEmpty;
  undo.show(
    messenger,
    SnackBar(
      content: Text(hasNotes ? '$heading ${notes.join(' ')}' : heading),
      persist: hasNotes,
      showCloseIcon: hasNotes,
      behavior: overForm ? SnackBarBehavior.floating : null,
      margin: overForm ? toastClearOfFormFooter : null,
    ),
  );
}
