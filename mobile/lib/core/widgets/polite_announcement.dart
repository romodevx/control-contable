import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

/// Politely announces each of [messages] not in [before], after the frame
/// that shows them: a live region alone is silent on VoiceOver, and these
/// texts appear without moving focus.
void announceNewPolitely(
  State state,
  List<String> messages, {
  List<String> before = const [],
}) {
  final fresh = messages.where((m) => !before.contains(m)).toList();
  if (fresh.isEmpty) return;
  WidgetsBinding.instance.addPostFrameCallback((_) {
    if (!state.mounted) return;
    final context = state.context;
    for (final message in fresh) {
      unawaited(
        SemanticsService.sendAnnouncement(
          View.of(context),
          message,
          Directionality.of(context),
        ),
      );
    }
  });
}
