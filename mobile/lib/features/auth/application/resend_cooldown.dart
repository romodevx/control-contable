import 'dart:async';

import 'package:control_contable/core/time/clock.dart';
import 'package:flutter/foundation.dart';

/// The server's silent resend cooldown (US-A1c): the app waits it out
/// visibly instead of sending a request that does nothing.
const Duration resendCooldown = Duration(seconds: 60);

const Duration _tick = Duration(seconds: 1);

/// The countdown before another code can be asked for. Measured against a
/// deadline on the [Clock], so time spent in the background still counts;
/// the timer only refreshes what is shown, once per second.
class ResendCooldown extends ChangeNotifier {
  /// An idle countdown on the clock [_now].
  new(this._now);

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

  /// Whether another code must still wait.
  bool get isRunning => remaining > Duration.zero;

  /// Starts the wait: [serverWait] (a `Retry-After`) when longer than the
  /// 60 s cooldown, the cooldown otherwise.
  void start([Duration? serverWait]) {
    final wait = serverWait != null && serverWait > resendCooldown
        ? serverWait
        : resendCooldown;
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
