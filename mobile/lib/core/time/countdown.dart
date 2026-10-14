import 'dart:async';

import 'package:control_contable/core/time/clock.dart';
import 'package:flutter/foundation.dart';

const Duration _tick = Duration(seconds: 1);

/// A visible wait (a cooldown, a `Retry-After`). Measured against a
/// deadline on the [Clock], so time spent in the background still counts;
/// the timer only refreshes what is shown, once per second.
class Countdown extends ChangeNotifier {
  /// An idle countdown on [clock].
  new(Clock clock) : _now = clock;

  final Clock _now;
  DateTime? _deadline;
  Timer? _ticker;
  int _shownSeconds = 0;

  /// How long is left; zero when idle.
  Duration get remaining {
    final deadline = _deadline;
    if (deadline == null) return Duration.zero;
    final left = deadline.difference(_now());
    return left > Duration.zero ? left : Duration.zero;
  }

  /// Whether the wait is still on.
  bool get isRunning => remaining > Duration.zero;

  /// Starts (or restarts) a wait of [wait].
  void run(Duration wait) {
    _deadline = _now().add(wait);
    _ticker?.cancel();
    _ticker = Timer.periodic(_tick, (_) => _refresh());
    _refresh();
  }

  void _refresh() {
    final seconds = remaining.inSeconds;
    if (seconds == 0) _ticker?.cancel();
    if (seconds == _shownSeconds) return;
    _shownSeconds = seconds;
    notifyListeners();
  }

  @override
  void dispose() {
    _ticker?.cancel();
    super.dispose();
  }
}
