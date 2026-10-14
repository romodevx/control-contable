import 'package:control_contable/core/time/countdown.dart';

/// The server's silent resend cooldown (US-A1c): the app waits it out
/// visibly instead of sending a request that does nothing.
const Duration resendCooldown = Duration(seconds: 60);

/// The countdown before another code can be asked for.
class ResendCooldown extends Countdown {
  /// An idle countdown on the given clock.
  new(super.clock);

  /// Starts the wait: [serverWait] (a `Retry-After`) when longer than the
  /// 60 s cooldown, the cooldown otherwise.
  void start([Duration? serverWait]) => run(
    serverWait != null && serverWait > resendCooldown
        ? serverWait
        : resendCooldown,
  );
}
