import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:flutter/material.dart';

/// The end of a paged record list: once built (it scrolled into view) it
/// asks for the next page, shows it loading, or offers a retry. Nothing
/// when every page is loaded.
class RecordPageFooter extends StatelessWidget {
  /// The footer of a list that [hasMore] pages.
  const new({
    required this.hasMore,
    required this.isLoading,
    required this.failed,
    required this.onLoadMore,
    required this.loadingMessage,
    required this.failedMessage,
    super.key,
  });

  /// Whether another page exists.
  final bool hasMore;

  /// Whether the next page is loading.
  final bool isLoading;

  /// Whether the last attempt failed (no automatic retry then).
  final bool failed;

  /// Loads the next page.
  final Future<void> Function() onLoadMore;

  /// Read while loading.
  final String loadingMessage;

  /// Read after a failure, before *Reintentar*.
  final String failedMessage;

  @override
  Widget build(BuildContext context) {
    if (!hasMore) return const SizedBox.shrink();
    if (!failed && !isLoading) {
      WidgetsBinding.instance.addPostFrameCallback(
        (_) => unawaited(onLoadMore()),
      );
    }
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: failed
          ? LoadFailed(
              message: failedMessage,
              onRetry: () => unawaited(onLoadMore()),
            )
          : LoadStatus(message: loadingMessage, busy: true),
    );
  }
}
