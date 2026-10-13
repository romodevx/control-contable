import 'package:control_contable/features/auth/application/resend_cooldown.dart';
import 'package:flutter_test/flutter_test.dart';

import 'auth_fakes.dart';

void main() {
  late FakeClock clock;
  late ResendCooldown cooldown;

  setUp(() {
    clock = FakeClock();
    cooldown = ResendCooldown(clock.call);
  });
  tearDown(() => cooldown.dispose());

  Future<void> elapse(WidgetTester tester, Duration by) async {
    clock.advance(by);
    await tester.pump(by);
  }

  testWidgets('idle until started', (tester) async {
    expect(cooldown.isRunning, isFalse);
    expect(cooldown.remaining, Duration.zero);
  });

  testWidgets('counts the 60 s cooldown down by the clock', (tester) async {
    var ticks = 0;
    cooldown
      ..addListener(() => ticks++)
      ..start();

    expect(cooldown.remaining, resendCooldown);
    for (var second = 0; second < 15; second++) {
      await elapse(tester, const Duration(seconds: 1));
    }
    expect(cooldown.remaining, const Duration(seconds: 45));
    expect(ticks, greaterThanOrEqualTo(15), reason: 'notifies every second');

    await elapse(tester, const Duration(seconds: 45));
    expect(cooldown.isRunning, isFalse);
    expect(cooldown.remaining, Duration.zero);
  });

  testWidgets('a longer Retry-After wins over the 60 s floor', (tester) async {
    cooldown.start(const Duration(seconds: 120));
    expect(cooldown.remaining, const Duration(seconds: 120));

    cooldown.start(const Duration(seconds: 5));
    expect(cooldown.remaining, resendCooldown, reason: 'never below 60 s');
    await elapse(tester, resendCooldown);
  });

  testWidgets('time spent in the background still counts', (tester) async {
    cooldown.start();

    // The timer may not fire while the app is suspended; the clock moves.
    clock.advance(const Duration(seconds: 50));
    await tester.pump(const Duration(seconds: 1));

    expect(cooldown.remaining, const Duration(seconds: 10));
    await elapse(tester, const Duration(seconds: 10));
  });
}
