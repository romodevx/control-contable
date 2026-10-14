import 'dart:async';

import 'package:flutter/material.dart';

/// The Undo toast of a record list (house rule: do it, then offer Undo).
///
/// - It never auto-dismisses: it stays until the user taps *Deshacer* or
///   closes it (a `SnackBar` is a polite live region, `role=status`).
/// - Deletes stack: a delete while the toast is up joins it ("2 ingresos
///   eliminados"), and *Deshacer* undoes them all.
///
/// Lives in a provider, not a screen: the toast outlives the list.
class StackedUndo<T> {
  final List<T> _pending = [];
  int _generation = 0;

  /// What the toast would undo now.
  List<T> get pending => List.unmodifiable(_pending);

  /// Adds [item] and shows the toast with the whole stack on [messenger].
  /// [message] reads the stack's size; [onUndo] gets the stack, which is
  /// then cleared. Closing the toast forgets the stack (nothing undone).
  void push(
    T item, {
    required ScaffoldMessengerState messenger,
    required String Function(int count) message,
    required String undoLabel,
    required void Function(List<T> items) onUndo,
  }) {
    _pending.add(item);
    final generation = ++_generation;
    final stack = List<T>.unmodifiable(_pending);
    messenger.hideCurrentSnackBar();
    final toast = messenger.showSnackBar(
      SnackBar(
        content: Text(message(stack.length)),
        persist: true,
        showCloseIcon: true,
        action: SnackBarAction(
          label: undoLabel,
          onPressed: () {
            _forget(generation);
            onUndo(stack);
          },
        ),
      ),
    );
    // Replaced by a newer stack: that one owns the items now.
    unawaited(toast.closed.then((_) => _forget(generation)));
  }

  void _forget(int generation) {
    if (generation == _generation) _pending.clear();
  }
}
