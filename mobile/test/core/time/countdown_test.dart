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

  testWidgets('time spent in the background still counts', (tester) async {
    countdown.run(const Duration(minutes: 3));

    await elapse(tester, const Duration(minutes: 5));

    expect(countdown.isRunning, isFalse);
  });
}
