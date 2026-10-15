import 'dart:async';

import 'package:control_contable/api/generated/models/withholding_preview.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter/foundation.dart';

/// How long the facts must stay still before the preview is asked: one
/// request per pause in typing, not one per keystroke.
const Duration withholdingPreviewDebounce = Duration(milliseconds: 400);

/// What the derived card can show.
enum WithholdingPreviewStatus {
  /// The facts the derivation needs are not all there yet.
  incomplete,

  /// The figures answer the current facts (or the next answer is on its
  /// way: the last figures stay meanwhile, so the card never flickers).
  current,

  /// The last request failed (offline, or the server refused the facts):
  /// the server derives the withholding anyway when the record is saved.
  failed,
}

/// Asks the server to derive a withholding (ADR-0008: the client never
/// computes one) whenever the facts of a form change, debounced by
/// [withholdingPreviewDebounce]. `query` reads the facts from the form, or
/// `null` while they are incomplete; an unchanged query asks nothing. Only
/// the latest answer lands, and a failure keeps the last figures.
class WithholdingPreviewLoader<Q extends Object> extends ChangeNotifier {
  /// Listens to `source` and asks `fetch` for each new `query`; `initial`
  /// is the stored derivation of the record being edited, which stands for
  /// the facts the form starts with.
  new({
    required this._source,
    required this._query,
    required this._fetch,
    WithholdingPreview? initial,
  }) : _preview = initial {
    _lastQuery = _query();
    _status = _lastQuery == null
        ? WithholdingPreviewStatus.incomplete
        : WithholdingPreviewStatus.current;
    if (_lastQuery != null && initial == null) _schedule();
    _source.addListener(_check);
  }

  final Listenable _source;
  final Q? Function() _query;
  final Future<WithholdingPreview> Function(Q query) _fetch;
  Q? _lastQuery;
  WithholdingPreview? _preview;
  late WithholdingPreviewStatus _status;
  Timer? _debounce;
  int _generation = 0;
  bool _isDisposed = false;

  /// The latest answer, kept while a newer one is on its way or failed;
  /// `null` before the first one.
  WithholdingPreview? get preview =>
      _status == WithholdingPreviewStatus.incomplete ? null : _preview;

  /// What the card can say.
  WithholdingPreviewStatus get status => _status;

  void _check() {
    final query = _query();
    if (query == _lastQuery) return;
    _lastQuery = query;
    _debounce?.cancel();
    if (query == null) {
      _generation++;
      _setStatus(WithholdingPreviewStatus.incomplete);
      return;
    }
    _schedule();
  }

  void _schedule() {
    _debounce?.cancel();
    _debounce = Timer(withholdingPreviewDebounce, () {
      final query = _lastQuery;
      if (query != null) unawaited(_load(query, ++_generation));
    });
  }

  Future<void> _load(Q query, int generation) async {
    WithholdingPreview? answer;
    try {
      answer = await _fetch(query);
    } on ApiFailure {
      answer = null;
    }
    if (_isDisposed || generation != _generation) return;
    if (answer != null) _preview = answer;
    _setStatus(
      answer == null
          ? WithholdingPreviewStatus.failed
          : WithholdingPreviewStatus.current,
      force: true,
    );
  }

  void _setStatus(WithholdingPreviewStatus status, {bool force = false}) {
    if (status == _status && !force) return;
    _status = status;
    notifyListeners();
  }

  @override
  void dispose() {
    _isDisposed = true;
    _debounce?.cancel();
    _source.removeListener(_check);
    super.dispose();
  }
}
