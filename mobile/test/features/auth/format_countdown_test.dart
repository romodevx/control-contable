import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('whole seconds read as m:ss', () {
    expect(formatCountdown(Duration.zero), '0:00');
    expect(formatCountdown(const Duration(seconds: 45)), '0:45');
    expect(formatCountdown(const Duration(minutes: 1)), '1:00');
    expect(formatCountdown(const Duration(minutes: 12, seconds: 5)), '12:05');
  });

  // M1-16 FU5: a truncated `0:00` showed while the wait was still running;
  // round up, as the contract rounds `Retry-After` up.
  test('a part second rounds up, so a running wait never reads 0:00', () {
    expect(formatCountdown(const Duration(milliseconds: 1)), '0:01');
    expect(formatCountdown(const Duration(milliseconds: 999)), '0:01');
    expect(
      formatCountdown(const Duration(seconds: 59, milliseconds: 1)),
      '1:00',
    );
  });
}
