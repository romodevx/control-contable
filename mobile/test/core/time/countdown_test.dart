import 'package:control_contable/core/time/countdown.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../features/auth/auth_fakes.dart';

void main() {
  late FakeClock clock;
  late Countdown countdown;

  setUp(() {
    clock = FakeClock();
    countdown = Countdown(clock.call);
  });
  tearDown(() => countdown.dispose());

  Future<void> elapse(WidgetTester tester, Duration by) async {
    clock.advance(by);
    await tester.pump(by);
  }

  testWidgets('idle until run', (tester) async {
    expect(countdown.isRunning, isFalse);
    expect(countdown.remaining, Duration.zero);
  });

  testWidgets('counts any wait down by the clock, notifying each second', (
    tester,
  ) async {
    var ticks = 0;
    countdown
      ..addListener(() => ticks++)
      ..run(const Duration(seconds: 3));

    expect(countdown.remaining, const Duration(seconds: 3));
    await elapse(tester, const Duration(seconds: 1));
    await elapse(tester, const Duration(seconds: 1));
    expect(countdown.remaining, const Duration(seconds: 1));
    expect(ticks, greaterThanOrEqualTo(3));

    await elapse(tester, const Duration(seconds: 1));
    expect(countdown.isRunning, isFalse);
  });

  // M1-15 QA regression: on the emulator the login wait stuck at "0:00" and
  // never went away. A timer tick lands a few ms before the wall-clock
  // deadline, so the last tick saw "0 whole seconds left" yet still running,
  // stopped the ticker, and no one was ever told the wait had ended (the
  // resend button would never come back either).
  testWidgets('a tick just before the deadline still ends the wait for '
      'listeners', (tester) async {
    final shown = <bool>[];
    countdown
      ..addListener(() => shown.add(countdown.isRunning))
      ..run(const Duration(seconds: 3));

    for (var tick = 0; tick < 3; tick++) {
      clock.advance(const Duration(milliseconds: 999)); // timer runs early
      await tester.pump(const Duration(seconds: 1));
    }
    clock.advance(const Duration(seconds: 2));
    await tester.pump(const Duration(seconds: 2));

    expect(countdown.isRunning, isFalse);
    expect(shown.last, isFalse, reason: 'listeners must see the end');
  });

  testWidgets('time spent in the background still counts', (tester) async {
    countdown.run(const Duration(minutes: 3));

    await elapse(tester, const Duration(minutes: 5));

    expect(countdown.isRunning, isFalse);
  });
}
